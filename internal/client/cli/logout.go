package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/spf13/cobra"
)

// LogoutCommand завершает сессию и очищает локальные данные пользователя.
func LogoutCommand(a *app.App) *cobra.Command {
	var force bool

	c := &cobra.Command{
		Use:   "logout",
		Short: "Logout from account.",
		Long: `Логаут очищает файл сессии, data.json и data_base.json, чтобы следующий
пользователь этой машины не увидел чужих данных.

Если в data.json есть несинхронизированные изменения, команда предупреждает
об этом и требует подтверждения.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Session == nil {
				fmt.Println("Not authorized")
				return nil
			}

			changes, err := unsyncedChanges(a)
			if err != nil {
				return err
			}
			if changes.Total() > 0 && !force {
				fmt.Printf("Warning: %d local change(s) were not synced with the server.\n", changes.Total())
				fmt.Println("They will be lost. Run 'gophkeeper sync' first to keep them.")
				confirmed, err := confirm("Log out anyway?")
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Println("Logout cancelled")
					return nil
				}
			}

			if err := a.ResetSession(); err != nil {
				return err
			}
			if err := a.Storage.Clear(); err != nil {
				return err
			}
			if err := a.BaseStorage.Clear(); err != nil {
				return err
			}
			fmt.Println("Successfully logged out")
			return nil
		},
	}

	c.Flags().BoolVarP(&force, "force", "f", false, "Logout without confirmation even if local changes are not synced")

	return c
}
