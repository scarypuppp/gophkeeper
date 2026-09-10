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

// --- CreateText ---

func TestCreateText_Success(t *testing.T) {
	ts := &mockTextsService{
		createFn: func(_ context.Context, _ int64, name, text, metadata string) (*entities.Text, error) {
			return &entities.Text{Name: name, Text: text, Metadata: metadata}, nil
		},
	}
	h := testTextHandler(ts)

	body := `{"name":"n","text":"content","metadata":"m"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader(body)), 1)
	h.CreateText(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.TextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestCreateText_Unauthorized(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader(`{}`))
	h.CreateText(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateText_BadJSON(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader("not-json")), 1)
	h.CreateText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateText_NameRequired(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader(`{"text":"c"}`)), 1)
	h.CreateText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateText_AlreadyExists(t *testing.T) {
	ts := &mockTextsService{
		createFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.Text, error) {
			return nil, service.ErrTextAlreadyExists
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateText(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateText_UnexpectedError(t *testing.T) {
	ts := &mockTextsService{
		createFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.Text, error) {
			return nil, assert.AnError
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPost, "/api/text", strings.NewReader(`{"name":"n"}`)), 1)
	h.CreateText(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetTexts ---

func TestGetTexts_Success(t *testing.T) {
	ts := &mockTextsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Text, error) {
			return []entities.Text{{Name: "a"}, {Name: "b"}}, nil
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text", nil), 1)
	h.GetTexts(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetTextsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestGetTexts_Unauthorized(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/text", nil)
	h.GetTexts(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTexts_UnexpectedError(t *testing.T) {
	ts := &mockTextsService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.Text, error) {
			return nil, assert.AnError
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text", nil), 1)
	h.GetTexts(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetText ---

func TestGetText_Success(t *testing.T) {
	ts := &mockTextsService{
		getFn: func(_ context.Context, _ int64, name string) (*entities.Text, error) {
			return &entities.Text{Name: name}, nil
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.GetText(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.TextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "n", resp.Name)
}

func TestGetText_Unauthorized(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/text/n", nil)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.GetText(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetText_MissingName(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text/", nil), 1)
	h.GetText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetText_NotFound(t *testing.T) {
	ts := &mockTextsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Text, error) {
			return nil, service.ErrTextNotExists
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.GetText(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetText_UnexpectedError(t *testing.T) {
	ts := &mockTextsService{
		getFn: func(_ context.Context, _ int64, _ string) (*entities.Text, error) {
			return nil, assert.AnError
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.GetText(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateText ---

func TestUpdateText_Success(t *testing.T) {
	ts := &mockTextsService{
		updateFn: func(_ context.Context, _ int64, name, text, metadata string) (*entities.Text, error) {
			return &entities.Text{Name: name, Text: text, Metadata: metadata}, nil
		},
	}
	h := testTextHandler(ts)

	body := `{"text":"newcontent","metadata":"m2"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/text/n", strings.NewReader(body)), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.UpdateText(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.TextResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "newcontent", resp.Text)
}

func TestUpdateText_Unauthorized(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/text/n", strings.NewReader(`{}`))
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.UpdateText(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateText_MissingName(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/text/", strings.NewReader(`{}`)), 1)
	h.UpdateText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateText_BadJSON(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/text/n", strings.NewReader("not-json")), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.UpdateText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateText_NotFound(t *testing.T) {
	ts := &mockTextsService{
		updateFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.Text, error) {
			return nil, service.ErrTextNotExists
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/text/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.UpdateText(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateText_UnexpectedError(t *testing.T) {
	ts := &mockTextsService{
		updateFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.Text, error) {
			return nil, assert.AnError
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/text/n", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.UpdateText(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- DeleteText ---

func TestDeleteText_Success(t *testing.T) {
	ts := &mockTextsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return nil
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.DeleteText(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteText_Unauthorized(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/text/n", nil)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.DeleteText(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteText_MissingName(t *testing.T) {
	h := testTextHandler(&mockTextsService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/text/", nil), 1)
	h.DeleteText(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteText_NotFound(t *testing.T) {
	ts := &mockTextsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return service.ErrTextNotExists
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.DeleteText(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteText_UnexpectedError(t *testing.T) {
	ts := &mockTextsService{
		deleteFn: func(_ context.Context, _ int64, _ string) error {
			return assert.AnError
		},
	}
	h := testTextHandler(ts)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/text/n", nil), 1)
	r = withPathParams(r, map[string]string{"text_name": "n"})
	h.DeleteText(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
