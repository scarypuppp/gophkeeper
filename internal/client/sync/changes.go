package sync

import (
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
)

// Changes — сводка расхождений локальных данных с состоянием на момент
// последней синхронизации. По ней видно, что уйдёт на сервер при sync.
type Changes struct {
	Added   int
	Updated int
	Deleted int
}

// Total возвращает общее число изменённых записей.
func (c Changes) Total() int {
	return c.Added + c.Updated + c.Deleted
}

// LocalChanges сравнивает локальные данные с BASE по контрольным суммам:
// запись считается изменённой по содержимому, а не по времени правки.
func LocalChanges(local storage.Data, base storage.Data) Changes {
	var changes Changes
	changes.count(fileChecksums(local.Files), fileChecksums(base.Files))
	changes.count(textChecksums(local.Texts), textChecksums(base.Texts))
	changes.count(cardChecksums(local.Cards), cardChecksums(base.Cards))
	changes.count(secretChecksums(local.Secrets), secretChecksums(base.Secrets))
	return changes
}

// count добавляет расхождения между записями одного типа.
func (c *Changes) count(local map[string]string, base map[string]string) {
	for name, sum := range local {
		baseSum, ok := base[name]
		switch {
		case !ok:
			c.Added++
		case baseSum != sum:
			c.Updated++
		}
	}
	for name := range base {
		if _, ok := local[name]; !ok {
			c.Deleted++
		}
	}
}
