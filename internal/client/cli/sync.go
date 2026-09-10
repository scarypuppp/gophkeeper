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

// SyncCommand синхронизирует локальные данные с сервером.
func SyncCommand(a *app.App) *cobra.Command {
	var (
		theirs bool
		yours  bool
	)

	c := &cobra.Command{
		Use:   "sync",
		Short: "Sync local data with the server.",
		Long: `Порядок выполнения:

  1. запросить у сервера полное состояние данных;
  2. локально вычислить решения по таблице (local / base / remote);
  3. разрешить все конфликты;
  4. отправить изменения на сервер и скачать недостающие данные;
  5. перезаписать data_base.json.

Изменения уходят на сервер только после того, как не осталось нерешённых
конфликтов. Без флагов клиент спрашивает по каждой конфликтной записи
отдельно, показывая версии side by side; выбор — 's' (server) или 'l' (local).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			var resolver sync.Resolver = sync.NewPromptResolver(os.Stdin, os.Stdout)
			switch {
			case theirs:
				resolver = sync.TakeRemote()
			case yours:
				resolver = sync.TakeLocal()
			}

			if err := runSync(a, resolver); err != nil {
				if errors.Is(err, interactor.ErrServerUnavailable) {
					return fmt.Errorf("%w, local data is unchanged", err)
				}
				return err
			}
			return nil
		},
	}

	c.Flags().BoolVar(&theirs, "theirs", false, "Resolve every conflict in favour of the server version")
	c.Flags().BoolVar(&yours, "yours", false, "Resolve every conflict in favour of the local version")
	c.MarkFlagsMutuallyExclusive("theirs", "yours")

	return c
}

// runSync выполняет синхронизацию и сообщает, если менять было нечего.
func runSync(a *app.App, resolver sync.Resolver) error {
	out := &countingWriter{out: os.Stdout}
	syncer := sync.NewSyncer(
		a.Storage,
		a.BaseStorage,
		a.Interactor,
		a.Session.AuthToken,
		a.Config.DownloadedFilesPath,
		resolver,
		out,
	)

	if err := syncer.Run(); err != nil {
		return err
	}
	if out.written == 0 {
		fmt.Println("Already up to date")
	}
	return nil
}
