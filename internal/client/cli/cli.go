package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
)

// ErrNotAuthorized возвращается командами, требующими входа в аккаунт.
var ErrNotAuthorized = errors.New("not authorized, use 'gophkeeper login' first")

// errNotImplemented возвращается заглушками команд, за которыми ещё нет логики.
var errNotImplemented = errors.New("not implemented")

// requireAuth проверяет, что пользователь вошёл в аккаунт.
// Ввод и просмотр данных возможны только после логина.
func requireAuth(a *app.App) error {
	if a.Session == nil || a.Session.AuthToken == "" {
		return ErrNotAuthorized
	}
	return nil
}

// confirm задаёт вопрос и ждёт от пользователя y или n.
// Пустой ответ считается отказом.
func confirm(question string) (bool, error) {
	fmt.Printf("%s [y/N]: ", question)

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read answer: %w", err)
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

// countingWriter считает, сколько байт было написано: по нему видно,
// сообщила ли синхронизация хоть об одном изменении.
type countingWriter struct {
	out     io.Writer
	written int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.out.Write(p)
	w.written += n
	return n, err
}
