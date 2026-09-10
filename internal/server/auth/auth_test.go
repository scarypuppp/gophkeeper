package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	assert.NoError(t, CheckPassword("secret", hash))
	assert.Error(t, CheckPassword("wrong", hash))
}

func TestCreateAndParseToken(t *testing.T) {
	const key = "test-secret"
	const userID int64 = 42
	const exp int64 = 3600

	token, err := CreateToken(key, userID, exp)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	got, err := ParseToken(key, token)
	require.NoError(t, err)
	assert.Equal(t, userID, got)
}

func TestParseToken_WrongKey(t *testing.T) {
	token, err := CreateToken("key-a", 1, 3600)
	require.NoError(t, err)

	_, err = ParseToken("key-b", token)
	assert.Error(t, err)
}

func TestParseToken_Expired(t *testing.T) {
	token, err := CreateToken("key", 1, -1)
	require.NoError(t, err)

	_, err = ParseToken("key", token)
	assert.Error(t, err)
}

func TestParseToken_Malformed(t *testing.T) {
	_, err := ParseToken("key", "not.a.token")
	assert.Error(t, err)
}
