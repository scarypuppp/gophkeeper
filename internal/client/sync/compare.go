// Package sync приводит локальные данные и данные сервера к общему состоянию.
//
// Решение по каждой записи принимается сравнением трёх версий: локальной,
// серверной и BASE — состояния на момент последней синхронизации (data_base.json).
// Записи сопоставляются по имени, изменение определяется по контрольной сумме.
package sync

// Action — что нужно сделать с записью, чтобы стороны сошлись.
type Action int

const (
	// ActionNone — стороны уже согласованы.
	ActionNone Action = iota
	// ActionPush — отправить локальную версию на сервер.
	ActionPush
	// ActionPull — взять версию сервера себе.
	ActionPull
	// ActionDeleteRemote — запись удалена локально, удалить её на сервере.
	ActionDeleteRemote
	// ActionDeleteLocal — запись удалена на сервере, удалить её локально.
	ActionDeleteLocal
	// ActionConflict — изменены обе стороны, нужен выбор пользователя.
	ActionConflict
)

// Side — выбранная при разрешении конфликта версия записи.
type Side int

const (
	// SideLocal — оставить локальную версию (yours).
	SideLocal Side = iota
	// SideRemote — взять версию сервера (theirs).
	SideRemote
)

// State — присутствие записи в одной из трёх версий и контрольная сумма
// её содержимого.
type State struct {
	Present  bool
	Checksum string
}

// Decide возвращает решение по одной записи.
//
// baseKnown говорит, существовал ли файл data_base.json: если нет, то все записи
// считаются новыми, и при совпадении имён серверная версия побеждает без запроса.
func Decide(local State, base State, remote State, baseKnown bool) Action {
	localChanged := changed(local, base)
	remoteChanged := changed(remote, base)

	switch {
	case !localChanged && !remoteChanged:
		return ActionNone

	case localChanged && !remoteChanged:
		if !local.Present {
			return ActionDeleteRemote
		}
		return ActionPush

	case !localChanged && remoteChanged:
		if !remote.Present {
			return ActionDeleteLocal
		}
		return ActionPull
	}

	// Изменились обе стороны.
	switch {
	case !local.Present && !remote.Present:
		// Запись удалили и там, и там — расхождения нет.
		return ActionNone
	case local.Present && remote.Present && local.Checksum == remote.Checksum:
		// Обе стороны пришли к одинаковому содержимому.
		return ActionNone
	case !baseKnown && !base.Present && local.Present && remote.Present:
		// Первая синхронизация: обе записи новые, серверная побеждает молча.
		return ActionPull
	}
	return ActionConflict
}

// Resolve превращает выбранную пользователем сторону в конкретное действие.
func Resolve(side Side, local State, remote State) Action {
	if side == SideRemote {
		if remote.Present {
			return ActionPull
		}
		return ActionDeleteLocal
	}
	if local.Present {
		return ActionPush
	}
	return ActionDeleteRemote
}

// changed сообщает, отличается ли версия от BASE.
func changed(side State, base State) bool {
	if side.Present != base.Present {
		return true
	}
	if !side.Present {
		return false
	}
	return side.Checksum != base.Checksum
}
