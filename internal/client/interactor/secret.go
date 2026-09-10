package interactor

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scarypuppp/gophkeeper/internal/api"
)

// ErrSecretNotFound возвращается, если сервер ответил 404 на операцию с секретом.
var ErrSecretNotFound = errors.New("secret not found")

// secretPath строит путь к секрету, экранируя имя: оно может содержать '/' и пробелы.
func secretPath(name string) string {
	return fmt.Sprintf("/api/secret/%s", url.PathEscape(name))
}

func (i *Interactor) CreateSecret(token string, secret api.CreateSecretRequest) (*api.SecretResponse, error) {
	var secretResponse api.SecretResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(secret).
		SetResult(&secretResponse).
		Post("/api/secret")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &secretResponse, nil
}

func (i *Interactor) GetSecrets(token string) (api.GetSecretsResponse, error) {
	var secretsResponse api.GetSecretsResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&secretsResponse).
		Get("/api/secret")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return secretsResponse, nil
}

func (i *Interactor) GetSecret(token string, name string) (*api.SecretResponse, error) {
	var secretResponse api.SecretResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetResult(&secretResponse).
		Get(secretPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrSecretNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &secretResponse, nil
}

func (i *Interactor) UpdateSecret(token string, name string, secret api.UpdateSecretRequest) (*api.SecretResponse, error) {
	var secretResponse api.SecretResponse
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		SetBody(secret).
		SetResult(&secretResponse).
		Put(secretPath(name))
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrSecretNotFound
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &secretResponse, nil
}

func (i *Interactor) DeleteSecret(token string, name string) error {
	resp, err := i.httpClient.R().
		SetAuthToken(token).
		Delete(secretPath(name))
	if err != nil {
		return i.requestError(err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return ErrSecretNotFound
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}
