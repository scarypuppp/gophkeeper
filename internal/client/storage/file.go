package storage

import (
	"fmt"
	"slices"
)

// FilePatch описывает частичное изменение записи о файле: nil-поля остаются прежними.
// Size и Checksum обновляют вместе, когда содержимое файла в каталоге downloads
// изменилось; FileHash присваивает сервер при синхронизации.
type FilePatch struct {
	Metadata *string
	Size     *int64
	Checksum *string
	FileHash *string
}

// Files возвращает записи о файлах, упорядоченные по имени.
func (s *LocalStorage) Files() []File {
	return sortedByName(s.files)
}

// GetFile возвращает запись о файле по имени.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) GetFile(fileName string) (File, error) {
	i := indexOf(s.files, fileName)
	if i < 0 {
		return File{}, fmt.Errorf("file %q: %w", fileName, ErrNotFound)
	}
	return s.files[i], nil
}

// AddFile добавляет запись о файле. Размер и контрольную сумму содержимого
// считает вызывающий: содержимое файла лежит вне хранилища, в каталоге downloads.
// Возвращает ErrAlreadyExists, если запись с таким именем уже есть.
func (s *LocalStorage) AddFile(file File) (File, error) {
	if file.FileName == "" {
		return File{}, ErrEmptyName
	}
	if indexOf(s.files, file.FileName) >= 0 {
		return File{}, fmt.Errorf("file %q: %w", file.FileName, ErrAlreadyExists)
	}

	file.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Files = append(data.Files, file)
	if err := s.commit(data); err != nil {
		return File{}, err
	}
	return file, nil
}

// UpdateFile меняет у записи о файле только переданные в patch поля.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) UpdateFile(fileName string, patch FilePatch) (File, error) {
	i := indexOf(s.files, fileName)
	if i < 0 {
		return File{}, fmt.Errorf("file %q: %w", fileName, ErrNotFound)
	}

	file := s.files[i]
	if patch.Metadata != nil {
		file.Metadata = *patch.Metadata
	}
	if patch.Size != nil {
		file.Size = *patch.Size
	}
	if patch.Checksum != nil {
		file.Checksum = *patch.Checksum
	}
	if patch.FileHash != nil {
		file.FileHash = *patch.FileHash
	}
	file.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Files[i] = file
	if err := s.commit(data); err != nil {
		return File{}, err
	}
	return file, nil
}

// DeleteFile удаляет запись о файле по имени. Само содержимое из каталога
// downloads удаляет вызывающий.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) DeleteFile(fileName string) error {
	i := indexOf(s.files, fileName)
	if i < 0 {
		return fmt.Errorf("file %q: %w", fileName, ErrNotFound)
	}

	data := s.Snapshot()
	data.Files = slices.Delete(data.Files, i, i+1)
	return s.commit(data)
}
