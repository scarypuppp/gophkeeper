package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// Права файла хранилища: в нём лежат пароли, поэтому доступ только у владельца.
const storageFileMode fs.FileMode = 0600

var (
	// ErrNotFound возвращается, если записи с указанным именем нет в хранилище.
	ErrNotFound = errors.New("record not found")
	// ErrAlreadyExists возвращается при попытке создать запись с занятым именем.
	ErrAlreadyExists = errors.New("record already exists")
	// ErrEmptyName возвращается, если имя записи не указано.
	ErrEmptyName = errors.New("record name is required")
	// ErrNotLoaded возвращается при попытке изменить хранилище до вызова Load:
	// иначе пустое хранилище затёрло бы данные, лежащие в файле.
	ErrNotLoaded = errors.New("storage is not loaded")
)

// LocalStorage — данные пользователя, хранящиеся в одном JSON-файле.
// Записи держатся в памяти; каждое изменение сразу сохраняется на диск,
// поэтому вызывающему не нужно помнить про Save.
type LocalStorage struct {
	secrets []Secret
	files   []File
	texts   []Text
	cards   []Card

	path string
	// loaded показывает, что содержимое файла уже прочитано и его можно менять.
	loaded bool
	// now возвращает время изменения записи; подменяется в тестах.
	now func() time.Time
}

// NewLocalStorage создаёт хранилище над файлом path. Диск при этом не читается:
// содержимое загружается вызовом Load.
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{
		path: path,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// Path возвращает путь к файлу хранилища.
func (s *LocalStorage) Path() string {
	return s.path
}

// Exists сообщает, существует ли файл хранилища.
// Синхронизация по этому признаку отличает отсутствующий BASE от пустого.
func (s *LocalStorage) Exists() (bool, error) {
	_, err := os.Stat(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", s.path, err)
	}
	return true, nil
}

// Load читает данные из файла. Отсутствие файла ошибкой не считается:
// хранилище остаётся пустым.
func (s *LocalStorage) Load() error {
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		s.setData(Data{})
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", s.path, err)
	}

	var data Data
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parse %s: %w", s.path, err)
		}
	}
	s.setData(data)
	return nil
}

// Save записывает данные на диск. Файл заменяется целиком и атомарно,
// поэтому прерванная запись не портит уже сохранённые данные.
func (s *LocalStorage) Save() error {
	if !s.loaded {
		return ErrNotLoaded
	}

	raw, err := json.MarshalIndent(s.Snapshot(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", s.path, err)
	}
	raw = append(raw, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(s.path)+".tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	defer os.Remove(tmp.Name())

	if err := writeAndClose(tmp, raw); err != nil {
		return fmt.Errorf("write %s: %w", tmp.Name(), err)
	}
	if err := os.Chmod(tmp.Name(), storageFileMode); err != nil {
		return fmt.Errorf("chmod %s: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("replace %s: %w", s.path, err)
	}
	return nil
}

// Clear очищает хранилище и записывает пустой файл: при логауте данные
// прошлого пользователя не должны остаться на машине.
// Содержимое файла при этом не читается — оно всё равно стирается.
func (s *LocalStorage) Clear() error {
	s.setData(Data{})
	return s.Save()
}

// Snapshot возвращает копию данных хранилища; изменения копии на хранилище
// не влияют.
func (s *LocalStorage) Snapshot() Data {
	return Data{
		Files:   slices.Clone(s.files),
		Texts:   slices.Clone(s.texts),
		Cards:   slices.Clone(s.cards),
		Secrets: slices.Clone(s.secrets),
	}
}

// Replace заменяет содержимое хранилища и сохраняет его.
// Значения записей берутся как есть: контрольные суммы и время изменения
// не пересчитываются, поэтому синхронизация может класть сюда данные сервера.
// Файл перезаписывается целиком, поэтому предварительный Load не нужен.
func (s *LocalStorage) Replace(data Data) error {
	s.setData(data)
	return s.Save()
}

// commit заменяет данные хранилища и сохраняет их на диск.
// Если запись не удалась, хранилище остаётся в прежнем состоянии,
// чтобы память не разошлась с содержимым файла.
func (s *LocalStorage) commit(data Data) error {
	if !s.loaded {
		return ErrNotLoaded
	}

	previous := s.Snapshot()
	s.setData(data)
	if err := s.Save(); err != nil {
		s.setData(previous)
		return err
	}
	return nil
}

// setData помещает данные в хранилище, отсеивая nil-срезы.
func (s *LocalStorage) setData(data Data) {
	s.loaded = true
	s.files = emptyIfNil(data.Files)
	s.texts = emptyIfNil(data.Texts)
	s.cards = emptyIfNil(data.Cards)
	s.secrets = emptyIfNil(data.Secrets)
}

// emptyIfNil заменяет nil-срез пустым, чтобы в JSON вместо null был [].
func emptyIfNil[T record](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

// writeAndClose пишет данные в файл и закрывает его, сбрасывая буферы ОС.
func writeAndClose(f *os.File, raw []byte) error {
	if _, err := f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
