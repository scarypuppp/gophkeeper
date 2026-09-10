package sync

import (
	"bytes"
	"strings"
	"testing"

	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanKind(t *testing.T) {
	local := map[string]string{"same": "a", "changed": "b", "new-local": "x"}
	base := map[string]string{"same": "a", "changed": "a", "deleted-remote": "a"}
	remote := map[string]string{"same": "a", "changed": "a", "deleted-remote": "a", "new-remote": "y"}

	steps := planKind(kindSecret, local, base, remote, true)

	actions := make(map[string]Action, len(steps))
	for _, st := range steps {
		assert.Equal(t, kindSecret, st.kind)
		actions[st.name] = st.action
	}

	// Согласованные записи в план не попадают.
	assert.NotContains(t, actions, "same")
	assert.Equal(t, map[string]Action{
		"changed":        ActionPush,
		"new-local":      ActionPush,
		"new-remote":     ActionPull,
		"deleted-remote": ActionDeleteRemote,
	}, actions)
}

func TestPromptResolver(t *testing.T) {
	conflict := Conflict{
		Kind:   kindSecret,
		Name:   "github",
		Local:  []Field{{Name: "login", Value: "alice"}, {Name: "password", Value: "local"}},
		Remote: []Field{{Name: "login", Value: "alice"}, {Name: "password", Value: "remote"}},
	}

	var out bytes.Buffer
	// Неверный ответ переспрашивается.
	side, err := NewPromptResolver(strings.NewReader("x\ns\n"), &out).Resolve(conflict)
	require.NoError(t, err)
	assert.Equal(t, SideRemote, side)

	printed := out.String()
	assert.Contains(t, printed, "local (l)")
	assert.Contains(t, printed, "server (s)")
	assert.Contains(t, printed, "remote")
	assert.Contains(t, printed, "Please answer 's' or 'l'.")

	side, err = NewPromptResolver(strings.NewReader("L\n"), &bytes.Buffer{}).Resolve(conflict)
	require.NoError(t, err)
	assert.Equal(t, SideLocal, side)
}

// TestPromptResolverWithoutAnswer проверяет, что при закрытом вводе конфликт
// не считается разрешённым молча.
func TestPromptResolverWithoutAnswer(t *testing.T) {
	_, err := NewPromptResolver(strings.NewReader(""), &bytes.Buffer{}).Resolve(Conflict{Kind: kindText, Name: "note"})
	assert.ErrorIs(t, err, ErrConflictNotResolved)
}

// TestConflictOfDeletedRecord проверяет, что удалённая сторона показывается
// прочерком, а не пустыми полями.
func TestConflictOfDeletedRecord(t *testing.T) {
	var out bytes.Buffer
	_, err := NewPromptResolver(strings.NewReader("s\n"), &out).Resolve(Conflict{
		Kind:   kindText,
		Name:   "note",
		Remote: []Field{{Name: "text", Value: "server text"}},
	})
	require.NoError(t, err)
	assert.Contains(t, out.String(), "(deleted)")
}

func TestFixedResolver(t *testing.T) {
	side, err := TakeRemote().Resolve(Conflict{})
	require.NoError(t, err)
	assert.Equal(t, SideRemote, side)

	side, err = TakeLocal().Resolve(Conflict{})
	require.NoError(t, err)
	assert.Equal(t, SideLocal, side)
}

func TestLocalChanges(t *testing.T) {
	base := storage.Data{
		Secrets: []storage.Secret{{Name: "same", Checksum: "a"}, {Name: "changed", Checksum: "a"}, {Name: "gone", Checksum: "a"}},
		Texts:   []storage.Text{{Name: "note", Checksum: "a"}},
	}
	local := storage.Data{
		Secrets: []storage.Secret{{Name: "same", Checksum: "a"}, {Name: "changed", Checksum: "b"}, {Name: "new", Checksum: "c"}},
		Texts:   []storage.Text{{Name: "note", Checksum: "a"}},
	}

	assert.Equal(t, Changes{Added: 1, Updated: 1, Deleted: 1}, LocalChanges(local, base))
	assert.Equal(t, 3, LocalChanges(local, base).Total())
	assert.Equal(t, 0, LocalChanges(base, base).Total())

	// Без BASE все локальные записи новые.
	assert.Equal(t, Changes{Added: 4}, LocalChanges(local, storage.Data{}))
}

// TestLocalChangesFileMetadata проверяет, что правка метаданных файла считается
// изменением: в контрольную сумму содержимого метаданные не входят.
func TestLocalChangesFileMetadata(t *testing.T) {
	base := storage.Data{Files: []storage.File{{FileName: "report.pdf", Checksum: "a", Metadata: "q3"}}}
	local := storage.Data{Files: []storage.File{{FileName: "report.pdf", Checksum: "a", Metadata: "q4"}}}

	assert.Equal(t, Changes{Updated: 1}, LocalChanges(local, base))
}
