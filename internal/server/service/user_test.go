package service

import (
	"context"
	"testing"

	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/auth"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
	mocks2 "github.com/scarypuppp/gophkeeper/internal/server/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupUserMocks(t *testing.T) (
	*gomock.Controller,
	*mocks2.MockUnitOfWork,
	*mocks2.MockUnitOfWork,
	*mocks2.MockUserRepository,
) {
	ctrl := gomock.NewController(t)
	uowMock := mocks2.NewMockUnitOfWork(ctrl)
	uowTxMock := mocks2.NewMockUnitOfWork(ctrl)
	userRepoMock := mocks2.NewMockUserRepository(ctrl)
	return ctrl, uowMock, uowTxMock, userRepoMock
}

func TestRegisterUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		_, uowMock, uowTxMock, userRepoMock := setupUserMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Users().Return(userRepoMock).Times(2)
		userRepoMock.EXPECT().
			GetByLogin(ctx, "user").
			Return(nil, repository.ErrNoRows)
		userRepoMock.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(&entities.User{ID: 1, Login: "user"}, nil)
		uowTxMock.EXPECT().Commit(ctx).Return(nil)

		us := NewUserService(uowMock)
		user, err := us.RegisterUser(ctx, "user", "12345")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user", user.Login)
		assert.NotZero(t, user.ID)
	})

	t.Run("incorrect login", func(t *testing.T) {
		_, uowMock, _, _ := setupUserMocks(t)

		us := NewUserService(uowMock)
		_, err := us.RegisterUser(ctx, "", "12345")

		assert.ErrorIs(t, err, entities.ErrIncorrectLoginLength)
	})

	t.Run("incorrect password", func(t *testing.T) {
		_, uowMock, _, _ := setupUserMocks(t)

		us := NewUserService(uowMock)
		_, err := us.RegisterUser(ctx, "user", "1")

		assert.ErrorIs(t, err, entities.ErrIncorrectPasswordLength)
	})

	t.Run("duplicate login", func(t *testing.T) {
		_, uowMock, uowTxMock, userRepoMock := setupUserMocks(t)

		uowMock.EXPECT().BeginTx(ctx).Return(uowTxMock, nil)
		uowTxMock.EXPECT().Rollback(ctx).Return(nil)
		uowTxMock.EXPECT().Users().Return(userRepoMock)
		userRepoMock.EXPECT().
			GetByLogin(ctx, "user").
			Return(&entities.User{ID: 1, Login: "user"}, nil)

		us := NewUserService(uowMock)
		_, err := us.RegisterUser(ctx, "user", "12345")

		assert.ErrorIs(t, err, ErrLoginAlreadyExists)
	})
}

func TestLoginUser(t *testing.T) {
	ctx := context.Background()
	passwordHash, _ := auth.HashPassword("12345")
	existingUser := &entities.User{ID: 1, Login: "user", Password: passwordHash}

	t.Run("success", func(t *testing.T) {
		_, uowMock, _, userRepoMock := setupUserMocks(t)

		uowMock.EXPECT().Users().Return(userRepoMock)
		userRepoMock.EXPECT().
			GetByLogin(ctx, "user").
			Return(existingUser, nil)

		us := NewUserService(uowMock)
		user, err := us.LoginUser(ctx, "user", "12345")

		assert.NoError(t, err)
		assert.Equal(t, existingUser.Login, user.Login)
		assert.Equal(t, existingUser.ID, user.ID)
	})

	t.Run("user not exist", func(t *testing.T) {
		_, uowMock, _, userRepoMock := setupUserMocks(t)

		uowMock.EXPECT().Users().Return(userRepoMock)
		userRepoMock.EXPECT().
			GetByLogin(ctx, "user2").
			Return(nil, repository.ErrNoRows)

		us := NewUserService(uowMock)
		user, err := us.LoginUser(ctx, "user2", "12345")

		assert.ErrorIs(t, err, ErrLoginPasswordNotExist)
		assert.Nil(t, user)
	})

	t.Run("incorrect password", func(t *testing.T) {
		_, uowMock, _, userRepoMock := setupUserMocks(t)

		uowMock.EXPECT().Users().Return(userRepoMock)
		userRepoMock.EXPECT().
			GetByLogin(ctx, "user").
			Return(existingUser, nil)

		us := NewUserService(uowMock)
		user, err := us.LoginUser(ctx, "user", "54321")

		assert.ErrorIs(t, err, ErrLoginPasswordNotExist)
		assert.Nil(t, user)
	})
}
