package sync

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// fileChecksums собирает контрольные суммы локальных записей о файлах.
func fileChecksums(items []storage.File) map[string]string {
	checksums := make(map[string]string, len(items))
	for _, item := range items {
		checksums[item.FileName] = contentChecksum(item.Checksum, item.Metadata)
	}
	return checksums
}

// remoteFileChecksums собирает контрольные суммы файлов на сервере.
func remoteFileChecksums(items map[string]api.FileItemResponse) map[string]string {
	return remoteChecksums(items, func(item api.FileItemResponse) string {
		return contentChecksum(item.Checksum, item.Metadata)
	})
}

// fileName возвращает имя файла — его идентификатор.
func fileName(item storage.File) string {
	return item.FileName
}

// filePath возвращает путь к локальной копии файла.
func (s *Syncer) filePath(name string) string {
	return filepath.Join(s.downloads, name)
}

// fileConflict собирает описание конфликта по файлу. Содержимое файлов
// сравнить нельзя, поэтому показываются имя, размер, контрольная сумма
// и время изменения обеих версий.
func (s *Syncer) fileConflict(st step, remote remoteData) (Conflict, error) {
	conflict := Conflict{Kind: kindFile, Name: st.name}

	if st.local.Present {
		local, err := s.storage.GetFile(st.name)
		if err != nil {
			return conflict, err
		}
		conflict.Local = fileFields(local.FileName, local.Size, local.Checksum, local.UpdatedAt)
	}
	if st.remote.Present {
		item := remote.files[st.name]
		conflict.Remote = fileFields(item.FileName, item.Size, item.Checksum, item.UpdatedAt)
	}
	return conflict, nil
}

// fileFields перечисляет поля файла для показа в конфликте.
func fileFields(name string, size int64, checksum string, updatedAt time.Time) []Field {
	return []Field{
		{Name: "name", Value: name},
		{Name: "size", Value: strconv.FormatInt(size, 10)},
		{Name: "checksum", Value: checksum},
		{Name: "updated_at", Value: updatedAt.Format(time.RFC3339)},
	}
}

// applyFile приводит файл к выбранной версии. Содержимое при синхронизации
// не скачивается: запись появляется локально без содержимого, а забрать его
// можно командой 'gophkeeper file download'.
func (s *Syncer) applyFile(st step, remote remoteData, result *storage.Data) error {
	switch st.action {
	case ActionPush:
		return s.pushFile(st, remote, result)

	case ActionPull:
		item := remote.files[st.name]
		if err := s.warnOutdatedCopy(item); err != nil {
			return err
		}
		result.Files = upsert(result.Files, fileName, storage.File{
			FileName:  item.FileName,
			FileHash:  item.FileHash,
			Size:      item.Size,
			Metadata:  item.Metadata,
			Checksum:  item.Checksum,
			UpdatedAt: item.UpdatedAt,
		})
		s.report(kindFile, st.name, createdOrUpdated(st.local.Present))

	case ActionDeleteRemote:
		item := remote.files[st.name]
		if err := s.client.DeleteFile(s.token, item.FileHash, item.FileName); err != nil {
			return err
		}
		s.report(kindFile, st.name, "deleted")

	case ActionDeleteLocal:
		result.Files = remove(result.Files, fileName, st.name)
		if err := os.Remove(s.filePath(st.name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove local copy: %w", err)
		}
		s.report(kindFile, st.name, "deleted")
	}
	return nil
}

// pushFile загружает файл на сервер: новый — целиком, существующий —
// содержимым и метаданными по отдельности, чтобы не гонять байты зря.
func (s *Syncer) pushFile(st step, remote remoteData, result *storage.Data) error {
	local, err := s.storage.GetFile(st.name)
	if err != nil {
		return err
	}

	if !st.remote.Present {
		if err := s.requireContent(local.FileName); err != nil {
			return err
		}
		uploaded, err := s.client.UploadFile(s.token, s.filePath(local.FileName), local.Metadata)
		if err != nil {
			return err
		}
		// Хэш файла назначает сервер, локальная запись его ещё не знает.
		local.FileHash = uploaded.FileHash
		result.Files = upsert(result.Files, fileName, local)
		s.report(kindFile, st.name, "created")
		return nil
	}

	item := remote.files[st.name]
	if local.Checksum != item.Checksum {
		if err := s.requireContent(local.FileName); err != nil {
			return err
		}
		if _, err := s.client.UpdateFileContent(s.token, item.FileHash, item.FileName, s.filePath(local.FileName)); err != nil {
			return err
		}
	}
	if local.Metadata != item.Metadata {
		if _, err := s.client.UpdateFileMetadata(s.token, item.FileHash, item.FileName, local.Metadata); err != nil {
			return err
		}
	}
	s.report(kindFile, st.name, "updated")
	return nil
}

// requireContent проверяет, что содержимое файла есть на этой машине:
// без него отправить файл на сервер нечем.
func (s *Syncer) requireContent(name string) error {
	_, err := os.Stat(s.filePath(name))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("content is not on this machine, run 'gophkeeper file download %s' first", name)
	}
	return err
}

// warnOutdatedCopy предупреждает, что локальная копия файла устарела:
// содержимое при синхронизации не скачивается.
func (s *Syncer) warnOutdatedCopy(item api.FileItemResponse) error {
	path := s.filePath(item.FileName)
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size() == item.Size {
		return nil
	}
	s.warn("local copy of '%s' is outdated, run 'gophkeeper file download %s'", item.FileName, item.FileName)
	return nil
}
