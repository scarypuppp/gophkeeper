package interactor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// ErrServerUnavailable возвращается, если до сервера не удалось достучаться:
// он не запущен, недоступен по сети или не ответил вовремя.
var ErrServerUnavailable = errors.New("server is unavailable")

type Interactor struct {
	httpClient *resty.Client
}

func NewInteractor(client *resty.Client) *Interactor {
	return &Interactor{client}
}

// requestError разбирает ошибку запроса: до сервера либо не достучались,
// либо запрос не удалось выполнить по другой причине.
// Неуспешные ответы сервера сюда не попадают — их разбирает serverError.
func (i *Interactor) requestError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return fmt.Errorf("%w at %s", ErrServerUnavailable, i.httpClient.BaseURL)
	}
	return err
}

// serverError формирует ошибку по неуспешному ответу сервера.
func serverError(resp *resty.Response) error {
	return fmt.Errorf("server returned %d: %s", resp.StatusCode(), resp.String())
}

// pingTimeout ограничивает проверку доступности сервера: команда status
// не должна ждать полный таймаут клиента, чтобы сказать "offline".
const pingTimeout = 3 * time.Second

// ErrUnauthorized возвращается, если сервер не принял токен сессии.
var ErrUnauthorized = errors.New("session is not accepted by the server")

// Ping проверяет, доступен ли сервер и принимает ли он токен.
// Возвращает ErrServerUnavailable, если до сервера не достучаться,
// и ErrUnauthorized, если сервер ответил, но токен не принял.
func (i *Interactor) Ping(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	resp, err := i.httpClient.R().
		SetContext(ctx).
		SetAuthToken(token).
		Get("/api/file")
	if err != nil {
		return i.requestError(err)
	}
	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.IsError() {
		return serverError(resp)
	}
	return nil
}
