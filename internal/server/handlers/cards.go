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

// CREATE CARD

// CreateCard godoc
//
//	@Summary		Создание банковской карты
//	@Description	Сохраняет данные банковской карты текущего пользователя. Имя карты уникально в рамках пользователя
//	@Tags			card
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		api.CreateCardRequest	true	"Данные карты"
//	@Success		200		{object}	api.CardResponse
//	@Failure		400		{string}	string	"Неверный формат запроса"
//	@Failure		401		{string}	string	"Пользователь не аутентифицирован"
//	@Failure		409		{string}	string	"Карта с таким именем уже существует"
//	@Failure		500		{string}	string	"Внутренняя ошибка"
//	@Router			/api/card [post]
func (h *Handler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var requestData api.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestData.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	card, err := h.cardsService.CreateCard(
		r.Context(),
		userID,
		requestData.Name,
		requestData.Number,
		requestData.Holder,
		requestData.ExpiresAt,
		requestData.CVV,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCardAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler CreateCard unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, cardResponse(card))
}

// GET CARDS

// GetCards godoc
//
//	@Summary		Список карт пользователя
//	@Description	Возвращает карты текущего пользователя без CVV и срока действия
//	@Tags			card
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	api.GetCardsResponse
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/card [get]
func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	cards, err := h.cardsService.GetCards(r.Context(), userID)
	if err != nil {
		h.logger.Error("Handler GetCards unhandled error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items := make(api.GetCardsResponse, 0, len(cards))
	for _, card := range cards {
		items = append(items, api.CardItemResponse{
			Name:      card.Name,
			Number:    card.Number,
			Holder:    card.Holder,
			Metadata:  card.Metadata,
			Checksum:  card.Checksum,
			CreatedAt: card.CreatedAt,
			UpdatedAt: card.UpdatedAt,
		})
	}

	writeJSON(w, h.logger, items)
}

// GET CARD

// GetCard godoc
//
//	@Summary		Получение карты
//	@Description	Возвращает данные карты текущего пользователя по имени
//	@Tags			card
//	@Security		BearerAuth
//	@Produce		json
//	@Param			card_name	path		string	true	"Имя карты"
//	@Success		200			{object}	api.CardResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Карта не найдена"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/card/{card_name} [get]
func (h *Handler) GetCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	cardName := pathParam(r, "card_name")
	if cardName == "" {
		http.Error(w, "card_name path parameter is required", http.StatusBadRequest)
		return
	}

	card, err := h.cardsService.GetCard(r.Context(), userID, cardName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCardNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler GetCard unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, cardResponse(card))
}

// UPDATE CARD

// UpdateCard godoc
//
//	@Summary		Обновление карты
//	@Description	Полностью перезаписывает данные карты текущего пользователя, найденной по имени
//	@Tags			card
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			card_name	path		string					true	"Имя карты"
//	@Param			body		body		api.UpdateCardRequest	true	"Новые данные карты"
//	@Success		200			{object}	api.CardResponse
//	@Failure		400			{string}	string	"Неверный формат запроса"
//	@Failure		401			{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404			{string}	string	"Карта не найдена"
//	@Failure		500			{string}	string	"Внутренняя ошибка"
//	@Router			/api/card/{card_name} [put]
func (h *Handler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	cardName := pathParam(r, "card_name")
	if cardName == "" {
		http.Error(w, "card_name path parameter is required", http.StatusBadRequest)
		return
	}

	var requestData api.UpdateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	card, err := h.cardsService.UpdateCard(
		r.Context(),
		userID,
		cardName,
		requestData.Number,
		requestData.Holder,
		requestData.ExpiresAt,
		requestData.CVV,
		requestData.Metadata,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCardNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler UpdateCard unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, h.logger, cardResponse(card))
}

// DELETE CARD

// DeleteCard godoc
//
//	@Summary		Удаление карты
//	@Description	Удаляет карту текущего пользователя по имени
//	@Tags			card
//	@Security		BearerAuth
//	@Produce		plain
//	@Param			card_name	path	string	true	"Имя карты"
//	@Success		204
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Пользователь не аутентифицирован"
//	@Failure		404	{string}	string	"Карта не найдена"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/card/{card_name} [delete]
func (h *Handler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	cardName := pathParam(r, "card_name")
	if cardName == "" {
		http.Error(w, "card_name path parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.cardsService.DeleteCard(r.Context(), userID, cardName); err != nil {
		switch {
		case errors.Is(err, service.ErrCardNotExists):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logger.Error("Handler DeleteCard unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// cardResponse собирает ответ с полными данными карты.
func cardResponse(card *entities.Card) api.CardResponse {
	return api.CardResponse{
		Name:      card.Name,
		Number:    card.Number,
		Holder:    card.Holder,
		ExpiresAt: card.ExpiresAt,
		CVV:       card.CVV,
		Metadata:  card.Metadata,
		Checksum:  card.Checksum,
		CreatedAt: card.CreatedAt,
		UpdatedAt: card.UpdatedAt,
	}
}
