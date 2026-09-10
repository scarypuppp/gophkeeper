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

// CREATE SECRET

// CreateSecret godoc
//
//	@Summary		Создание секрета
//	@Description	Создаёт секрет текущего пользователя. Имя секрета уникально в рамках пользователя
//	@Tags			secret
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.CreateSecretRequest	true	"Данные секрета"
//	@Success		200		{object}	api.SecretResponse
//	@Failure		400		{string}	string	"Неверный формат запроса"
//	@Failure		401		{string}	string	"Пользователь не аутентифицирован"
//	@Failure		409		{string}	string	"Секрет с таким именем уже существует"
//	@Failure		500		{string}	string	"Внутренняя ошибка"
//	@Router			/api/secret [post]
func (h *Handler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var requestData api.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestData.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	secret, err := h.secretsService.CreateSecret(
		r.Context(),
		userID,
		requestData.Name,
		requestData.Login,
		requestData.Password,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSecretAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler CreateSecret unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, secretResponse(secret))
}

// GET SECRETS

// GetSecrets godoc
//
//	@Summary		Список секретов пользователя
//	@Description	Возвращает секреты текущего пользователя без паролей
//	@Tags			secret
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	api.GetSecretsResponse
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/secret [get]
func (h *Handler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secrets, err := h.secretsService.GetSecrets(r.Context(), userID)
	if err != nil {
		h.logger.Error("Handler GetSecrets unhandled error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items := make(api.GetSecretsResponse, 0, len(secrets))
	for _, secret := range secrets {
		items = append(items, api.SecretItemResponse{
			Name:      secret.Name,
			Login:     secret.Login,
			Metadata:  secret.Metadata,
			Checksum:  secret.Checksum,
			CreatedAt: secret.CreatedAt,
			UpdatedAt: secret.UpdatedAt,
		})
	}

	writeJSON(w, h.logger, items)
}

// GET SECRET

// GetSecret godoc
//
//	@Summary		Получение секрета
//	@Description	Возвращает секрет текущего пользователя по имени
//	@Tags			secret
//	@Security		BearerAuth
//	@Produce		json
//	@Param			secret_name	path		string	true	"Имя секрета"
//	@Success		200			{object}	api.SecretResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Секрет не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/secret/{secret_name} [get]
func (h *Handler) GetSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secretName := pathParam(r, "secret_name")
	if secretName == "" {
		http.Error(w, "secret_name path parameter is required", http.StatusBadRequest)
		return
	}

	secret, err := h.secretsService.GetSecret(r.Context(), userID, secretName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSecretNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler GetSecret unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, secretResponse(secret))
}

// UPDATE SECRET

// UpdateSecret godoc
//
//	@Summary		Обновление секрета
//	@Description	Полностью перезаписывает содержимое секрета текущего пользователя, найденного по имени
//	@Tags			secret
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			secret_name	path		string					true	"Имя секрета"
//	@Param			body		body		api.UpdateSecretRequest	true	"Новые данные секрета"
//	@Success		200			{object}	api.SecretResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Секрет не найден"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/secret/{secret_name} [put]
func (h *Handler) UpdateSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secretName := pathParam(r, "secret_name")
	if secretName == "" {
		http.Error(w, "secret_name path parameter is required", http.StatusBadRequest)
		return
	}

	var requestData api.UpdateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secret, err := h.secretsService.UpdateSecret(
		r.Context(),
		userID,
		secretName,
		requestData.Login,
		requestData.Password,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSecretNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler UpdateSecret unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, secretResponse(secret))
}

// DELETE SECRET

// DeleteSecret godoc
//
//	@Summary		Удаление секрета
//	@Description	Удаляет секрет текущего пользователя по имени
//	@Tags			secret
//	@Security		BearerAuth
//	@Produce		plain
//	@Param			secret_name	path	string	true	"Имя секрета"
//	@Success		204
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404	{string}	string	"Секрет не найден"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/secret/{secret_name} [delete]
func (h *Handler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	secretName := pathParam(r, "secret_name")
	if secretName == "" {
		http.Error(w, "secret_name path parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.secretsService.DeleteSecret(r.Context(), userID, secretName); err != nil {
		switch {
		case errors.Is(err, service.ErrSecretNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler DeleteSecret unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// secretResponse собирает ответ с полными данными секрета.
func secretResponse(secret *entities.Secret) api.SecretResponse {
	return api.SecretResponse{
		Name:      secret.Name,
		Login:     secret.Login,
		Password:  secret.Password,
		Metadata:  secret.Metadata,
		Checksum:  secret.Checksum,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}
}
