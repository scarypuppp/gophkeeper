package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/scarypuppp/gophkeeper/internal/checksum"
	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
	"github.com/scarypuppp/gophkeeper/internal/server/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupSecretMocks(t *testing.T) (*mocks.MockUnitOfWork, *mocks.MockSecretRepository) {
	ctrl := gomock.NewController(t)
	uowMock := mocks.NewMockUnitOfWork(ctrl)
	secretRepoMock := mocks.NewMockSecretRepository(ctrl)
	return uowMock, secretRepoMock
}

func TestGetSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		existing := &entities.Secret{Id: 1, Owner: 1, Name: "name"}

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(existing, nil)

		s := NewSecretsService(uowMock)
		secret, err := s.GetSecret(ctx, 1, "name")

		assert.NoError(t, err)
		assert.Equal(t, existing, secret)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewSecretsService(uowMock)
		secret, err := s.GetSecret(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrSecretNotExists)
		assert.Nil(t, secret)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, wantErr)

		s := NewSecretsService(uowMock)
		secret, err := s.GetSecret(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, secret)
	})
}

func TestGetSecrets(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		want := []entities.Secret{{Id: 1, Owner: 1, Name: "a"}, {Id: 2, Owner: 1, Name: "b"}}

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetSecretsByOwner(ctx, int64(1)).Return(want, nil)

		s := NewSecretsService(uowMock)
		secrets, err := s.GetSecrets(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, want, secrets)
	})

	t.Run("error", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetSecretsByOwner(ctx, int64(1)).Return(nil, wantErr)

		s := NewSecretsService(uowMock)
		secrets, err := s.GetSecrets(ctx, 1)

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, secrets)
	})
}

func TestCreateSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		created := &entities.Secret{Id: 1, Owner: 1, Name: "name", Login: "login", Password: "pass", Metadata: "meta"}

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().
			CreateSecret(ctx, entities.Secret{
				Owner:    1,
				Name:     "name",
				Login:    "login",
				Password: "pass",
				Metadata: "meta",
				Checksum: checksum.Secret("login", "pass", "meta"),
			}).
			Return(created, nil)

		s := NewSecretsService(uowMock)
		secret, err := s.CreateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.NoError(t, err)
		assert.Equal(t, created, secret)
	})

	t.Run("already exists", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().
			CreateSecret(ctx, gomock.Any()).
			Return(nil, repository.ErrUniqueViolation)

		s := NewSecretsService(uowMock)
		secret, err := s.CreateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.ErrorIs(t, err, ErrSecretAlreadyExists)
		assert.Nil(t, secret)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().
			CreateSecret(ctx, gomock.Any()).
			Return(nil, wantErr)

		s := NewSecretsService(uowMock)
		secret, err := s.CreateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, secret)
	})
}

func TestUpdateSecret(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Now().Add(-time.Hour)

	t.Run("success", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		current := &entities.Secret{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		updatedAt := time.Now()

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		expected := entities.Secret{
			Id:        1,
			Owner:     1,
			Name:      "name",
			Login:     "newLogin",
			Password:  "newPass",
			Metadata:  "newMeta",
			Checksum:  checksum.Secret("newLogin", "newPass", "newMeta"),
			CreatedAt: createdAt,
		}
		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().UpdateSecret(ctx, expected).Return(updatedAt, nil)

		s := NewSecretsService(uowMock)
		secret, err := s.UpdateSecret(ctx, 1, "name", "newLogin", "newPass", "newMeta")

		assert.NoError(t, err)
		assert.Equal(t, expected.Id, secret.Id)
		assert.Equal(t, expected.Login, secret.Login)
		assert.Equal(t, updatedAt, secret.UpdatedAt)
	})

	t.Run("current not exist", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(nil, repository.ErrNoRows)

		s := NewSecretsService(uowMock)
		secret, err := s.UpdateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.ErrorIs(t, err, ErrSecretNotExists)
		assert.Nil(t, secret)
	})

	t.Run("update not exist", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		current := &entities.Secret{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().UpdateSecret(ctx, gomock.Any()).Return(time.Time{}, repository.ErrNoRows)

		s := NewSecretsService(uowMock)
		secret, err := s.UpdateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.ErrorIs(t, err, ErrSecretNotExists)
		assert.Nil(t, secret)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		current := &entities.Secret{Id: 1, Owner: 1, Name: "name", CreatedAt: createdAt}
		wantErr := errors.New("db error")

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().GetByName(ctx, int64(1), "name").Return(current, nil)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().UpdateSecret(ctx, gomock.Any()).Return(time.Time{}, wantErr)

		s := NewSecretsService(uowMock)
		secret, err := s.UpdateSecret(ctx, 1, "name", "login", "pass", "meta")

		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, secret)
	})
}

func TestDeleteSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().DeleteSecret(ctx, int64(1), "name").Return(nil)

		s := NewSecretsService(uowMock)
		err := s.DeleteSecret(ctx, 1, "name")

		assert.NoError(t, err)
	})

	t.Run("not exist", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().DeleteSecret(ctx, int64(1), "name").Return(repository.ErrNoRows)

		s := NewSecretsService(uowMock)
		err := s.DeleteSecret(ctx, 1, "name")

		assert.ErrorIs(t, err, ErrSecretNotExists)
	})

	t.Run("unexpected error", func(t *testing.T) {
		uowMock, secretRepoMock := setupSecretMocks(t)
		wantErr := errors.New("db error")

		uowMock.EXPECT().Secrets().Return(secretRepoMock)
		secretRepoMock.EXPECT().DeleteSecret(ctx, int64(1), "name").Return(wantErr)

		s := NewSecretsService(uowMock)
		err := s.DeleteSecret(ctx, 1, "name")

		assert.ErrorIs(t, err, wantErr)
	})
}
