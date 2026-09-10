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

// --- CreateSecret ---

func TestCreateSecret_Success(t *testing.T) {
	ss := &mockSecretsService{
		createFn: func(_ context.Context, _ int64, name, login, password, metadata string) (*entities.Secret, error) {
			return &entities.Secret{Name: name, Login: login, Password: password, Metadata: metadata}, nil
		},
	}
	h := testSecretHandler(ss)

	body := `{"name":"n","login":"l","password":"p","metadata":"m"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader(body)), 1)
	h.CreateSecret(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.SecretResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestCreateSecret_Unauthorized(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader(`{}`))
	h.CreateSecret(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateSecret_BadJSON(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader("not-json")), 1)
	h.CreateSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSecret_NameRequired(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader(`{"login":"l"}`)), 1)
	h.CreateSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSecret_AlreadyExists(t *testing.T) {
	ss := &mockSecretsService{
		createFn: func(_ context.Context, _ int64, _, _, _, _ string) (*entities.Secret, error) {
			return nil, service.ErrSecretAlreadyExists
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateSecret(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateSecret_UnexpectedError(t *testing.T) {
	ss := &mockSecretsService{
		createFn: func(_ context.Context, _ int64, _, _, _, _ string) (*entities.Secret, error) {
			return nil, assert.AnError
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/secret", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateSecret(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetSecrets ---

func TestGetSecrets_Success(t *testing.T) {
	ss := &mockSecretsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Secret, error) {
			return []entities.Secret{{Name: "a"}, {Name: "b"}}, nil
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret", nil), 1)
	h.GetSecrets(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetSecretsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestGetSecrets_Unauthorized(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/secret", nil)
	h.GetSecrets(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSecrets_UnexpectedError(t *testing.T) {
	ss := &mockSecretsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Secret, error) {
			return nil, assert.AnError
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret", nil), 1)
	h.GetSecrets(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetSecret ---

func TestGetSecret_Success(t *testing.T) {
	ss := &mockSecretsService{
		getFn: func(_ context.Context, _ int64, name string) (*entities.Secret, error) {
			return &entities.Secret{Name: name}, nil
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.GetSecret(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.SecretResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestGetSecret_Unauthorized(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/secret/n", nil)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.GetSecret(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSecret_MissingName(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret/", nil), 1)
	h.GetSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSecret_NotFound(t *testing.T) {
	ss := &mockSecretsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Secret, error) {
			return nil, service.ErrSecretNotExists
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.GetSecret(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSecret_UnexpectedError(t *testing.T) {
	ss := &mockSecretsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Secret, error) {
			return nil, assert.AnError
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.GetSecret(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateSecret ---

func TestUpdateSecret_Success(t *testing.T) {
	ss := &mockSecretsService{
		updateFn: func(_ context.Context, _ int64, name, login, password, metadata string) (*entities.Secret, error) {
			return &entities.Secret{Name: name, Login: login, Password: password, Metadata: metadata}, nil
		},
	}
	h := testSecretHandler(ss)

	body := `{"login":"l2","password":"p2","metadata":"m2"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/secret/n", strings.NewReader(body)), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.UpdateSecret(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.SecretResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "l2", resp.Login)
}

func TestUpdateSecret_Unauthorized(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/secret/n", strings.NewReader(`{}`))
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.UpdateSecret(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateSecret_MissingName(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/secret/", strings.NewReader(`{}`)), 1)
	h.UpdateSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSecret_BadJSON(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/secret/n", strings.NewReader("not-json")), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.UpdateSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSecret_NotFound(t *testing.T) {
	ss := &mockSecretsService{
		updateFn: func(_ context.Context, _ int64, _, _, _, _ string) (*entities.Secret, error) {
			return nil, service.ErrSecretNotExists
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/secret/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.UpdateSecret(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateSecret_UnexpectedError(t *testing.T) {
	ss := &mockSecretsService{
		updateFn: func(_ context.Context, _ int64, _, _, _, _ string) (*entities.Secret, error) {
			return nil, assert.AnError
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/secret/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.UpdateSecret(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- DeleteSecret ---

func TestDeleteSecret_Success(t *testing.T) {
	ss := &mockSecretsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return nil
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.DeleteSecret(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteSecret_Unauthorized(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/secret/n", nil)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.DeleteSecret(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteSecret_MissingName(t *testing.T) {
	h := testSecretHandler(&mockSecretsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/secret/", nil), 1)
	h.DeleteSecret(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteSecret_NotFound(t *testing.T) {
	ss := &mockSecretsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return service.ErrSecretNotExists
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.DeleteSecret(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteSecret_UnexpectedError(t *testing.T) {
	ss := &mockSecretsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return assert.AnError
		},
	}
	h := testSecretHandler(ss)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/secret/n", nil), 1)
	r = withPathParams(r, map[string]string{"secret_name": "n"})
	h.DeleteSecret(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
