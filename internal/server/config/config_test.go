package config

import (
	"os"
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

func TestParseConfig_TokenExpFromEnv(t *testing.T) {
	fullEnv(t)
	t.Setenv("TOKEN_EXP", "600")
	cfg, err := parseConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, int64(600), cfg.TokenExpSeconds)
}

func TestParseConfig_TokenExpEnvOverridesFlag(t *testing.T) {
	fullEnv(t)
	t.Setenv("TOKEN_EXP", "600")
	cfg, err := parseConfig([]string{"-e", "3600"})
	require.NoError(t, err)
	assert.Equal(t, int64(600), cfg.TokenExpSeconds)
}

func TestParseConfig_S3FromEnv(t *testing.T) {
	fullEnv(t)
	t.Setenv("S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("S3_ACCESS_KEY", "access")
	t.Setenv("S3_SECRET_KEY", "secret")
	t.Setenv("S3_BUCKET", "bucket")

	cfg, err := parseConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9000", cfg.S3Endpoint)
	assert.Equal(t, "access", cfg.S3AccessKey)
	assert.Equal(t, "secret", cfg.S3SecretKey)
	assert.Equal(t, "bucket", cfg.S3Bucket)
}

func TestParseConfig_EnvOverridesFlags(t *testing.T) {
	fullEnv(t)
	cfg, err := parseConfig([]string{
		"-k", "flagkey",
		"-a", "0.0.0.0:9090",
		"-d", "postgres://localhost/other",
	})
	require.NoError(t, err)
	assert.Equal(t, "s", cfg.SecretKey)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "postgres://localhost/db", cfg.DatabaseURI)
}

func TestParseConfig_MissingSecretKeyMessage(t *testing.T) {
	t.Setenv("SECRET_KEY", "")
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://localhost/db")
	_, err := parseConfig([]string{"-a", "localhost:8080", "-d", "postgres://localhost/db"})
	assert.ErrorContains(t, err, "SECRET_KEY")
	assert.ErrorContains(t, err, "-k")
}

func TestParseConfig_EnvParseError(t *testing.T) {
	fullEnv(t)
	t.Setenv("TOKEN_EXP", "not-a-number")
	_, err := parseConfig(nil)
	assert.Error(t, err)
}

func TestParseConfig_InvalidFlag(t *testing.T) {
	fullEnv(t)
	_, err := parseConfig([]string{"-e", "not-a-number"})
	assert.Error(t, err)
}

func TestParseConfig_UnknownFlag(t *testing.T) {
	fullEnv(t)
	_, err := parseConfig([]string{"-unknown", "value"})
	assert.Error(t, err)
}

func TestGetConfig(t *testing.T) {
	fullEnv(t)

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"gophkeeper"}

	cfg, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "s", cfg.SecretKey)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "postgres://localhost/db", cfg.DatabaseURI)
}

func TestGetConfig_UsesArgsFlags(t *testing.T) {
	t.Setenv("SECRET_KEY", "")
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"gophkeeper", "-k", "flagkey", "-a", "localhost:8080", "-d", "postgres://localhost/db"}

	cfg, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "flagkey", cfg.SecretKey)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "postgres://localhost/db", cfg.DatabaseURI)
}
