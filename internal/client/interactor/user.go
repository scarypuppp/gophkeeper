package interactor

import (
	"github.com/scarypuppp/gophkeeper/internal/api"
)

func (i *Interactor) Register(login string, password string) error {
	resp, err := i.httpClient.R().
		SetBody(api.RegisterRequest{Login: login, Password: password}).
		Post("/api/user/register")
	if err != nil {
		return i.requestError(err)
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}

func (i *Interactor) Login(login string, password string) (*api.LoginResponse, error) {
	var loginResponse api.LoginResponse
	resp, err := i.httpClient.R().
		SetBody(api.LoginRequest{Login: login, Password: password}).
		SetResult(&loginResponse).
		Post("/api/user/login")
	if err != nil {
		return nil, i.requestError(err)
	}
	if resp.IsError() {
		return nil, serverError(resp)
	}
	return &loginResponse, nil
}
