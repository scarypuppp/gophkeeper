package storage

import (
	"fmt"
	"slices"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
)

// TextPatch описывает частичное изменение текста: nil-поля остаются прежними.
type TextPatch struct {
	Text     *string
	Metadata *string
}

// Texts возвращает текстовые записи, упорядоченные по имени.
func (s *LocalStorage) Texts() []Text {
	return sortedByName(s.texts)
}

// GetText возвращает текстовую запись по имени.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) GetText(name string) (Text, error) {
	i := indexOf(s.texts, name)
	if i < 0 {
		return Text{}, fmt.Errorf("text %q: %w", name, ErrNotFound)
	}
	return s.texts[i], nil
}

// AddText добавляет текстовую запись; контрольную сумму и время изменения
// проставляет хранилище.
// Возвращает ErrAlreadyExists, если запись с таким именем уже есть.
func (s *LocalStorage) AddText(text Text) (Text, error) {
	if text.Name == "" {
		return Text{}, ErrEmptyName
	}
	if indexOf(s.texts, text.Name) >= 0 {
		return Text{}, fmt.Errorf("text %q: %w", text.Name, ErrAlreadyExists)
	}

	text.Checksum = checksum.Text(text.Text, text.Metadata)
	text.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Texts = append(data.Texts, text)
	if err := s.commit(data); err != nil {
		return Text{}, err
	}
	return text, nil
}

// UpdateText меняет у текстовой записи только переданные в patch поля.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) UpdateText(name string, patch TextPatch) (Text, error) {
	i := indexOf(s.texts, name)
	if i < 0 {
		return Text{}, fmt.Errorf("text %q: %w", name, ErrNotFound)
	}

	text := s.texts[i]
	if patch.Text != nil {
		text.Text = *patch.Text
	}
	if patch.Metadata != nil {
		text.Metadata = *patch.Metadata
	}
	text.Checksum = checksum.Text(text.Text, text.Metadata)
	text.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Texts[i] = text
	if err := s.commit(data); err != nil {
		return Text{}, err
	}
	return text, nil
}

// DeleteText удаляет текстовую запись по имени.
// Возвращает ErrNotFound, если записи с таким именем нет.
func (s *LocalStorage) DeleteText(name string) error {
	i := indexOf(s.texts, name)
	if i < 0 {
		return fmt.Errorf("text %q: %w", name, ErrNotFound)
	}

	data := s.Snapshot()
	data.Texts = slices.Delete(data.Texts, i, i+1)
	return s.commit(data)
}
