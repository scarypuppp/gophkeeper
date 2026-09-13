package entities

import (
	"fmt"
)

const (
	MinLoginLength    = 3
	MaxLoginLength    = 64
	MinPasswordLength = 5
	MaxPasswordLength = 72
)

var (
	// ErrIncorrectLoginLength возвращается при несоответствии длины login допустимому диапазону.
	ErrIncorrectLoginLength = fmt.Errorf("login length should be between %d and %d", MinLoginLength, MaxLoginLength)
	// ErrIncorrectPasswordLength возвращается при несоответствии длины password допустимому диапазону.
	ErrIncorrectPasswordLength = fmt.Errorf("password length should be between %d and %d", MinPasswordLength, MaxPasswordLength)
)

// User представляет зарегистрированного пользователя системы.
type User struct {
	ID       int64  `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

// ValidateLogin проверяет, что длина login укладывается в допустимый диапазон.
func ValidateLogin(login string) error {
	if len(login) < MinLoginLength || len(login) > MaxLoginLength {
		return ErrIncorrectLoginLength
	}
	return nil
}

// ValidatePassword проверяет, что длина password укладывается в допустимый диапазон.
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
		return ErrIncorrectPasswordLength
	}
	return nil
}
