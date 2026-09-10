package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Register ---

func TestRegister_Success(t *testing.T) {
	us := &mockUserService{
		registerFn: func(_ context.Context, _, _ string) (*entities.User, error) {
			return &entities.User{ID: 1, Login: "alice"}, nil
		},
	}
	h := testHandler(us)

	body := `{"login":"alice","password":"secret"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(body))
	h.Register(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRegister_LoginExists(t *testing.T) {
	us := &mockUserService{
		registerFn: func(_ context.Context, _, _ string) (*entities.User, error) {
			return nil, service.ErrLoginAlreadyExists
		},
	}
	h := testHandler(us)

	body := `{"login":"alice","password":"secret"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(body))
	h.Register(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_BadJSON(t *testing.T) {
	h := testHandler(&mockUserService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader("not-json"))
	h.Register(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	us := &mockUserService{
		loginFn: func(_ context.Context, _, _ string) (*entities.User, error) {
			return &entities.User{ID: 7, Login: "alice"}, nil
		},
	}
	h := testHandler(us)

	body := `{"login":"alice","password":"secret"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(body))
	h.Login(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.LoginResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(7), resp.UserID)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestLogin_WrongCredentials(t *testing.T) {
	us := &mockUserService{
		loginFn: func(_ context.Context, _, _ string) (*entities.User, error) {
			return nil, service.ErrLoginPasswordNotExist
		},
	}
	h := testHandler(us)

	body := `{"login":"alice","password":"wrong"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(body))
	h.Login(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
