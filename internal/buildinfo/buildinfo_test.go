package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOrUnknown проверяет, что бинарник, собранный без ldflags, честно
// сообщает об отсутствии сведений о сборке.
func TestOrUnknown(t *testing.T) {
	assert.Equal(t, unknown, orUnknown(""))
	assert.Equal(t, "v1.2.3", orUnknown("v1.2.3"))
}

func TestValuesFromLinker(t *testing.T) {
	prevVersion, prevDate, prevCommit := version, date, commit
	t.Cleanup(func() { version, date, commit = prevVersion, prevDate, prevCommit })

	version, date, commit = "v1.2.3", "2026-09-08T12:00:00Z", "abc1234"
	assert.Equal(t, "v1.2.3", Version())
	assert.Equal(t, "2026-09-08T12:00:00Z", Date())
	assert.Equal(t, "abc1234", Commit())
}
