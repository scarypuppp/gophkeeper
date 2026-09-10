package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

// newTestStorage создаёт загруженное хранилище над файлом во временном каталоге
// с фиксированными часами.
func newTestStorage(t *testing.T) *LocalStorage {
	t.Helper()
	s := NewLocalStorage(filepath.Join(t.TempDir(), "data.json"))
	s.now = func() time.Time { return testNow }
	require.NoError(t, s.Load())
	return s
}

func TestLoadMissingFile(t *testing.T) {
	s := NewLocalStorage(filepath.Join(t.TempDir(), "nested", "data.json"))

	exists, err := s.Exists()
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, s.Load())
	assert.Empty(t, s.Secrets())
	assert.Empty(t, s.Files())
}

func TestLoadBrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0600))

	require.ErrorContains(t, NewLocalStorage(path).Load(), "parse")
}

func TestSaveCreatesDirAndKeepsFilePrivate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "data.json")
	s := NewLocalStorage(path)
	require.NoError(t, s.Load())

	require.NoError(t, s.Save())

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, storageFileMode, info.Mode().Perm())

	exists, err := s.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRoundTrip(t *testing.T) {
	s := newTestStorage(t)

	_, err := s.AddSecret(Secret{Name: "github", Login: "alice", Password: "qwerty", Metadata: "work"})
	require.NoError(t, err)
	_, err = s.AddCard(Card{Name: "visa", Number: "4111111111111111", Holder: "ALICE", ExpiresAt: "12/29", CVV: "123"})
	require.NoError(t, err)
	_, err = s.AddText(Text{Name: "note", Text: "hello", Metadata: "m"})
	require.NoError(t, err)
	_, err = s.AddFile(File{FileName: "report.pdf", Size: 42, Checksum: "abc", Metadata: "m"})
	require.NoError(t, err)

	loaded := NewLocalStorage(s.Path())
	require.NoError(t, loaded.Load())
	assert.Equal(t, s.Snapshot(), loaded.Snapshot())
}

// TestFileFormat проверяет, что имена полей в data.json совпадают с ТЗ.
func TestFileFormat(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice", Password: "qwerty", Metadata: "work"})
	require.NoError(t, err)
	_, err = s.AddCard(Card{Name: "visa", Number: "4111", Holder: "ALICE", ExpiresAt: "12/29", CVV: "123"})
	require.NoError(t, err)

	raw, err := os.ReadFile(s.Path())
	require.NoError(t, err)

	var file map[string][]map[string]any
	require.NoError(t, json.Unmarshal(raw, &file))

	assert.ElementsMatch(t, []string{"Files", "Texts", "Cards", "Secrets"}, keys(file))
	assert.ElementsMatch(t,
		[]string{"name", "login", "password", "meta", "checksum", "updated_at"},
		keys(file["Secrets"][0]))
	assert.ElementsMatch(t,
		[]string{"name", "number", "holder", "expires", "cvv", "meta", "checksum", "updated_at"},
		keys(file["Cards"][0]))
	// Пустые срезы сохраняются как [], а не как null.
	assert.NotNil(t, file["Files"])
	assert.Empty(t, file["Files"])
}

func TestAddSecret(t *testing.T) {
	s := newTestStorage(t)

	added, err := s.AddSecret(Secret{Name: "github", Login: "alice", Password: "qwerty", Metadata: "work"})
	require.NoError(t, err)
	assert.Equal(t, checksum.Secret("alice", "qwerty", "work"), added.Checksum)
	assert.Equal(t, testNow, added.UpdatedAt)

	_, err = s.AddSecret(Secret{Name: "github", Login: "bob"})
	assert.ErrorIs(t, err, ErrAlreadyExists)

	_, err = s.AddSecret(Secret{Login: "bob"})
	assert.ErrorIs(t, err, ErrEmptyName)

	// Неудачные добавления не попали ни в память, ни в файл.
	assert.Len(t, s.Secrets(), 1)
	reloaded := NewLocalStorage(s.Path())
	require.NoError(t, reloaded.Load())
	assert.Len(t, reloaded.Secrets(), 1)
}

