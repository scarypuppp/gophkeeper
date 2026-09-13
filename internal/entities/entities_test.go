package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogin(t *testing.T) {
	assert.NoError(t, ValidateLogin("alice"))
	assert.ErrorIs(t, ValidateLogin("ab"), ErrIncorrectLoginLength)
	assert.ErrorIs(t, ValidateLogin(""), ErrIncorrectLoginLength)
}

func TestValidatePassword(t *testing.T) {
	assert.NoError(t, ValidatePassword("secret"))
	assert.ErrorIs(t, ValidatePassword("abc"), ErrIncorrectPasswordLength)
	assert.ErrorIs(t, ValidatePassword(""), ErrIncorrectPasswordLength)
}
