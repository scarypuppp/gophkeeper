package interactor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scarypuppp/gophkeeper/internal/api"
)

// ErrCardNotFound возвращается, если сервер ответил 404 на операцию с картой.
var ErrCardNotFound = errors.New("card not found")

// cardPath строит путь к карте, экранируя имя: оно может содержать '/' и пробелы.
func cardPath(name string) string {
	return fmt.Sprintf("/api/card/%s", url.PathEscape(name))
}

func (i *Interactor) CreateCard(token string, card api.CreateCardRequest) (*api.CardResponse, error) {
	var cardResponse api.CardResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(card).
		SetResult(&cardResponse).
		Post("/api/card")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &cardResponse, nil
}

func (i *Interactor) GetCards(token string) (api.GetCardsResponse, error) {
	var cardsResponse api.GetCardsResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&cardsResponse).
		Get("/api/card")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return cardsResponse, nil
}

func (i *Interactor) GetCard(token string, name string) (*api.CardResponse, error) {
	var cardResponse api.CardResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&cardResponse).
		Get(cardPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrCardNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &cardResponse, nil
}

func (i *Interactor) UpdateCard(token string, name string, card api.UpdateCardRequest) (*api.CardResponse, error) {
	var cardResponse api.CardResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(card).
		SetResult(&cardResponse).
		Put(cardPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrCardNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &cardResponse, nil
}

func (i *Interactor) DeleteCard(token string, name string) error {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		Delete(cardPath(name))
	if err != nil {
		return i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return ErrCardNotFound
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}
