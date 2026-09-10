package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/filestorage"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
)

// fileHashSize задаёт длину случайного хэша файла в байтах (до hex-кодирования).
const fileHashSize = 16

// ErrFileNotExists возвращается, если файл с указанным хэшем не найден у пользователя.
var ErrFileNotExists = errors.New("file not exists")

// ErrFileAlreadyExists возвращается при попытке загрузить файл с уже существующим у пользователя хэшем.
var ErrFileAlreadyExists = errors.New("file already exists")

// FileService реализует бизнес-логику загрузки, получения и удаления файлов пользователя.
type FileService struct {
	uow     repository.UnitOfWork
	storage *filestorage.S3Storage
}

// NewFileService создаёт новый FileService с переданными UnitOfWork и S3-хранилищем.
func NewFileService(uow repository.UnitOfWork, storage *filestorage.S3Storage) *FileService {
	return &FileService{uow: uow, storage: storage}
}

// GetFileByName возвращает метаданные файла пользователя по хэшу.
// Возвращает ErrFileNotExists, если файл с таким именем у пользователя отсутствует.
func (fs *FileService) GetFileByName(ctx context.Context, userID int64, fileHash string) (*entities.File, error) {
	file, err := fs.uow.Files().GetByName(ctx, userID, fileHash)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrFileNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetFileByName: %w", err)
	}
	return file, nil
}

// GetFile возвращает метаданные файла пользователя по хэшу и имени файла.
// Возвращает ErrFileNotExists, если такой файл у пользователя отсутствует.
func (fs *FileService) GetFile(ctx context.Context, userID int64, fileHash string, fileName string) (*entities.File, error) {
	file, err := fs.uow.Files().GetByHashAndName(ctx, userID, fileHash, fileName)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrFileNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetFile: %w", err)
	}
	return file, nil
}

// GetFileContent возвращает содержимое файла из S3-хранилища.
func (fs *FileService) GetFileContent(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	return fs.storage.GetFile(ctx, storagePath)
}

// GetFiles возвращает файлы пользователя.
func (fs *FileService) GetFiles(ctx context.Context, userID int64) ([]entities.File, error) {
	files, err := fs.uow.Files().GetFilesByOwner(ctx, userID)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrFileNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetFiles: %w", err)
	}
	return files, err
}

// CreateFile загружает содержимое файла в S3-хранилище и сохраняет его метаданные.
// Возвращает ErrFileAlreadyExists, если у пользователя уже есть файл с таким именем.
func (fs *FileService) CreateFile(
	ctx context.Context,
	userID int64,
	fileName string,
	metadata string,
	size int64,
	file io.Reader,
) (*entities.File, error) {
	// Валидация
	_, err := fs.uow.Files().GetByName(ctx, userID, fileName)
	if err == nil {
		return nil, ErrFileAlreadyExists
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, fmt.Errorf("CreateFile: %w", err)
	}

	// Сохранение файла в хранилище
	fileHash, err := generateFileHash()
	if err != nil {
		return nil, fmt.Errorf("CreateFile: %w", err)
	}
	storageFileName := strings.ReplaceAll(fileName, " ", "_")
	storagePath := fmt.Sprintf("%d/%s", userID, storageFileName)
	checksum, err := fs.storage.PutFile(ctx, storagePath, file)
	if err != nil {
		return nil, fmt.Errorf("CreateFile: upload: %w", err)
	}

	// Сохранение информации о файле в БД
	created, err := fs.uow.Files().CreateFile(ctx, entities.File{
		Owner:           userID,
		FileName:        fileName,
		FileHash:        fileHash,
		StorageFilePath: storagePath,
		Metadata:        metadata,
		Checksum:        checksum,
		Size:            size,
	})
	if err != nil {
		_ = fs.storage.DeleteFile(ctx, storagePath)
		return nil, fmt.Errorf("CreateFile: %w", err)
	}
	return created, nil
}

// UpdateFileMetadata обновляет метаданные файла пользователя, найденного по хэшу и имени.
// Содержимое файла в хранилище не меняется.
// Возвращает ErrFileNotExists, если такого файла у пользователя нет.
func (fs *FileService) UpdateFileMetadata(
	ctx context.Context,
	userID int64,
	fileHash string,
	fileName string,
	metadata string,
) (*entities.File, error) {
	file, err := fs.GetFile(ctx, userID, fileHash, fileName)
	if err != nil {
		return nil, err
	}
	updatedAt, err := fs.uow.Files().UpdateFileMetadata(ctx, userID, fileHash, fileName, metadata)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, ErrFileNotExists
		}
		return nil, fmt.Errorf("UpdateFileMetadata: %w", err)
	}
	file.Metadata = metadata
	file.UpdatedAt = updatedAt
	return file, nil
}

// UpdateFileContent перезаписывает содержимое файла пользователя, найденного по хэшу и имени.
// Хэш, имя, путь в хранилище и метаданные остаются прежними, обновляются
// контрольная сумма, размер и время загрузки.
// Возвращает ErrFileNotExists, если такого файла у пользователя нет.
func (fs *FileService) UpdateFileContent(
	ctx context.Context,
	userID int64,
	fileHash string,
	fileName string,
	size int64,
	content io.Reader,
) (*entities.File, error) {
	file, err := fs.GetFile(ctx, userID, fileHash, fileName)
	if err != nil {
		return nil, err
	}

	checksum, err := fs.storage.PutFile(ctx, file.StorageFilePath, content)
	if err != nil {
		return nil, fmt.Errorf("UpdateFileContent: upload: %w", err)
	}

	updatedAt, err := fs.uow.Files().UpdateFileContent(ctx, userID, fileHash, fileName, checksum, size)
	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, ErrFileNotExists
		}
		return nil, fmt.Errorf("UpdateFileContent: %w", err)
	}

	file.Checksum = checksum
	file.Size = size
	file.UpdatedAt = updatedAt
	return file, nil
}

// DeleteFile удаляет файл пользователя из хранилища и его метаданные.
// Возвращает ErrFileNotExists, если файла с таким хэшем и именем у пользователя нет.
func (fs *FileService) DeleteFile(ctx context.Context, userID int64, fileHash string, fileName string) error {
	file, err := fs.GetFile(ctx, userID, fileHash, fileName)
	if err != nil {
		return err
	}
	if err := fs.storage.DeleteFile(ctx, file.StorageFilePath); err != nil {
		return fmt.Errorf("DeleteFile: %w", err)
	}
	if err := fs.uow.Files().DeleteFile(ctx, userID, fileHash, fileName); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return ErrFileNotExists
		}
		return fmt.Errorf("DeleteFile: %w", err)
	}
	return nil
}

// generateFileHash генерирует случайный hex-хэш, идентифицирующий файл.
func generateFileHash() (string, error) {
	buf := make([]byte, fileHashSize)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generateFileHash: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
