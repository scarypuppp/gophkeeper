package interactor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scarypuppp/gophkeeper/internal/api"
)

// ErrTextNotFound возвращается, если сервер ответил 404 на операцию с текстом.
var ErrTextNotFound = errors.New("text not found")

// textPath строит путь к тексту, экранируя имя: оно может содержать '/' и пробелы.
func textPath(name string) string {
	return fmt.Sprintf("/api/text/%s", url.PathEscape(name))
}

func (i *Interactor) CreateText(token string, text api.CreateTextRequest) (*api.TextResponse, error) {
	var textResponse api.TextResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(text).
		SetResult(&textResponse).
		Post("/api/text")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &textResponse, nil
}

func (i *Interactor) GetTexts(token string) (api.GetTextsResponse, error) {
	var textsResponse api.GetTextsResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&textsResponse).
		Get("/api/text")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return textsResponse, nil
}

func (i *Interactor) GetText(token string, name string) (*api.TextResponse, error) {
	var textResponse api.TextResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&textResponse).
		Get(textPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrTextNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &textResponse, nil
}

func (i *Interactor) UpdateText(token string, name string, text api.UpdateTextRequest) (*api.TextResponse, error) {
	var textResponse api.TextResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(text).
		SetResult(&textResponse).
		Put(textPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrTextNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &textResponse, nil
}

func (i *Interactor) DeleteText(token string, name string) error {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		Delete(textPath(name))
	if err != nil {
		return i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return ErrTextNotFound
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}
