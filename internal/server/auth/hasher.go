package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword возвращает bcrypt-хэш переданного пароля.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword проверяет, соответствует ли password переданному bcrypt-хэшу.
// Возвращает nil при совпадении и ошибку в противном случае.
func CheckPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