func TestGetSecret(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice"})
	require.NoError(t, err)

	got, err := s.GetSecret("github")
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Login)

	_, err = s.GetSecret("gitlab")
	assert.ErrorIs(t, err, ErrNotFound)
}

// TestUpdateSecretPartial проверяет, что меняются только переданные поля.
func TestUpdateSecretPartial(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice", Password: "qwerty", Metadata: "work"})
	require.NoError(t, err)

	later := testNow.Add(time.Hour)
	s.now = func() time.Time { return later }

	password := "123456"
	updated, err := s.UpdateSecret("github", SecretPatch{Password: &password})
	require.NoError(t, err)

	assert.Equal(t, "alice", updated.Login)
	assert.Equal(t, "123456", updated.Password)
	assert.Equal(t, "work", updated.Metadata)
	assert.Equal(t, checksum.Secret("alice", "123456", "work"), updated.Checksum)
	assert.Equal(t, later, updated.UpdatedAt)

	reloaded := NewLocalStorage(s.Path())
	require.NoError(t, reloaded.Load())
	stored, err := reloaded.GetSecret("github")
	require.NoError(t, err)
	assert.Equal(t, updated, stored)

	_, err = s.UpdateSecret("gitlab", SecretPatch{Password: &password})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteSecret(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice"})
	require.NoError(t, err)
	_, err = s.AddSecret(Secret{Name: "gitlab", Login: "bob"})
	require.NoError(t, err)

	require.NoError(t, s.DeleteSecret("github"))
	assert.Len(t, s.Secrets(), 1)

	reloaded := NewLocalStorage(s.Path())
	require.NoError(t, reloaded.Load())
	assert.Len(t, reloaded.Secrets(), 1)

	assert.ErrorIs(t, s.DeleteSecret("github"), ErrNotFound)
}

func TestCardCRUD(t *testing.T) {
	s := newTestStorage(t)

	added, err := s.AddCard(Card{Name: "visa", Number: "4111111111111111", Holder: "ALICE", ExpiresAt: "12/29", CVV: "123", Metadata: "m"})
	require.NoError(t, err)
	assert.Equal(t, checksum.Card("4111111111111111", "ALICE", "12/29", "123", "m"), added.Checksum)

	holder := "BOB"
	updated, err := s.UpdateCard("visa", CardPatch{Holder: &holder})
	require.NoError(t, err)
	assert.Equal(t, "BOB", updated.Holder)
	assert.Equal(t, "4111111111111111", updated.Number)
	assert.Equal(t, checksum.Card("4111111111111111", "BOB", "12/29", "123", "m"), updated.Checksum)

	require.NoError(t, s.DeleteCard("visa"))
	_, err = s.GetCard("visa")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestTextCRUD(t *testing.T) {
	s := newTestStorage(t)

	added, err := s.AddText(Text{Name: "note", Text: "hello", Metadata: "m"})
	require.NoError(t, err)
	assert.Equal(t, checksum.Text("hello", "m"), added.Checksum)

	text := "bye"
	updated, err := s.UpdateText("note", TextPatch{Text: &text})
	require.NoError(t, err)
	assert.Equal(t, "bye", updated.Text)
	assert.Equal(t, "m", updated.Metadata)
	assert.Equal(t, checksum.Text("bye", "m"), updated.Checksum)

	require.NoError(t, s.DeleteText("note"))
	_, err = s.GetText("note")
	assert.ErrorIs(t, err, ErrNotFound)
}

// TestFileCRUD проверяет, что размер и контрольную сумму содержимого
// хранилище берёт у вызывающего и само не считает.
func TestFileCRUD(t *testing.T) {
	s := newTestStorage(t)

	added, err := s.AddFile(File{FileName: "report.pdf", Size: 42, Checksum: "abc", Metadata: "m"})
	require.NoError(t, err)
	assert.Equal(t, int64(42), added.Size)
	assert.Equal(t, "abc", added.Checksum)
	assert.Equal(t, testNow, added.UpdatedAt)
	assert.Empty(t, added.FileHash)

	size, sum, hash := int64(100), "def", "server-hash"
	updated, err := s.UpdateFile("report.pdf", FilePatch{Size: &size, Checksum: &sum, FileHash: &hash})
	require.NoError(t, err)
	assert.Equal(t, int64(100), updated.Size)
	assert.Equal(t, "def", updated.Checksum)
	assert.Equal(t, "server-hash", updated.FileHash)
	assert.Equal(t, "m", updated.Metadata)

	require.NoError(t, s.DeleteFile("report.pdf"))
	_, err = s.GetFile("report.pdf")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestListsSortedByName(t *testing.T) {
	s := newTestStorage(t)
	for _, name := range []string{"gitlab", "aws", "github"} {
		_, err := s.AddSecret(Secret{Name: name})
		require.NoError(t, err)
	}

	names := make([]string, 0, 3)
	for _, secret := range s.Secrets() {
		names = append(names, secret.Name)
	}
	assert.Equal(t, []string{"aws", "github", "gitlab"}, names)
}

// TestSnapshotIsCopy проверяет, что через снимок нельзя изменить хранилище.
func TestSnapshotIsCopy(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice"})
	require.NoError(t, err)

	snapshot := s.Snapshot()
	snapshot.Secrets[0].Login = "mallory"

	got, err := s.GetSecret("github")
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Login)
}

// TestReplaceKeepsValuesAsIs проверяет, что данные сервера кладутся дословно:
// синхронизация не должна пересчитывать суммы и время изменения.
func TestReplaceKeepsValuesAsIs(t *testing.T) {
	s := newTestStorage(t)
	remote := Data{Secrets: []Secret{{
		Name:      "github",
		Login:     "alice",
		Password:  "qwerty",
		Checksum:  "checksum-from-server",
		UpdatedAt: testNow.Add(-time.Hour),
	}}}

	require.NoError(t, s.Replace(remote))

	got, err := s.GetSecret("github")
	require.NoError(t, err)
	assert.Equal(t, "checksum-from-server", got.Checksum)
	assert.Equal(t, testNow.Add(-time.Hour), got.UpdatedAt)

	reloaded := NewLocalStorage(s.Path())
	require.NoError(t, reloaded.Load())
	assert.Equal(t, s.Snapshot(), reloaded.Snapshot())
}

func TestClear(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.AddSecret(Secret{Name: "github", Login: "alice"})
	require.NoError(t, err)

	require.NoError(t, s.Clear())
	assert.Empty(t, s.Secrets())

	// Файл остаётся на месте, но пустой.
	exists, err := s.Exists()
	require.NoError(t, err)
	assert.True(t, exists)

	reloaded := NewLocalStorage(s.Path())
	require.NoError(t, reloaded.Load())
	assert.Equal(t, Data{Files: []File{}, Texts: []Text{}, Cards: []Card{}, Secrets: []Secret{}}, reloaded.Snapshot())
}

func keys[T any](m map[string]T) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}

// TestMutationBeforeLoad проверяет, что незагруженное хранилище не затирает
// файл с данными.
func TestMutationBeforeLoad(t *testing.T) {
	loaded := newTestStorage(t)
	_, err := loaded.AddSecret(Secret{Name: "github", Login: "alice"})
	require.NoError(t, err)

	fresh := NewLocalStorage(loaded.Path())
	_, err = fresh.AddSecret(Secret{Name: "gitlab", Login: "bob"})
	assert.ErrorIs(t, err, ErrNotLoaded)
	assert.ErrorIs(t, fresh.Save(), ErrNotLoaded)

	require.NoError(t, fresh.Load())
	assert.Len(t, fresh.Secrets(), 1)
}

// TestReplaceAndClearWithoutLoad проверяет, что перезапись целиком не требует
// предварительного чтения файла.
func TestReplaceAndClearWithoutLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data_base.json")

	require.NoError(t, NewLocalStorage(path).Replace(Data{Texts: []Text{{Name: "note", Text: "hi"}}}))
	reloaded := NewLocalStorage(path)
	require.NoError(t, reloaded.Load())
	assert.Len(t, reloaded.Texts(), 1)

	require.NoError(t, NewLocalStorage(path).Clear())
	require.NoError(t, reloaded.Load())
	assert.Empty(t, reloaded.Texts())
}
