package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/scarypuppp/gophkeeper/internal/server/auth"
)

type contextKey string

// userIDKey ключ контекста, по которому middleware Auth сохраняет идентификатор пользователя.
const userIDKey contextKey = "userID"

// ContextWithUserID возвращает контекст с сохранённым идентификатором пользователя.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext возвращает идентификатор пользователя из контекста.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// Auth middleware проверяет наличие валидного Bearer-токена в заголовке Authorization.
// При успешной проверке сохраняет идентификатор пользователя в контексте запроса.
// Возвращает HTTP 401, если токен отсутствует, имеет неверный формат или невалиден.
func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		userID, err := auth.ParseToken(m.cfg.SecretKey, parts[1])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ctx := ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
