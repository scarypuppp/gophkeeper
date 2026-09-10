package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/interactor"
	"github.com/scarypuppp/gophkeeper/internal/client/sync"
	"github.com/spf13/cobra"
)

// LoginCommand выполняет вход и первую синхронизацию с сервером.
func LoginCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "login <login> <password>",
		Short: "Login to gophkeeper.",
		Long: `Логин сохраняет токен в файле сессии и выполняет первую синхронизацию.

Если сервер недоступен, вход всё равно считается успешным, но выводится
сообщение о том, что синхронизация не выполнена.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Session != nil {
				fmt.Printf("Already logged in as '%s'\n", a.Session.Login)
				return nil
			}
			login, password := args[0], args[1]

			resp, err := a.Interactor.Login(login, password)
			if err != nil {
				if errors.Is(err, interactor.ErrServerUnavailable) {
					return fmt.Errorf("%w, login requires a connection to the server", err)
				}
				return err
			}

			if err = a.SetAuthData(login, resp.AccessToken); err != nil {
				return err
			}
			fmt.Printf("Successfully logged in as '%s'\n", login)

			// Вход состоялся, поэтому недоступный сервер не считается ошибкой:
			// синхронизацию можно повторить командой 'gophkeeper sync'.
			if err := runSync(a, sync.NewPromptResolver(os.Stdin, os.Stdout)); err != nil {
				fmt.Printf("Warning: sync was not performed: %s\n", err)
			}
			return nil
		},
	}

	return c
}
