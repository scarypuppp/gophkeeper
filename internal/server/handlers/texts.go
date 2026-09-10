package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/middlewares"
	"github.com/scarypuppp/gophkeeper/internal/server/service"
	"go.uber.org/zap"
)

// CREATE TEXT

// CreateText godoc
//
//	@Summary		Создание текстовых данных
//	@Description	Сохраняет произвольные текстовые данные текущего пользователя. Имя уникально в рамках пользователя
//	@Tags			text
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.CreateTextRequest	true	"Текстовые данные"
//	@Success		200		{object}	api.TextResponse
//	@Failure		400		{string}	string	"Неверный формат запроса"
//	@Failure		401		{string}	string	"Пользователь не аутентифицирован"
//	@Failure		409		{string}	string	"Текст с таким именем уже существует"
//	@Failure		500		{string}	string	"Внутренняя ошибка"
//	@Router			/api/text [post]
func (h *Handler) CreateText(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var requestData api.CreateTextRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestData.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	text, err := h.textsService.CreateText(
		r.Context(),
		userID,
		requestData.Name,
		requestData.Text,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTextAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler CreateText unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, textResponse(text))
}

// GET TEXTS

// GetTexts godoc
//
//	@Summary		Список текстовых данных пользователя
//	@Description	Возвращает имена и метаданные текстов текущего пользователя без их содержимого
//	@Tags			text
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	api.GetTextsResponse
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/text [get]
func (h *Handler) GetTexts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	texts, err := h.textsService.GetTexts(r.Context(), userID)
	if err != nil {
		h.logger.Error("Handler GetTexts unhandled error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items := make(api.GetTextsResponse, 0, len(texts))
	for _, text := range texts {
		items = append(items, api.TextItemResponse{
			Name:      text.Name,
			Metadata:  text.Metadata,
			Checksum:  text.Checksum,
			CreatedAt: text.CreatedAt,
			UpdatedAt: text.UpdatedAt,
		})
	}

	writeJSON(w, h.logger, items)
}

// GET TEXT

// GetText godoc
//
//	@Summary		Получение текстовых данных
//	@Description	Возвращает текстовые данные текущего пользователя по имени
//	@Tags			text
//	@Security		BearerAuth
//	@Produce		json
//	@Param			text_name	path		string	true	"Имя текста"
//	@Success		200			{object}	api.TextResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Текст не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/text/{text_name} [get]
func (h *Handler) GetText(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	textName := pathParam(r, "text_name")
	if textName == "" {
		http.Error(w, "text_name path parameter is required", http.StatusBadRequest)
		return
	}

	text, err := h.textsService.GetText(r.Context(), userID, textName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTextNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler GetText unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, textResponse(text))
}

// UPDATE TEXT

// UpdateText godoc
//
//	@Summary		Обновление текстовых данных
//	@Description	Полностью перезаписывает текстовые данные текущего пользователя, найденные по имени
//	@Tags			text
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			text_name	path		string					true	"Имя текста"
//	@Param			body		body		api.UpdateTextRequest	true	"Новые текстовые данные"
//	@Success		200			{object}	api.TextResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Текст не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/text/{text_name} [put]
func (h *Handler) UpdateText(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	textName := pathParam(r, "text_name")
	if textName == "" {
		http.Error(w, "text_name path parameter is required", http.StatusBadRequest)
		return
	}

	var requestData api.UpdateTextRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	text, err := h.textsService.UpdateText(
		r.Context(),
		userID,
		textName,
		requestData.Text,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTextNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler UpdateText unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, textResponse(text))
}

// DELETE TEXT

// DeleteText godoc
//
//	@Summary		Удаление текстовых данных
//	@Description	Удаляет текстовые данные текущего пользователя по имени
//	@Tags			text
//	@Security		BearerAuth
//	@Produce		plain
//	@Param			text_name	path	string	true	"Имя текста"
//	@Success		204
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404	{string}	string	"Текст не найден"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/text/{text_name} [delete]
func (h *Handler) DeleteText(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	textName := pathParam(r, "text_name")
	if textName == "" {
		http.Error(w, "text_name path parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.textsService.DeleteText(r.Context(), userID, textName); err != nil {
		switch {
		case errors.Is(err, service.ErrTextNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler DeleteText unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// textResponse собирает ответ с полными текстовыми данными.
func textResponse(text *entities.Text) api.TextResponse {
	return api.TextResponse{
		Name:      text.Name,
		Text:      text.Text,
		Metadata:  text.Metadata,
		Checksum:  text.Checksum,
		CreatedAt: text.CreatedAt,
		UpdatedAt: text.UpdatedAt,
	}
}
