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

// --- CreateCard ---

func TestCreateCard_Success(t *testing.T) {
	cs := &mockCardsService{
		createFn: func(_ context.Context, _ int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error) {
			return &entities.Card{Name: name, Number: number, Holder: holder, ExpiresAt: expiresAt, CVV: cvv, Metadata: metadata}, nil
		},
	}
	h := testCardHandler(cs)

	body := `{"name":"n","number":"4111","holder":"h","expires_at":"12/30","cvv":"123","metadata":"m"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader(body)), 1)
	h.CreateCard(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.CardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestCreateCard_Unauthorized(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader(`{}`))
	h.CreateCard(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateCard_BadJSON(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader("not-json")), 1)
	h.CreateCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCard_NameRequired(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader(`{"number":"4111"}`)), 1)
	h.CreateCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCard_AlreadyExists(t *testing.T) {
	cs := &mockCardsService{
		createFn: func(_ context.Context, _ int64, _, _, _, _, _, _ string) (*entities.Card, error) {
			return nil, service.ErrCardAlreadyExists
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateCard(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateCard_UnexpectedError(t *testing.T) {
	cs := &mockCardsService{
		createFn: func(_ context.Context, _ int64, _, _, _, _, _, _ string) (*entities.Card, error) {
			return nil, assert.AnError
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/card", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateCard(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetCards ---

func TestGetCards_Success(t *testing.T) {
	cs := &mockCardsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Card, error) {
			return []entities.Card{{Name: "a"}, {Name: "b"}}, nil
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card", nil), 1)
	h.GetCards(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetCardsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestGetCards_Unauthorized(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/card", nil)
	h.GetCards(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCards_UnexpectedError(t *testing.T) {
	cs := &mockCardsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Card, error) {
			return nil, assert.AnError
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card", nil), 1)
	h.GetCards(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetCard ---

func TestGetCard_Success(t *testing.T) {
	cs := &mockCardsService{
		getFn: func(_ context.Context, _ int64, name string) (*entities.Card, error) {
			return &entities.Card{Name: name}, nil
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.GetCard(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.CardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestGetCard_Unauthorized(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/card/n", nil)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.GetCard(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCard_MissingName(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card/", nil), 1)
	h.GetCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCard_NotFound(t *testing.T) {
	cs := &mockCardsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Card, error) {
			return nil, service.ErrCardNotExists
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.GetCard(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCard_UnexpectedError(t *testing.T) {
	cs := &mockCardsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Card, error) {
			return nil, assert.AnError
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.GetCard(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateCard ---

func TestUpdateCard_Success(t *testing.T) {
	cs := &mockCardsService{
		updateFn: func(_ context.Context, _ int64, name, number, holder, expiresAt, cvv, metadata string) (*entities.Card, error) {
			return &entities.Card{Name: name, Number: number, Holder: holder, ExpiresAt: expiresAt, CVV: cvv, Metadata: metadata}, nil
		},
	}
	h := testCardHandler(cs)

	body := `{"number":"4222","holder":"h2","expires_at":"01/31","cvv":"999","metadata":"m2"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/card/n", strings.NewReader(body)), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.UpdateCard(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.CardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "4222", resp.Number)
}

func TestUpdateCard_Unauthorized(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/card/n", strings.NewReader(`{}`))
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.UpdateCard(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateCard_MissingName(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/card/", strings.NewReader(`{}`)), 1)
	h.UpdateCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCard_BadJSON(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/card/n", strings.NewReader("not-json")), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.UpdateCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCard_NotFound(t *testing.T) {
	cs := &mockCardsService{
		updateFn: func(_ context.Context, _ int64, _, _, _, _, _, _ string) (*entities.Card, error) {
			return nil, service.ErrCardNotExists
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/card/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.UpdateCard(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateCard_UnexpectedError(t *testing.T) {
	cs := &mockCardsService{
		updateFn: func(_ context.Context, _ int64, _, _, _, _, _, _ string) (*entities.Card, error) {
			return nil, assert.AnError
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/card/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.UpdateCard(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- DeleteCard ---

func TestDeleteCard_Success(t *testing.T) {
	cs := &mockCardsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return nil
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.DeleteCard(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteCard_Unauthorized(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/card/n", nil)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.DeleteCard(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteCard_MissingName(t *testing.T) {
	h := testCardHandler(&mockCardsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/card/", nil), 1)
	h.DeleteCard(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCard_NotFound(t *testing.T) {
	cs := &mockCardsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return service.ErrCardNotExists
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.DeleteCard(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCard_UnexpectedError(t *testing.T) {
	cs := &mockCardsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return assert.AnError
		},
	}
	h := testCardHandler(cs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/card/n", nil), 1)
	r = withPathParams(r, map[string]string{"card_name": "n"})
	h.DeleteCard(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
