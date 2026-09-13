package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/server/middlewares"
	"github.com/scarypuppp/gophkeeper/internal/server/service"
	"go.uber.org/zap"
)

// maxUploadSize ограничивает размер тела запроса,
// multipartMemory — объём формы, который держим в памяти.
const (
	maxUploadSize   = api.MaxFileUploadSize
	multipartMemory = 32 << 20 // 32 MiB
)

// multipartOverhead — запас на заголовки частей формы и поле metadata.
const multipartOverhead = 1 << 20 // 1 MiB

var uploadTooLargeMessage = fmt.Sprintf("file is too large: limit is %d bytes", int64(maxUploadSize))

// CREATE FILE

// CreateFile godoc
//
//	@Summary		Загрузка файла
//	@Description	Загружает файл в S3-хранилище и сохраняет его метаданные для текущего пользователя
//	@Tags			file
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file	true	"Загружаемый файл"
//	@Param			metadata	formData	string	false	"Произвольные метаданные"
//	@Success		200		{object}	api.CreateFileResponse
//	@Failure		400		{string}	string	"Неверный формат запроса"
//	@Failure		409		{string}	string	"Файл с таким содержимым уже загружен"
//	@Failure		413		{string}	string	"Файл превышает допустимый размер"
//	@Failure		500		{string}	string	"Внутренняя ошибка"
//	@Router			/api/file/upload [post]
func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	file, header, ok := openUploadedFile(w, r)
	if !ok {
		return
	}
	defer file.Close()

	created, err := h.fileService.CreateFile(
		r.Context(),
		userID,
		header.Filename,
		r.FormValue("metadata"),
		header.Size,
		file,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFileAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler CreateFile unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, api.CreateFileResponse{
		FileName:  created.FileName,
		FileHash:  created.FileHash,
		Metadata:  created.Metadata,
		Checksum:  created.Checksum,
		Size:      created.Size,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	})
}

// GET FILE

