package interactor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerUnavailable проверяет, что недоступный сервер отличается
// от любой другой ошибки и что в сообщении виден его адрес.
func TestServerUnavailable(t *testing.T) {
	// Порт 1 не слушает никто, соединение отвергается сразу.
	const address = "http://127.0.0.1:1"
	i := NewInteractor(resty.New().SetBaseURL(address))

	_, err := i.Login("alice", "password123")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrServerUnavailable)
	assert.Contains(t, err.Error(), address)

	_, err = i.GetSecrets("token")
	assert.ErrorIs(t, err, ErrServerUnavailable)

	err = i.DownloadFile("token", "hash", "name", nil)
	assert.ErrorIs(t, err, ErrServerUnavailable)
}

// TestServerErrorIsNotUnavailable проверяет, что ответ сервера с ошибкой
// не выдаётся за недоступность.
func TestServerErrorIsNotUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	i := NewInteractor(resty.New().SetBaseURL(server.URL))

	_, err := i.GetSecrets("token")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrServerUnavailable)
	assert.Contains(t, err.Error(), "500")
}
