package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAgentConfig(t *testing.T) {
	t.Setenv(serverURLEnv, "http://example.com:8080")

	cfg, err := GetAgentConfig()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com:8080", cfg.BaseURL)
	assert.Equal(t, "./data.json", cfg.DataFilePath)
	assert.Equal(t, "./data_base.json", cfg.BaseDataFilePath)
}

// TestGetAgentConfigWithoutServerURL проверяет, что без адреса сервера
// клиент не запускается, а не ходит молча в localhost.
func TestGetAgentConfigWithoutServerURL(t *testing.T) {
	t.Setenv(serverURLEnv, "")

	_, err := GetAgentConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), serverURLEnv)
}
