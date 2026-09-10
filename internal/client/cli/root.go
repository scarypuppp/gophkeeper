package cli

import (
	"github.com/scarypuppp/gophkeeper/internal/buildinfo"
	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/spf13/cobra"
)

// RootCommand собирает дерево команд клиента gophkeeper.
func RootCommand(app *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Client of the gophkeeper password manager.",
		Long: `gophkeeper хранит пароли, банковские карты, произвольные тексты и файлы.

Данные редактируются локально и попадают на сервер во время синхронизации
(gophkeeper sync). Работа с данными возможна только после логина.`,

		// Версию сборки показывают и 'gophkeeper version', и флаг --version.
		Version: buildinfo.Version(),

		// Ошибку печатает main, справку по ошибке команды не показываем.
		SilenceErrors: true,
		SilenceUsage:  true,

		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	c.AddCommand(VersionCommand())
	c.AddCommand(RegisterCommand(app))
	c.AddCommand(LoginCommand(app))
	c.AddCommand(LogoutCommand(app))
	c.AddCommand(StatusCommand(app))
	c.AddCommand(SyncCommand(app))
	c.AddCommand(FileCommand(app))
	c.AddCommand(TextCommand(app))
	c.AddCommand(SecretCommand(app))
	c.AddCommand(CardCommand(app))

	return c
}
