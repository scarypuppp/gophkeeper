package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/interactor"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/scarypuppp/gophkeeper/internal/client/sync"
	"github.com/spf13/cobra"
)

// StatusCommand показывает учётную запись, доступность сервера и локальные изменения.
func StatusCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "status",
		Short: "See current gophkeeper status.",
		Long: `Показывает, под какой учётной записью работает клиент, отвечает ли сервер
и сколько локальных изменений ещё не ушло на него.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			table := newTable()
			fmt.Fprintf(table, "account:\t%s\n", accountStatus(a))
			fmt.Fprintf(table, "server:\t%s\n", serverStatus(a))

			changes, err := localChangesStatus(a)
			if err != nil {
				return err
			}
			fmt.Fprintf(table, "local changes:\t%s\n", changes)
			fmt.Fprintf(table, "last sync:\t%s\n", lastSyncStatus(a))
			return table.Flush()
		},
	}

	return c
}

// accountStatus описывает текущую учётную запись.
func accountStatus(a *app.App) string {
	if requireAuth(a) != nil {
		return "not logged in"
	}
	return a.Session.Login
}

// serverStatus проверяет доступность сервера.
func serverStatus(a *app.App) string {
	var token string
	if a.Session != nil {
		token = a.Session.AuthToken
	}

	err := a.Interactor.Ping(token)
	switch {
	case err == nil:
		return fmt.Sprintf("online (%s)", a.Config.BaseURL)
	case errors.Is(err, interactor.ErrServerUnavailable):
		return fmt.Sprintf("offline (%s)", a.Config.BaseURL)
	case errors.Is(err, interactor.ErrUnauthorized):
		// Сервер отвечает, но токен не принят: без входа это норма,
		// а с сессией означает, что она больше не годится.
		if requireAuth(a) != nil {
			return fmt.Sprintf("online (%s)", a.Config.BaseURL)
		}
		return fmt.Sprintf("online (%s), session expired, log in again", a.Config.BaseURL)
	}
	return fmt.Sprintf("unknown (%s): %s", a.Config.BaseURL, err)
}

// localChangesStatus описывает расхождение локальных данных с состоянием
// последней синхронизации.
func localChangesStatus(a *app.App) (string, error) {
	changes, err := unsyncedChanges(a)
	if err != nil {
		return "", err
	}
	if changes.Total() == 0 {
		return "none", nil
	}

	parts := make([]string, 0, 3)
	for _, part := range []struct {
		count int
		name  string
	}{
		{changes.Added, "added"},
		{changes.Updated, "updated"},
		{changes.Deleted, "deleted"},
	} {
		if part.count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", part.count, part.name))
		}
	}
	return fmt.Sprintf("%d (%s), run 'gophkeeper sync'", changes.Total(), strings.Join(parts, ", ")), nil
}

// lastSyncStatus возвращает время последней синхронизации: его показывает
// время изменения data_base.json, который sync переписывает в самом конце.
func lastSyncStatus(a *app.App) string {
	info, err := os.Stat(a.BaseStorage.Path())
	if err != nil {
		return "never"
	}
	return info.ModTime().Local().Format(time.DateTime)
}

// unsyncedChanges считает локальные изменения, не ушедшие на сервер.
// Если синхронизации ещё не было, новыми считаются все локальные записи.
func unsyncedChanges(a *app.App) (sync.Changes, error) {
	exists, err := a.BaseStorage.Exists()
	if err != nil {
		return sync.Changes{}, err
	}
	base := storage.Data{}
	if exists {
		if err := a.BaseStorage.Load(); err != nil {
			return sync.Changes{}, err
		}
		base = a.BaseStorage.Snapshot()
	}
	return sync.LocalChanges(a.Storage.Snapshot(), base), nil
}
