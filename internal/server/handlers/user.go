package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/scarypuppp/gophkeeper/internal/api"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/auth"
	"github.com/scarypuppp/gophkeeper/internal/server/service"
	"go.uber.org/zap"
)

//
// REGISTER
//

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.RegisterRequest	true	"Данные пользователя"
//	@Success		200	{object}	api.LoginResponse
//	@Failure		400	{string}	string	"Неверный формат запроса или недопустимая длина логина либо пароля"
//	@Failure		409	{string}	string	"Логин уже занят"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData api.RegisterRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.userService.RegisterUser(r.Context(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, entities.ErrIncorrectLoginLength),
			errors.Is(err, entities.ErrIncorrectPasswordLength):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrLoginAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler Register unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	token, err := auth.CreateToken(h.config.SecretKey, user.ID, h.config.TokenExpSeconds)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LOGIN

// Login godoc
//
//	@Summary		Аутентификация пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.LoginRequest	true	"Данные пользователя"
//	@Success		200	{object}	api.LoginResponse
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Неверный логин или пароль"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData api.LoginRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.userService.LoginUser(r.Context(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginPasswordNotExist):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			h.logger.Error("Handler Login unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	token, err := auth.CreateToken(h.config.SecretKey, user.ID, h.config.TokenExpSeconds)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	writeJSON(w, h.logger, api.LoginResponse{UserID: user.ID, AccessToken: token})
}
