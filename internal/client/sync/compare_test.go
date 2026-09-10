package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// present возвращает состояние существующей записи с указанной суммой.
func present(sum string) State {
	return State{Present: true, Checksum: sum}
}

// absent возвращает состояние отсутствующей записи.
func absent() State {
	return State{}
}

func TestDecide(t *testing.T) {
	tests := []struct {
		name   string
		local  State
		base   State
		remote State
		want   Action
	}{
		{"ничего не изменилось", present("a"), present("a"), present("a"), ActionNone},
		{"изменено локально", present("b"), present("a"), present("a"), ActionPush},
		{"изменено на сервере", present("a"), present("a"), present("b"), ActionPull},
		{"изменено с двух сторон", present("b"), present("a"), present("c"), ActionConflict},
		{"изменено одинаково с двух сторон", present("b"), present("a"), present("b"), ActionNone},

		{"новая локальная запись", present("a"), absent(), absent(), ActionPush},
		{"новая запись на сервере", absent(), absent(), present("a"), ActionPull},

		{"удалено локально", absent(), present("a"), present("a"), ActionDeleteRemote},
		{"удалено на сервере", present("a"), present("a"), absent(), ActionDeleteLocal},
		{"удалено с двух сторон", absent(), present("a"), absent(), ActionNone},

		{"удалено локально, изменено на сервере", absent(), present("a"), present("b"), ActionConflict},
		{"изменено локально, удалено на сервере", present("b"), present("a"), absent(), ActionConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Decide(tt.local, tt.base, tt.remote, true))
		})
	}
}

// TestDecideWithoutBase проверяет правило первой синхронизации: при совпадении
// имён серверная версия побеждает без запроса.
func TestDecideWithoutBase(t *testing.T) {
	assert.Equal(t, ActionPull, Decide(present("a"), absent(), present("b"), false))
	assert.Equal(t, ActionNone, Decide(present("a"), absent(), present("a"), false))
	assert.Equal(t, ActionPush, Decide(present("a"), absent(), absent(), false))
	assert.Equal(t, ActionPull, Decide(absent(), absent(), present("a"), false))

	// Запись, известная BASE, разрешается по общим правилам.
	assert.Equal(t, ActionConflict, Decide(present("b"), present("a"), present("c"), false))
}

func TestResolve(t *testing.T) {
	assert.Equal(t, ActionPull, Resolve(SideRemote, present("a"), present("b")))
	assert.Equal(t, ActionDeleteLocal, Resolve(SideRemote, present("a"), absent()))
	assert.Equal(t, ActionPush, Resolve(SideLocal, present("a"), present("b")))
	assert.Equal(t, ActionDeleteRemote, Resolve(SideLocal, absent(), present("b")))
}
