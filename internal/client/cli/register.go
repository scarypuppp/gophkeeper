package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/spf13/cobra"
)

// RegisterCommand регистрирует нового пользователя на сервере.
func RegisterCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "register <login> <password>",
		Short: "Register a new user.",
		Long:  "",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Session != nil {
				fmt.Printf("Already logged in as '%s'. Log out first\n", a.Session.Login)
				return nil
			}
			login, password := args[0], args[1]

			if err := a.Interactor.Register(login, password); err != nil {
				return err
			}
			fmt.Printf("Successfully registered user with login '%s'\n", login)
			return nil
		},
	}

	return c
}
