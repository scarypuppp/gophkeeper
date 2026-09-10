package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/config"
	"github.com/scarypuppp/gophkeeper/internal/server/middlewares"
	"go.uber.org/zap"
)

type mockUserService struct {
	registerFn func(ctx context.Context, login, password string) (*entities.User, error)
	loginFn    func(ctx context.Context, login, password string) (*entities.User, error)
}

func (m *mockUserService) RegisterUser(ctx context.Context, login, password string) (*entities.User, error) {
	return m.registerFn(ctx, login, password)
}
func (m *mockUserService) LoginUser(ctx context.Context, login, password string) (*entities.User, error) {
	return m.loginFn(ctx, login, password)
}

type mockSecretsService struct {
	getFn    func(ctx context.Context, userID int64, name string) (*entities.Secret, error)
	getAllFn func(ctx context.Context, userID int64) ([]entities.Secret, error)
	createFn func(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error)
	updateFn func(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error)
	deleteFn func(ctx context.Context, userID int64, name string) error
}

func (m *mockSecretsService) GetSecret(ctx context.Context, userID int64, name string) (*entities.Secret, error) {
	return m.getFn(ctx, userID, name)
}
func (m *mockSecretsService) GetSecrets(ctx context.Context, userID int64) ([]entities.Secret, error) {
	return m.getAllFn(ctx, userID)
}
func (m *mockSecretsService) CreateSecret(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error) {
	return m.createFn(ctx, userID, name, login, password, metadata)
}
func (m *mockSecretsService) UpdateSecret(ctx context.Context, userID int64, name, login, password, metadata string) (*entities.Secret, error) {
	return m.updateFn(ctx, userID, name, login, password, metadata)
}
func (m *mockSecretsService) DeleteSecret(ctx context.Context, userID int64, name string) error {
	return m.deleteFn(ctx, userID, name)
}

type mockCardsService struct {
	getFn    func(ctx context.Context, userID int64, name string) (*entities.Card, error)
	getAllFn func(ctx context.Context, userID int64) ([]entities.Card, error)
	createFn func(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error)
	updateFn func(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error)
	deleteFn func(ctx context.Context, userID int64, name string) error
}

func (m *mockCardsService) GetCard(ctx context.Context, userID int64, name string) (*entities.Card, error) {
	return m.getFn(ctx, userID, name)
}
func (m *mockCardsService) GetCards(ctx context.Context, userID int64) ([]entities.Card, error) {
	return m.getAllFn(ctx, userID)
}
func (m *mockCardsService) CreateCard(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error) {
	return m.createFn(ctx, userID, name, number, holder, expiresAt, cvv, metadata)
}
func (m *mockCardsService) UpdateCard(ctx context.Context, userID int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error) {
	return m.updateFn(ctx, userID, name, number, holder, expiresAt, cvv, metadata)
}
func (m *mockCardsService) DeleteCard(ctx context.Context, userID int64, name string) error {
	return m.deleteFn(ctx, userID, name)
}

type mockTextsService struct {
	getFn    func(ctx context.Context, userID int64, name string) (*entities.Text, error)
	getAllFn func(ctx context.Context, userID int64) ([]entities.Text, error)
	createFn func(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error)
	updateFn func(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error)
	deleteFn func(ctx context.Context, userID int64, name string) error
}

func (m *mockTextsService) GetText(ctx context.Context, userID int64, name string) (*entities.Text, error) {
	return m.getFn(ctx, userID, name)
}
func (m *mockTextsService) GetTexts(ctx context.Context, userID int64) ([]entities.Text, error) {
	return m.getAllFn(ctx, userID)
}
func (m *mockTextsService) CreateText(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error) {
	return m.createFn(ctx, userID, name, text, metadata)
}
func (m *mockTextsService) UpdateText(ctx context.Context, userID int64, name, text, metadata string) (*entities.Text, error) {
	return m.updateFn(ctx, userID, name, text, metadata)
}
func (m *mockTextsService) DeleteText(ctx context.Context, userID int64, name string) error {
	return m.deleteFn(ctx, userID, name)
}

type mockFileService struct {
	getByNameFn      func(ctx context.Context, userID int64, fileHash string) (*entities.File, error)
	getFn            func(ctx context.Context, userID int64, fileHash, fileName string) (*entities.File, error)
	getContentFn     func(ctx context.Context, storagePath string) (io.ReadCloser, error)
	getAllFn         func(ctx context.Context, userID int64) ([]entities.File, error)
	createFn         func(ctx context.Context, userID int64, fileName, metadata string, size int64, file io.Reader) (*entities.File, error)
	updateMetadataFn func(ctx context.Context, userID int64, fileHash, fileName, metadata string) (*entities.File, error)
	updateContentFn  func(ctx context.Context, userID int64, fileHash, fileName string, size int64, content io.Reader) (*entities.File, error)
	deleteFn         func(ctx context.Context, userID int64, fileHash, fileName string) error
}

func (m *mockFileService) GetFileByName(ctx context.Context, userID int64, fileName string) (*entities.File, error) {
	return m.getByNameFn(ctx, userID, fileName)
}
func (m *mockFileService) GetFile(ctx context.Context, userID int64, fileHash, fileName string) (*entities.File, error) {
	return m.getFn(ctx, userID, fileHash, fileName)
}
func (m *mockFileService) GetFileContent(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	return m.getContentFn(ctx, storagePath)
}
func (m *mockFileService) GetFiles(ctx context.Context, userID int64) ([]entities.File, error) {
	return m.getAllFn(ctx, userID)
}
func (m *mockFileService) CreateFile(ctx context.Context, userID int64, fileName, metadata string, size int64, file io.Reader) (*entities.File, error) {
	return m.createFn(ctx, userID, fileName, metadata, size, file)
}
func (m *mockFileService) UpdateFileMetadata(ctx context.Context, userID int64, fileHash, fileName, metadata string) (*entities.File, error) {
	return m.updateMetadataFn(ctx, userID, fileHash, fileName, metadata)
}
func (m *mockFileService) UpdateFileContent(ctx context.Context, userID int64, fileHash, fileName string, size int64, content io.Reader) (*entities.File, error) {
	return m.updateContentFn(ctx, userID, fileHash, fileName, size, content)
}
func (m *mockFileService) DeleteFile(ctx context.Context, userID int64, fileHash, fileName string) error {
	return m.deleteFn(ctx, userID, fileHash, fileName)
}

func testHandler(us userService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), us, nil, nil, nil, nil)
}

func testSecretHandler(ss secretsService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), nil, nil, ss, nil, nil)
}

func testCardHandler(cs cardsService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), nil, nil, nil, cs, nil)
}

func testTextHandler(ts textsService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), nil, nil, nil, nil, ts)
}

func testFileHandler(fs fileService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), nil, fs, nil, nil, nil)
}

func withUserID(r *http.Request, id int64) *http.Request {
	ctx := middlewares.ContextWithUserID(r.Context(), id)
	return r.WithContext(ctx)
}

// withPathParams подставляет в запрос chi route-контекст с указанными параметрами пути,
// имитируя то, что обычно делает роутер перед вызовом хендлера.
func withPathParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}
