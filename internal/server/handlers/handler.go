package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/config"
	"go.uber.org/zap"
)

type userService interface {
	RegisterUser(ctx context.Context, login, password string) (*entities.User, error)
	LoginUser(ctx context.Context, login, password string) (*entities.User, error)
}

type fileService interface {
	GetFileByName(ctx context.Context, userID int64, fileName string) (*entities.File, error)
	GetFile(ctx context.Context, userID int64, fileHash, fileName string) (*entities.File, error)
	GetFileContent(ctx context.Context, storagePath string) (io.ReadCloser, error)
	GetFiles(ctx context.Context, userID int64) ([]entities.File, error)
	CreateFile(ctx context.Context, userID int64, fileName, metadata string, size int64, file io.Reader) (*entities.File, error)
	UpdateFileMetadata(ctx context.Context, userID int64, fileHash, fileName, metadata string) (*entities.File, error)
	UpdateFileContent(ctx context.Context, userID int64, fileHash, fileName string, size int64, content io.Reader) (*entities.File, error)
	DeleteFile(ctx context.Context, userID int64, fileHash, fileName string) error
}

type secretsService interface {
	GetSecret(ctx context.Context, userID int64, name string) (*entities.Secret, error)
	GetSecrets(ctx context.Context, userID int64) ([]entities.Secret, error)
	CreateSecret(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error)
	UpdateSecret(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error)
	DeleteSecret(ctx context.Context, userID int64, name string) error
}
type cardsService interface {
	GetCard(ctx context.Context, userID int64, name string) (*entities.Card, error)
	GetCards(ctx context.Context, userID int64) ([]entities.Card, error)
	CreateCard(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error)
	UpdateCard(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error)
	DeleteCard(ctx context.Context, userID int64, name string) error
}

type textsService interface {
	GetText(ctx context.Context, userID int64, name string) (*entities.Text, error)
	GetTexts(ctx context.Context, userID int64) ([]entities.Text, error)
	CreateText(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error)
	UpdateText(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error)
	DeleteText(ctx context.Context, userID int64, name string) error
}

type Handler struct {
	config         *config.Config
	logger         *zap.Logger
	userService    userService
	fileService    fileService
	secretsService secretsService
	cardsService   cardsService
	textsService   textsService
}

func NewHandler(
	cfg *config.Config,
	logger *zap.Logger,
	userService userService,
	fileService fileService,
	secretsService secretsService,
	cardsService cardsService,
	textsService textsService,
) *Handler {
	return &Handler{
		config:         cfg,
		logger:         logger,
		userService:    userService,
		fileService:    fileService,
		secretsService: secretsService,
		cardsService:   cardsService,
		textsService:   textsService,
	}
}

// pathParam возвращает параметр пути по имени.
// chi отдаёт сегмент как есть, если путь содержит закодированный '/', поэтому значение декодируется вручную.
func pathParam(r *http.Request, key string) string {
	value := chi.URLParam(r, key)
	if decoded, err := url.PathUnescape(value); err == nil {
		return decoded
	}
	return value
}

// writeJSON сериализует value и отправляет его клиенту со статусом 200.
func writeJSON(w http.ResponseWriter, logger *zap.Logger, value any) {
	response, err := json.Marshal(value)
	if err != nil {
		logger.Error("Handler writeJSON marshal error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
