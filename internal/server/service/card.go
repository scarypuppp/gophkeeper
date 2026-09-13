package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
)

// ErrCardNotExists возвращается, если карта с указанным именем не найдена у пользователя.
var ErrCardNotExists = errors.New("card not exists")

// ErrCardAlreadyExists возвращается при попытке создать карту с уже занятым у пользователя именем.
var ErrCardAlreadyExists = errors.New("card already exists")

// CardsService реализует бизнес-логику работы с банковскими картами пользователя.
// Идентификатором карты в рамках пользователя является её имя.
type CardsService struct {
	uow repository.UnitOfWork
}

func NewCardsService(uow repository.UnitOfWork) *CardsService {
	return &CardsService{uow: uow}
}

// GetCard возвращает карту пользователя по имени.
// Возвращает ErrCardNotExists, если карты с таким именем нет.
func (s *CardsService) GetCard(ctx context.Context, userID int64, name string) (*entities.Card, error) {
	card, err := s.uow.Cards().GetByName(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrCardNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetCard: %w", err)
	}
	return card, nil
}

// GetCards возвращает все карты пользователя.
func (s *CardsService) GetCards(ctx context.Context, userID int64) ([]entities.Card, error) {
	cards, err := s.uow.Cards().GetCardsByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetCards: %w", err)
	}
	return cards, nil
}

// CreateCard создаёт карту пользователя.
// Возвращает ErrCardAlreadyExists, если у пользователя уже есть карта с таким именем.
func (s *CardsService) CreateCard(
	ctx context.Context,
	userID int64,
	name string,
	number string,
	holder string,
	expiresAt string,
	cvv string,
	metadata string,
) (*entities.Card, error) {
	created, err := s.uow.Cards().CreateCard(ctx, entities.Card{
		Owner:     userID,
		Name:      name,
		Number:    number,
		Holder:    holder,
		ExpiresAt: expiresAt,
		CVV:       cvv,
		Metadata:  metadata,
		Checksum:  checksum.Card(number, holder, expiresAt, cvv, metadata),
	})
	if errors.Is(err, repository.ErrUniqueViolation) {
		return nil, ErrCardAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("CreateCard: %w", err)
	}
	return created, nil
}

// UpdateCard обновляет содержимое карты пользователя, найденной по имени.
// Возвращает ErrCardNotExists, если карты с таким именем нет.
func (s *CardsService) UpdateCard(
	ctx context.Context,
	userID int64,
	name string,
	number string,
	holder string,
	expiresAt string,
	cvv string,
	metadata string,
) (*entities.Card, error) {
	current, err := s.GetCard(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	card := entities.Card{
		Id:        current.Id,
		Owner:     userID,
		Name:      name,
		Number:    number,
		Holder:    holder,
		ExpiresAt: expiresAt,
		CVV:       cvv,
		Metadata:  metadata,
		Checksum:  checksum.Card(number, holder, expiresAt, cvv, metadata),
		CreatedAt: current.CreatedAt,
	}
	updatedAt, err := s.uow.Cards().UpdateCard(ctx, card)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrCardNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateCard: %w", err)
	}
	card.UpdatedAt = updatedAt
	return &card, nil
}

// DeleteCard удаляет карту пользователя по имени.
// Возвращает ErrCardNotExists, если карты с таким именем нет.
func (s *CardsService) DeleteCard(ctx context.Context, userID int64, name string) error {
	err := s.uow.Cards().DeleteCard(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return ErrCardNotExists
	}
	if err != nil {
		return fmt.Errorf("DeleteCard: %w", err)
	}
	return nil
}
