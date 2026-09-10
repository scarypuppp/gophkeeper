//go:generate mockgen -destination=mocks/mock_unit_of_work.go -package=mocks . UnitOfWork
//go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks . UserRepository
//go:generate mockgen -destination=mocks/mock_file_repository.go -package=mocks . FileRepository
//go:generate mockgen -destination=mocks/mock_secret_repository.go -package=mocks . SecretRepository
//go:generate mockgen -destination=mocks/mock_card_repository.go -package=mocks . CardRepository
//go:generate mockgen -destination=mocks/mock_text_repository.go -package=mocks . TextRepository
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/entities"
)

// ErrNoRows возвращается репозиторием, когда запрос не вернул ни одной строки.
var ErrNoRows = errors.New("no rows returned")

// UnitOfWork объединяет все репозитории и управляет транзакциями базы данных.
// BeginTx открывает новую транзакцию и возвращает UnitOfWork, работающий внутри неё.
type UnitOfWork interface {
	Users() UserRepository
	Files() FileRepository
	Secrets() SecretRepository
	Cards() CardRepository
	Texts() TextRepository

	BeginTx(ctx context.Context) (UnitOfWork, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// UserRepository интерфейс для операций с пользователями в хранилище.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByLogin(ctx context.Context, login string) (*entities.User, error)
	CreateUser(ctx context.Context, user entities.User) (*entities.User, error)
}

// FileRepository интерфейс для операций с метаданными файлов в хранилище.
type FileRepository interface {
	GetByName(ctx context.Context, owner int64, fileHash string) (*entities.File, error)
	GetByHashAndName(ctx context.Context, owner int64, fileHash, fileName string) (*entities.File, error)
	GetFilesByOwner(ctx context.Context, owner int64) ([]entities.File, error)
	CreateFile(ctx context.Context, file entities.File) (*entities.File, error)
	UpdateFileMetadata(ctx context.Context, owner int64, fileHash, fileName, metadata string) (time.Time, error)
	UpdateFileContent(ctx context.Context, owner int64, fileHash, fileName, checksum string, size int64) (time.Time, error)
	DeleteFile(ctx context.Context, owner int64, fileHash, fileName string) error
}

// SecretRepository интерфейс для операций с секретами в хранилище.
// Секрет идентифицируется парой владелец + имя.
type SecretRepository interface {
	GetByName(ctx context.Context, owner int64, name string) (*entities.Secret, error)
	GetSecretsByOwner(ctx context.Context, owner int64) ([]entities.Secret, error)
	CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, secret entities.Secret) (time.Time, error)
	DeleteSecret(ctx context.Context, owner int64, name string) error
}

// CardRepository интерфейс для операций с банковскими картами в хранилище.
// Карта идентифицируется парой владелец + имя.
type CardRepository interface {
	GetByName(ctx context.Context, owner int64, name string) (*entities.Card, error)
	GetCardsByOwner(ctx context.Context, owner int64) ([]entities.Card, error)
	CreateCard(ctx context.Context, card entities.Card) (*entities.Card, error)
	UpdateCard(ctx context.Context, card entities.Card) (time.Time, error)
	DeleteCard(ctx context.Context, owner int64, name string) error
}

// TextRepository интерфейс для операций с произвольными текстовыми данными в хранилище.
// Текст идентифицируется парой владелец + имя.
type TextRepository interface {
	GetByName(ctx context.Context, owner int64, name string) (*entities.Text, error)
	GetTextsByOwner(ctx context.Context, owner int64) ([]entities.Text, error)
	CreateText(ctx context.Context, text entities.Text) (*entities.Text, error)
	UpdateText(ctx context.Context, text entities.Text) (time.Time, error)
	DeleteText(ctx context.Context, owner int64, name string) error
}
