package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

// newMultipartRequest строит multipart-запрос с полем "file" и опциональными
// текстовыми полями (например, "metadata"), как это делает клиент при загрузке файла.
func newMultipartRequest(t *testing.T, method, url, fileName, content string, fields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, value := range fields {
		require.NoError(t, w.WriteField(key, value))
	}

	part, err := w.CreateFormFile("file", fileName)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	r := httptest.NewRequest(method, url, &buf)
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r
}

// --- CreateFile ---

func TestCreateFile_Success(t *testing.T) {
	fs := &mockFileService{
		createFn: func(_ context.Context, _ int64, fileName, metadata string, size int64, _ io.Reader) (*entities.File, error) {
			return &entities.File{FileName: fileName, Metadata: metadata, Size: size}, nil
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPost, "/api/file/upload", "doc.txt", "content", map[string]string{"metadata": "m"})
	r = withUserID(r, 1)
	w := httptest.NewRecorder()
	h.CreateFile(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.CreateFileResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "doc.txt", resp.FileName)
	assert.Equal(t, "m", resp.Metadata)
}

func TestCreateFile_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	r := newMultipartRequest(t, http.MethodPost, "/api/file/upload", "doc.txt", "content", nil)
	w := httptest.NewRecorder()
	h.CreateFile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateFile_MissingFileField(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.WriteField("metadata", "m"))
	require.NoError(t, mw.Close())

	r := httptest.NewRequest(http.MethodPost, "/api/file/upload", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r = withUserID(r, 1)
	w := httptest.NewRecorder()
	h.CreateFile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateFile_AlreadyExists(t *testing.T) {
	fs := &mockFileService{
		createFn: func(_ context.Context, _ int64, _, _ string, _ int64, _ io.Reader) (*entities.File, error) {
			return nil, service.ErrFileAlreadyExists
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPost, "/api/file/upload", "doc.txt", "content", nil)
	r = withUserID(r, 1)
	w := httptest.NewRecorder()
	h.CreateFile(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCreateFile_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		createFn: func(_ context.Context, _ int64, _, _ string, _ int64, _ io.Reader) (*entities.File, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPost, "/api/file/upload", "doc.txt", "content", nil)
	r = withUserID(r, 1)
	w := httptest.NewRecorder()
	h.CreateFile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetFile ---

func TestGetFile_Success(t *testing.T) {
	fs := &mockFileService{
		getFn: func(_ context.Context, _ int64, fileHash, fileName string) (*entities.File, error) {
			return &entities.File{FileHash: fileHash, FileName: fileName, StorageFilePath: "path"}, nil
		},
		getContentFn: func(_ context.Context, storagePath string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("data")), nil
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.GetFile(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "data", w.Body.String())
}

func TestGetFile_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/file/hash/name", nil)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.GetFile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFile_MissingParams(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file//name", nil), 1)
	r = withPathParams(r, map[string]string{"file_name": "name"})
	h.GetFile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetFile_NotFound(t *testing.T) {
	fs := &mockFileService{
		getFn: func(_ context.Context, _ int64, _, _ string) (*entities.File, error) {
			return nil, service.ErrFileNotExists
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.GetFile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetFile_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		getFn: func(_ context.Context, _ int64, _, _ string) (*entities.File, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.GetFile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetFile_ContentError(t *testing.T) {
	fs := &mockFileService{
		getFn: func(_ context.Context, _ int64, fileHash, fileName string) (*entities.File, error) {
			return &entities.File{FileHash: fileHash, FileName: fileName, StorageFilePath: "path"}, nil
		},
		getContentFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.GetFile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetFiles ---

func TestGetFiles_Success(t *testing.T) {
	fs := &mockFileService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.File, error) {
			return []entities.File{{FileName: "a"}, {FileName: "b"}}, nil
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file", nil), 1)
	h.GetFiles(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetFilesResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestGetFiles_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/file", nil)
	h.GetFiles(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFiles_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		getAllFn: func(_ context.Context, _ int64) ([]entities.File, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodGet, "/api/file", nil), 1)
	h.GetFiles(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateFileMetadata ---

func TestUpdateFileMetadata_Success(t *testing.T) {
	fs := &mockFileService{
		updateMetadataFn: func(_ context.Context, _ int64, fileHash, fileName, metadata string) (*entities.File, error) {
			return &entities.File{FileHash: fileHash, FileName: fileName, Metadata: metadata}, nil
		},
	}
	h := testFileHandler(fs)

	body := `{"metadata":"new"}`
	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/file/hash/name/metadata", strings.NewReader(body)), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.UpdateFileMetadata(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.FileResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "new", resp.Metadata)
}

func TestUpdateFileMetadata_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/file/hash/name/metadata", strings.NewReader(`{}`))
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.UpdateFileMetadata(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateFileMetadata_MissingParams(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/file//name/metadata", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"file_name": "name"})
	h.UpdateFileMetadata(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFileMetadata_BadJSON(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/file/hash/name/metadata", strings.NewReader("not-json")), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.UpdateFileMetadata(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFileMetadata_NotFound(t *testing.T) {
	fs := &mockFileService{
		updateMetadataFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.File, error) {
			return nil, service.ErrFileNotExists
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/file/hash/name/metadata", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.UpdateFileMetadata(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateFileMetadata_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		updateMetadataFn: func(_ context.Context, _ int64, _, _, _ string) (*entities.File, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodPut, "/api/file/hash/name/metadata", strings.NewReader(`{}`)), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.UpdateFileMetadata(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- UpdateFile (content) ---

func TestUpdateFile_Success(t *testing.T) {
	fs := &mockFileService{
		updateContentFn: func(_ context.Context, _ int64, fileHash, fileName string, size int64, _ io.Reader) (*entities.File, error) {
			return &entities.File{FileHash: fileHash, FileName: fileName, Size: size, Checksum: "newsum"}, nil
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPut, "/api/file/hash/name", "name", "newcontent", nil)
	r = withUserID(r, 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.FileResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "newsum", resp.Checksum)
}

func TestUpdateFile_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	r := newMultipartRequest(t, http.MethodPut, "/api/file/hash/name", "name", "content", nil)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateFile_MissingParams(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	r := newMultipartRequest(t, http.MethodPut, "/api/file//name", "name", "content", nil)
	r = withUserID(r, 1)
	r = withPathParams(r, map[string]string{"file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFile_MissingFileField(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.Close())

	r := httptest.NewRequest(http.MethodPut, "/api/file/hash/name", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r = withUserID(r, 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateFile_NotFound(t *testing.T) {
	fs := &mockFileService{
		updateContentFn: func(_ context.Context, _ int64, _, _ string, _ int64, _ io.Reader) (*entities.File, error) {
			return nil, service.ErrFileNotExists
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPut, "/api/file/hash/name", "name", "content", nil)
	r = withUserID(r, 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateFile_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		updateContentFn: func(_ context.Context, _ int64, _, _ string, _ int64, _ io.Reader) (*entities.File, error) {
			return nil, assert.AnError
		},
	}
	h := testFileHandler(fs)

	r := newMultipartRequest(t, http.MethodPut, "/api/file/hash/name", "name", "content", nil)
	r = withUserID(r, 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	w := httptest.NewRecorder()
	h.UpdateFile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- DeleteFile ---

func TestDeleteFile_Success(t *testing.T) {
	fs := &mockFileService{
		deleteFn: func(_ context.Context, _ int64, _, _ string) error {
			return nil
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.DeleteFile(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteFile_Unauthorized(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/file/hash/name", nil)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.DeleteFile(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteFile_MissingParams(t *testing.T) {
	h := testFileHandler(&mockFileService{})

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/file//name", nil), 1)
	r = withPathParams(r, map[string]string{"file_name": "name"})
	h.DeleteFile(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteFile_NotFound(t *testing.T) {
	fs := &mockFileService{
		deleteFn: func(_ context.Context, _ int64, _, _ string) error {
			return service.ErrFileNotExists
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.DeleteFile(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteFile_UnexpectedError(t *testing.T) {
	fs := &mockFileService{
		deleteFn: func(_ context.Context, _ int64, _, _ string) error {
			return assert.AnError
		},
	}
	h := testFileHandler(fs)

	w := httptest.NewRecorder()
	r := withUserID(httptest.NewRequest(http.MethodDelete, "/api/file/hash/name", nil), 1)
	r = withPathParams(r, map[string]string{"file_hash": "hash", "file_name": "name"})
	h.DeleteFile(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
