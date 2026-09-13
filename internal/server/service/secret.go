package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
)

// ErrSecretNotExists возвращается, если секрет с указанным именем не найден у пользователя.
var ErrSecretNotExists = errors.New("secret not exists")

// ErrSecretAlreadyExists возвращается при попытке создать секрет с уже занятым у пользователя именем.
var ErrSecretAlreadyExists = errors.New("secret already exists")

// SecretsService реализует бизнес-логику работы с секретами пользователя.
// Идентификатором секрета в рамках пользователя является его имя.
type SecretsService struct {
	uow repository.UnitOfWork
}

func NewSecretsService(uow repository.UnitOfWork) *SecretsService {
	return &SecretsService{uow: uow}
}

// GetSecret возвращает секрет пользователя по имени.
// Возвращает ErrSecretNotExists, если секрета с таким именем нет.
func (s *SecretsService) GetSecret(ctx context.Context, userID int64, name string) (*entities.Secret, error) {
	secret, err := s.uow.Secrets().GetByName(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrSecretNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("GetSecret: %w", err)
	}
	return secret, nil
}

// GetSecrets возвращает все секреты пользователя.
func (s *SecretsService) GetSecrets(ctx context.Context, userID int64) ([]entities.Secret, error) {
	secrets, err := s.uow.Secrets().GetSecretsByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetSecrets: %w", err)
	}
	return secrets, nil
}

// CreateSecret создаёт секрет пользователя.
// Возвращает ErrSecretAlreadyExists, если у пользователя уже есть секрет с таким именем.
func (s *SecretsService) CreateSecret(
	ctx context.Context,
	userID int64,
	name string,
	login string,
	password string,
	metadata string,
) (*entities.Secret, error) {
	created, err := s.uow.Secrets().CreateSecret(ctx, entities.Secret{
		Owner:    userID,
		Name:     name,
		Login:    login,
		Password: password,
		Metadata: metadata,
		Checksum: checksum.Secret(login, password, metadata),
	})
	if errors.Is(err, repository.ErrUniqueViolation) {
		return nil, ErrSecretAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("CreateSecret: %w", err)
	}
	return created, nil
}

// UpdateSecret обновляет содержимое секрета пользователя, найденного по имени.
// Возвращает ErrSecretNotExists, если секрета с таким именем нет.
func (s *SecretsService) UpdateSecret(
	ctx context.Context,
	userID int64,
	name string,
	login string,
	password string,
	metadata string,
) (*entities.Secret, error) {
	current, err := s.GetSecret(ctx, userID, name)
	if err != nil {
		return nil, err
	}

	secret := entities.Secret{
		Id:        current.Id,
		Owner:     userID,
		Name:      name,
		Login:     login,
		Password:  password,
		Metadata:  metadata,
		Checksum:  checksum.Secret(login, password, metadata),
		CreatedAt: current.CreatedAt,
	}
	updatedAt, err := s.uow.Secrets().UpdateSecret(ctx, secret)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrSecretNotExists
	}
	if err != nil {
		return nil, fmt.Errorf("UpdateSecret: %w", err)
	}
	secret.UpdatedAt = updatedAt
	return &secret, nil
}

// DeleteSecret удаляет секрет пользователя по имени.
// Возвращает ErrSecretNotExists, если секрета с таким именем нет.
func (s *SecretsService) DeleteSecret(ctx context.Context, userID int64, name string) error {
	err := s.uow.Secrets().DeleteSecret(ctx, userID, name)
	if errors.Is(err, repository.ErrNoRows) {
		return ErrSecretNotExists
	}
	if err != nil {
		return fmt.Errorf("DeleteSecret: %w", err)
	}
	return nil
}
