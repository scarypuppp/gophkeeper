package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/entities"
	"github.com/scarypuppp/gophkeeper/internal/server/auth"
	"github.com/scarypuppp/gophkeeper/internal/server/repository"
)

// ErrLoginAlreadyExists возвращается при попытке зарегистрировать уже занятый login.
var ErrLoginAlreadyExists = errors.New("user with such login already exists")

// ErrLoginPasswordNotExist возвращается, если пользователь с указанными login и password не найден.
var ErrLoginPasswordNotExist = errors.New("user with such login and password does not exist")

// UserService реализует бизнес-логику регистрации и аутентификации пользователей.
type UserService struct {
	uow repository.UnitOfWork
}

// NewUserService создаёт новый UserService с переданным UnitOfWork.
func NewUserService(uow repository.UnitOfWork) *UserService {
	return &UserService{uow}
}

// RegisterUser создаёт нового пользователя с указанными login и password.
// Валидирует длину login и password, проверяет уникальность login.
// Возвращает ErrLoginAlreadyExists, если login уже занят.
func (us *UserService) RegisterUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {

	err := entities.ValidateLogin(login)
	if err != nil {
		return nil, err
	}
	err = entities.ValidatePassword(password)
	if err != nil {
		return nil, err
	}

	tx, err := us.uow.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Users().GetByLogin(ctx, login)
	if err == nil {
		return nil, ErrLoginAlreadyExists
	}
	if !errors.Is(err, repository.ErrNoRows) {
		return nil, fmt.Errorf("RegisterUser: check login: %w", err)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: hash password: %w", err)
	}

	user, err := tx.Users().CreateUser(ctx, entities.User{Login: login, Password: passwordHash})
	if err != nil {
		return nil, fmt.Errorf("RegisterUser: create user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("RegisterUser: commit: %w", err)
	}
	return user, nil
}

// LoginUser проверяет login и password, возвращает User при успешной аутентификации.
// Возвращает ErrLoginPasswordNotExist, если пользователь не найден или пароль неверен.
func (us *UserService) LoginUser(
	ctx context.Context,
	login string,
	password string,
) (*entities.User, error) {
	user, err := us.uow.Users().GetByLogin(ctx, login)
	if errors.Is(err, repository.ErrNoRows) {
		return nil, ErrLoginPasswordNotExist
	}
	if err != nil {
		return nil, fmt.Errorf("LoginUser: %w", err)
	}
	err = auth.CheckPassword(password, user.Password)
	if err != nil {
		return nil, ErrLoginPasswordNotExist
	}

	return user, nil
}
