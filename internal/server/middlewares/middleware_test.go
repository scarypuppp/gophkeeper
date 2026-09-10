package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scarypuppp/gophkeeper/internal/server/auth"
	"github.com/scarypuppp/gophkeeper/internal/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testMiddleware() Middleware {
	return NewMiddleware(&config.Config{SecretKey: "test", TokenExpSeconds: 3600}, zap.NewNop())
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// --- Auth ---

func TestAuth_Success(t *testing.T) {
	token, err := auth.CreateToken("test", 42, 3600)
	require.NoError(t, err)

	mw := testMiddleware()
	var gotID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = r.Context().Value(UserIDKey).(int64)
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	mw.Auth(next).ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(42), gotID)
}

func TestAuth_NoHeader(t *testing.T) {
	mw := testMiddleware()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.Auth(http.HandlerFunc(okHandler)).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_BadFormat(t *testing.T) {
	mw := testMiddleware()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Token abc")
	mw.Auth(http.HandlerFunc(okHandler)).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	mw := testMiddleware()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer not.a.token")
	mw.Auth(http.HandlerFunc(okHandler)).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- LogRequest / LogResponse ---

func TestLogRequest(t *testing.T) {
	mw := testMiddleware()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, int64(1))
	mw.LogRequest(http.HandlerFunc(okHandler)).ServeHTTP(w, r.WithContext(ctx))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogResponse(t *testing.T) {
	mw := testMiddleware()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	ctx := context.WithValue(r.Context(), UserIDKey, int64(1))
	mw.LogResponse(http.HandlerFunc(okHandler)).ServeHTTP(w, r.WithContext(ctx))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}
