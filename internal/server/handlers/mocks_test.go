package handlers

import (
	"context"
	"net/http"

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

func testHandler(us userService) *Handler {
	cfg := &config.Config{SecretKey: "test", TokenExpSeconds: 3600}
	return NewHandler(cfg, zap.NewNop(), us, nil, nil, nil, nil)
}

func withUserID(r *http.Request, id int64) *http.Request {
	ctx := context.WithValue(r.Context(), middlewares.UserIDKey, id)
	return r.WithContext(ctx)
}
