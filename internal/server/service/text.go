package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
)

// ErrTextNotExists возвращается, если текст с указанным именем не найден у пользователя.
var ErrTextNotExists = errors.New("text not exists")

// ErrTextAlreadyExists возвращается при попытке создать текст с уже занятым у пользователя именем.
var ErrTextAlreadyExists = errors.New("text already exists")

// TextsService реализует бизнес-логику работы с произвольными текстовыми данными пользователя.
// Идентификатором текста в рамках пользователя является его имя.
type TextsService struct {
	uow repository.UnitOfWork
}

func NewTextsService(uow repository.UnitOfWork) *TextsService {
	return &TextsService{uow: uow}
}

// GetText возвращает текст пользователя по имени.
// Возвращает ErrTextNotExists, если текста с таким именем нет.
func (s *TextsService) GetText(ctx context.Context, userID int64, name string) (*entities.Text, error) {
	text, err := s.uow.Texts().GetByName(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrTextNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetText: %w", err)
	}
	return text, nil
}

// GetTexts возвращает все тексты пользователя.
func (s *TextsService) GetTexts(ctx context.Context, userID int64) ([]entities.Text, error) {
	texts, err := s.uow.Texts().GetTextsByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetTexts: %w", err)
	}
	return texts, nil
}

// CreateText создаёт текст пользователя.
// Возвращает ErrTextAlreadyExists, если у пользователя уже есть текст с таким именем.
func (s *TextsService) CreateText(
	ctx context.Context,
	userID int64,
	name string,
	text string,
	metadata string,
) (*entities.Text, error) {
	_, err := s.uow.Texts().GetByName(ctx, userID, name)
	if err == nil {
		return nil, ErrTextAlreadyExists
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, fmt.Errorf("CreateText: %w", err)
	}

	created, err := s.uow.Texts().CreateText(ctx, entities.Text{
		Owner:    userID,
		Name:     name,
		Text:     text,
		Metadata: metadata,
		Checksum: checksum.Text(text, metadata),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateText: %w", err)
	}
	return created, nil
}

// UpdateText обновляет содержимое текста пользователя, найденного по имени.
// Возвращает ErrTextNotExists, если текста с таким именем нет.
func (s *TextsService) UpdateText(
	ctx context.Context,
	userID int64,
	name string,
	text string,
	metadata string,
) (*entities.Text, error) {
	current, err := s.GetText(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	updated := entities.Text{
		Id:        current.Id,
		Owner:     userID,
		Name:      name,
		Text:      text,
		Metadata:  metadata,
		Checksum:  checksum.Text(text, metadata),
		CreatedAt: current.CreatedAt,
	}
	updatedAt, err := s.uow.Texts().UpdateText(ctx, updated)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrTextNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateText: %w", err)
	}
	updated.UpdatedAt = updatedAt
	return &updated, nil
}

// DeleteText удаляет текст пользователя по имени.
// Возвращает ErrTextNotExists, если текста с таким именем нет.
func (s *TextsService) DeleteText(ctx context.Context, userID int64, name string) error {
	err := s.uow.Texts().DeleteText(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return ErrTextNotExists
	}
	if err != nil {
		return fmt.Errorf("DeleteText: %w", err)
	}
	return nil
}
