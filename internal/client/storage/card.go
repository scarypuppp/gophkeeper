package storage

import (
	"fmt"
	"slices"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
)

// CardPatch описывает частичное изменение карты: nil-поля остаются прежними.
type CardPatch struct {
	Number    *string
	Holder    *string
	ExpiresAt *string
	CVV       *string
	Metadata  *string
}

// Cards возвращает карты, упорядоченные по имени.
func (s *LocalStorage) Cards() []Card {
	return sortedByName(s.cards)
}

// GetCard возвращает карту по имени.
// Возвращает ErrNotFound, если карты с таким именем нет.
func (s *LocalStorage) GetCard(name string) (Card, error) {
	i := indexOf(s.cards, name)
	if i < 0 {
		return Card{}, fmt.Errorf("card %q: %w", name, ErrNotFound)
	}
	return s.cards[i], nil
}

// AddCard добавляет карту; контрольную сумму и время изменения
// проставляет хранилище.
// Возвращает ErrAlreadyExists, если карта с таким именем уже есть.
func (s *LocalStorage) AddCard(card Card) (Card, error) {
	if card.Name == "" {
		return Card{}, ErrEmptyName
	}
	if indexOf(s.cards, card.Name) >= 0 {
		return Card{}, fmt.Errorf("card %q: %w", card.Name, ErrAlreadyExists)
	}

	card.Checksum = checksum.Card(card.Number, card.Holder, card.ExpiresAt, card.CVV, card.Metadata)
	card.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Cards = append(data.Cards, card)
	if err := s.commit(data); err != nil {
		return Card{}, err
	}
	return card, nil
}

// UpdateCard меняет у карты только переданные в patch поля.
// Возвращает ErrNotFound, если карты с таким именем нет.
func (s *LocalStorage) UpdateCard(name string, patch CardPatch) (Card, error) {
	i := indexOf(s.cards, name)
	if i < 0 {
		return Card{}, fmt.Errorf("card %q: %w", name, ErrNotFound)
	}

	card := s.cards[i]
	if patch.Number != nil {
		card.Number = *patch.Number
	}
	if patch.Holder != nil {
		card.Holder = *patch.Holder
	}
	if patch.ExpiresAt != nil {
		card.ExpiresAt = *patch.ExpiresAt
	}
	if patch.CVV != nil {
		card.CVV = *patch.CVV
	}
	if patch.Metadata != nil {
		card.Metadata = *patch.Metadata
	}
	card.Checksum = checksum.Card(card.Number, card.Holder, card.ExpiresAt, card.CVV, card.Metadata)
	card.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Cards[i] = card
	if err := s.commit(data); err != nil {
		return Card{}, err
	}
	return card, nil
}

// DeleteCard удаляет карту по имени.
// Возвращает ErrNotFound, если карты с таким именем нет.
func (s *LocalStorage) DeleteCard(name string) error {
	i := indexOf(s.cards, name)
	if i < 0 {
		return fmt.Errorf("card %q: %w", name, ErrNotFound)
	}

	data := s.Snapshot()
	data.Cards = slices.Delete(data.Cards, i, i+1)
	return s.commit(data)
}
