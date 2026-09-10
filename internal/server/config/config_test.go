package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fullEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SECRET_KEY", "s")
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://localhost/db")
}

func TestParseConfig_FromEnv(t *testing.T) {
	fullEnv(t)
	cfg, err := parseConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, "s", cfg.SecretKey)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, int64(defaultTokenExpiresSeconds), cfg.TokenExpSeconds)
}

func TestParseConfig_FromFlags(t *testing.T) {
	t.Setenv("SECRET_KEY", "")
	cfg, err := parseConfig([]string{
		"-k", "mykey",
		"-a", "0.0.0.0:8080",
		"-d", "postgres://localhost/test",
		"-e", "3600",
	})
	require.NoError(t, err)
	assert.Equal(t, "mykey", cfg.SecretKey)
	assert.Equal(t, "0.0.0.0:8080", cfg.Address)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURI)
	assert.Equal(t, int64(3600), cfg.TokenExpSeconds)
}

func TestParseConfig_MissingAddress(t *testing.T) {
	t.Setenv("SECRET_KEY", "s")
	t.Setenv("DATABASE_URI", "postgres://localhost/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "localhost:9090")
	_, err := parseConfig(nil)
	assert.ErrorContains(t, err, "service address")
}

func TestParseConfig_MissingDatabaseURI(t *testing.T) {
	t.Setenv("SECRET_KEY", "s")
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "localhost:9090")
	_, err := parseConfig(nil)
	assert.ErrorContains(t, err, "database uri")
}

func TestParseConfig_MissingSecretKey(t *testing.T) {
	t.Setenv("SECRET_KEY", "")
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://localhost/db")
	_, err := parseConfig(nil)
	assert.ErrorContains(t, err, "secret key")
}

func TestParseConfig_Defaults(t *testing.T) {
	t.Setenv("SECRET_KEY", "s")
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://localhost/db")
	cfg, err := parseConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, int64(defaultTokenExpiresSeconds), cfg.TokenExpSeconds)
}
