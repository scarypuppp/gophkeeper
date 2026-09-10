package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophkeeper/internal/entities"
)

// FileRepositoryPostgres реализует FileRepository поверх PostgreSQL.
type FileRepositoryPostgres struct {
	exec sqlx.ExtContext
}

// NewFileRepositoryPostgres создаёт FileRepositoryPostgres, работающий без транзакции.
func NewFileRepositoryPostgres(db *sqlx.DB) *FileRepositoryPostgres {
	return &FileRepositoryPostgres{exec: db}
}

// NewFileRepositoryPostgresTx создаёт FileRepositoryPostgres, работающий внутри транзакции.
func NewFileRepositoryPostgresTx(tx *sqlx.Tx) *FileRepositoryPostgres {
	return &FileRepositoryPostgres{exec: tx}
}

func (r *FileRepositoryPostgres) GetByName(ctx context.Context, owner int64, fileName string) (*entities.File, error) {
	var file entities.File
	err := sqlx.GetContext(ctx, r.exec, &file,
		`SELECT id, owner, file_name, file_hash, storage_file_path, metadata, checksum, size, created_at, updated_at
			   FROM files
			   WHERE owner = $1 AND file_name = $2`,
		owner, fileName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	return &file, nil
}

func (r *FileRepositoryPostgres) GetByHashAndName(ctx context.Context, owner int64, fileHash, fileName string) (*entities.File, error) {
	var file entities.File
	err := sqlx.GetContext(ctx, r.exec, &file,
		`SELECT id, owner, file_name, file_hash, storage_file_path, metadata, checksum, size, created_at, updated_at
			   FROM files
			   WHERE owner = $1 AND file_hash = $2 AND file_name = $3`,
		owner, fileHash, fileName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByHashAndName: %w", err)
	}
	return &file, nil
}

func (r *FileRepositoryPostgres) GetFilesByOwner(ctx context.Context, owner int64) ([]entities.File, error) {
	var files []entities.File
	err := sqlx.SelectContext(ctx, r.exec, &files,
		`SELECT id, owner, file_name, file_hash, storage_file_path, metadata, checksum, size, created_at, updated_at
			   FROM files WHERE owner = $1`,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("GetFilesByOwner: %w", err)
	}
	return files, nil
}

func (r *FileRepositoryPostgres) CreateFile(ctx context.Context, file entities.File) (*entities.File, error) {
	const insertQuery = `
    INSERT INTO files (owner, file_name, file_hash, storage_file_path, metadata, checksum, size)
    VALUES (:owner, :file_name, :file_hash, :storage_file_path, :metadata, :checksum, :size)
    RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, insertQuery, file)
	if err != nil {
		return nil, fmt.Errorf("CreateFile: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("CreateFile: rows iteration: %w", err)
		}
		return nil, fmt.Errorf("CreateFile: no id returned after insert")
	}
	if err := rows.Scan(&file.Id, &file.CreatedAt, &file.UpdatedAt); err != nil {
		return nil, fmt.Errorf("CreateFile: scan returned row: %w", err)
	}
	return &file, nil
}

// UpdateFileMetadata обновляет метаданные файла. Содержимое файла не меняется.
func (r *FileRepositoryPostgres) UpdateFileMetadata(ctx context.Context, owner int64, fileHash, fileName, metadata string) (time.Time, error) {
	var updatedAt time.Time
	err := sqlx.GetContext(ctx, r.exec, &updatedAt,
		`UPDATE files SET metadata = $1, updated_at = now()
         WHERE owner = $2 AND file_hash = $3 AND file_name = $4
         RETURNING updated_at`,
		metadata, owner, fileHash, fileName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNoRows
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("UpdateFileMetadata: %w", err)
	}
	return updatedAt, nil
}

// UpdateFileContent обновляет данные файла после перезаписи его содержимого:
// контрольную сумму, размер и время загрузки. Метаданные не затрагиваются.
func (r *FileRepositoryPostgres) UpdateFileContent(
	ctx context.Context,
	owner int64,
	fileHash, fileName, checksum string,
	size int64,
) (time.Time, error) {
	var updatedAt time.Time
	err := sqlx.GetContext(ctx, r.exec, &updatedAt,
		`UPDATE files SET checksum = $1, size = $2, updated_at = now()
         WHERE owner = $3 AND file_hash = $4 AND file_name = $5
         RETURNING updated_at`,
		checksum, size, owner, fileHash, fileName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNoRows
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("UpdateFileContent: %w", err)
	}
	return updatedAt, nil
}

func (r *FileRepositoryPostgres) DeleteFile(ctx context.Context, owner int64, fileHash, fileName string) error {
	result, err := r.exec.ExecContext(ctx,
		`DELETE FROM files WHERE owner = $1 AND file_hash = $2 AND file_name = $3`, owner, fileHash, fileName,
	)
	if err != nil {
		return fmt.Errorf("DeleteFile: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteFile: %w", err)
	}
	if affected == 0 {
		return ErrNoRows
	}
	return nil
}