// GetFile godoc
//
//	@Summary		Получение файла
//	@Description	Возвращает содержимое файла пользователя по хэшу и имени файла
//	@Tags			file
//	@Security		BearerAuth
//	@Produce		octet-stream
//	@Param			file_hash	path		string	true	"Хэш файла"
//	@Param			file_name	path		string	true	"Имя файла"
//	@Success		200			{file}		binary	"Содержимое файла"
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		404			{string}	string	"Файл не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/file/{file_hash}/{file_name} [get]
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	fileHash := pathParam(r, "file_hash")
	fileName := pathParam(r, "file_name")
	if fileHash == "" || fileName == "" {
		http.Error(w, "file_hash and file_name path parameters are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	file, err := h.fileService.GetFile(ctx, userID, fileHash, fileName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFileNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler GetFile unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	body, err := h.fileService.GetFileContent(ctx, file.StorageFilePath)
	if err != nil {
		h.logger.Error("Handler GetFile unhandled error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(file.FileName)))
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := io.Copy(w, body); err != nil {
		http.Error(w, "Failed to stream file", http.StatusInternalServerError)
	}
}

// GET FILES

// GetFiles godoc
//
//	@Summary		Список файлов пользователя
//	@Description	Возвращает список файлов, загруженных текущим пользователем
//	@Tags			file
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	api.GetFilesResponse
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/file [get]
func (h *Handler) GetFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	files, err := h.fileService.GetFiles(r.Context(), userID)
	if err != nil {
		h.logger.Error("Handler GetFiles unhandled error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items := make(api.GetFilesResponse, 0, len(files))
	for _, file := range files {
		items = append(items, api.FileItemResponse{
			FileName:  file.FileName,
			FileHash:  file.FileHash,
			FileURL:   fmt.Sprintf("/api/file/%s/%s", url.PathEscape(file.FileHash), url.PathEscape(file.FileName)),
			Metadata:  file.Metadata,
			Checksum:  file.Checksum,
			Size:      file.Size,
			CreatedAt: file.CreatedAt,
			UpdatedAt: file.UpdatedAt,
		})
	}

	writeJSON(w, h.logger, items)
}

// UPDATE FILE METADATA

// UpdateFileMetadata godoc
//
//	@Summary		Обновление метаданных файла
//	@Description	Меняет только метаданные файла, содержимое в хранилище остаётся прежним
//	@Tags			file
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			file_hash	path		string					true	"Хэш файла"
//	@Param			file_name	path		string					true	"Имя файла"
//	@Param			body		body		api.UpdateFileRequest	true	"Новые метаданные"
//	@Success		200			{object}	api.FileResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Файл не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/file/{file_hash}/{file_name}/metadata [put]
func (h *Handler) UpdateFileMetadata(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	fileHash := pathParam(r, "file_hash")
	fileName := pathParam(r, "file_name")
	if fileHash == "" || fileName == "" {
		http.Error(w, "file_hash and file_name path parameters are required", http.StatusBadRequest)
		return
	}

	var requestData api.UpdateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, err := h.fileService.UpdateFileMetadata(r.Context(), userID, fileHash, fileName, requestData.Metadata)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFileNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler UpdateFileMetadata unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, api.FileResponse{
		FileName:  file.FileName,
		FileHash:  file.FileHash,
		Metadata:  file.Metadata,
		Checksum:  file.Checksum,
		Size:      file.Size,
		CreatedAt: file.CreatedAt,
		UpdatedAt: file.UpdatedAt,
	})
}

// UPDATE FILE CONTENT

// UpdateFile godoc
//
//	@Summary		Обновление файла
//	@Description	Перезаписывает содержимое файла пользователя. Хэш, имя и метаданные остаются прежними,
//	@Description	обновляются контрольная сумма, размер и время загрузки
//	@Tags			file
//	@Security		BearerAuth
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file_hash	path		string	true	"Хэш файла"
//	@Param			file_name	path		string	true	"Имя файла"
//	@Param			file		formData	file	true	"Новое содержимое файла"
//	@Success		200			{object}	api.FileResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Файл не найден"
//	@Failure		413			{string}	string	"Файл превышает допустимый размер"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/file/{file_hash}/{file_name} [put]
func (h *Handler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	fileHash := pathParam(r, "file_hash")
	fileName := pathParam(r, "file_name")
	if fileHash == "" || fileName == "" {
		http.Error(w, "file_hash and file_name path parameters are required", http.StatusBadRequest)
		return
	}

	content, header, ok := openUploadedFile(w, r)
	if !ok {
		return
	}
	defer content.Close()

	file, err := h.fileService.UpdateFileContent(
		r.Context(),
		userID,
		fileHash,
		fileName,
		header.Size,
		content,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFileNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler UpdateFile unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, api.FileResponse{
		FileName:  file.FileName,
		FileHash:  file.FileHash,
		Metadata:  file.Metadata,
		Checksum:  file.Checksum,
		Size:      file.Size,
		CreatedAt: file.CreatedAt,
		UpdatedAt: file.UpdatedAt,
	})
}

// DELETE FILE

// DeleteFile godoc
//
//	@Summary		Удаление файла
//	@Description	Удаляет файл пользователя из хранилища вместе с его метаданными
//	@Tags			file
//	@Security		BearerAuth
//	@Produce		plain
//	@Param			file_hash	path	string	true	"Хэш файла"
//	@Param			file_name	path	string	true	"Имя файла"
//	@Success		204
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404	{string}	string	"Файл не найден"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/file/{file_hash}/{file_name} [delete]
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	fileHash := pathParam(r, "file_hash")
	fileName := pathParam(r, "file_name")
	if fileHash == "" || fileName == "" {
		http.Error(w, "file_hash and file_name path parameters are required", http.StatusBadRequest)
		return
	}

	if err := h.fileService.DeleteFile(r.Context(), userID, fileHash, fileName); err != nil {
		switch {
		case errors.Is(err, service.ErrFileNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler DeleteFile unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// openUploadedFile разбирает multipart-форму и отдаёт поле "file".
// При ошибке сама пишет ответ и возвращает false.
func openUploadedFile(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+multipartOverhead)

	if err := r.ParseMultipartForm(multipartMemory); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, uploadTooLargeMessage, http.StatusRequestEntityTooLarge)
			return nil, nil, false
		}
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return nil, nil, false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file form field is required", http.StatusBadRequest)
		return nil, nil, false
	}
	if header.Size > maxUploadSize {
		file.Close()
		http.Error(w, uploadTooLargeMessage, http.StatusRequestEntityTooLarge)
		return nil, nil, false
	}
	return file, header, true
}
