package storage

import (
	"fmt"
	"slices"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
)

// SecretPatch описывает частичное изменение секрета: nil-поля остаются прежними.
type SecretPatch struct {
	Login    *string
	Password *string
	Metadata *string
}

// Secrets возвращает секреты, упорядоченные по имени.
func (s *LocalStorage) Secrets() []Secret {
	return sortedByName(s.secrets)
}

// GetSecret возвращает секрет по имени.
// Возвращает ErrNotFound, если секрета с таким именем нет.
func (s *LocalStorage) GetSecret(name string) (Secret, error) {
	i := indexOf(s.secrets, name)
	if i < 0 {
		return Secret{}, fmt.Errorf("secret %q: %w", name, ErrNotFound)
	}
	return s.secrets[i], nil
}

// AddSecret добавляет секрет; контрольную сумму и время изменения
// проставляет хранилище.
// Возвращает ErrAlreadyExists, если секрет с таким именем уже есть.
func (s *LocalStorage) AddSecret(secret Secret) (Secret, error) {
	if secret.Name == "" {
		return Secret{}, ErrEmptyName
	}
	if indexOf(s.secrets, secret.Name) >= 0 {
		return Secret{}, fmt.Errorf("secret %q: %w", secret.Name, ErrAlreadyExists)
	}

	secret.Checksum = checksum.Secret(secret.Login, secret.Password, secret.Metadata)
	secret.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Secrets = append(data.Secrets, secret)
	if err := s.commit(data); err != nil {
		return Secret{}, err
	}
	return secret, nil
}

// UpdateSecret меняет у секрета только переданные в patch поля.
// Возвращает ErrNotFound, если секрета с таким именем нет.
func (s *LocalStorage) UpdateSecret(name string, patch SecretPatch) (Secret, error) {
	i := indexOf(s.secrets, name)
	if i < 0 {
		return Secret{}, fmt.Errorf("secret %q: %w", name, ErrNotFound)
	}

	secret := s.secrets[i]
	if patch.Login != nil {
		secret.Login = *patch.Login
	}
	if patch.Password != nil {
		secret.Password = *patch.Password
	}
	if patch.Metadata != nil {
		secret.Metadata = *patch.Metadata
	}
	secret.Checksum = checksum.Secret(secret.Login, secret.Password, secret.Metadata)
	secret.UpdatedAt = s.now()

	data := s.Snapshot()
	data.Secrets[i] = secret
	if err := s.commit(data); err != nil {
		return Secret{}, err
	}
	return secret, nil
}

// DeleteSecret удаляет секрет по имени.
// Возвращает ErrNotFound, если секрета с таким именем нет.
func (s *LocalStorage) DeleteSecret(name string) error {
	i := indexOf(s.secrets, name)
	if i < 0 {
		return fmt.Errorf("secret %q: %w", name, ErrNotFound)
	}

	data := s.Snapshot()
	data.Secrets = slices.Delete(data.Secrets, i, i+1)
	return s.commit(data)
}
