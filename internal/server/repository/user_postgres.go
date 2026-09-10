package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophkeeper/internal/entities"
)

// UserRepositoryPostgres реализует UserRepository поверх PostgreSQL.
type UserRepositoryPostgres struct {
	exec sqlx.ExtContext
}

// NewUserRepositoryPostgres создаёт UserRepositoryPostgres, работающий без транзакции.
func NewUserRepositoryPostgres(db *sqlx.DB) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{exec: db}
}

// NewUserRepositoryPostgresTx создаёт UserRepositoryPostgres, работающий внутри транзакции.
func NewUserRepositoryPostgresTx(tx *sqlx.Tx) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{exec: tx}
}

func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	var user entities.User
	err := sqlx.GetContext(ctx, r.exec, &user,
		`SELECT id, login, password FROM users WHERE id = $1`, id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	var user entities.User
	err := sqlx.GetContext(ctx, r.exec, &user,
		`SELECT id, login, password FROM users WHERE login = $1`, login,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByLogin: %w", err)
	}
	return &user, nil
}
func (r *UserRepositoryPostgres) CreateUser(ctx context.Context, user entities.User) (*entities.User, error) {
	const insertQuery = `
    INSERT INTO users (login, password)
    VALUES (:login, :password)
    RETURNING id`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, insertQuery, user)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("CreateUser: rows iteration: %w", err)
		}
		return nil, fmt.Errorf("CreateUser: no id returned after insert")
	}
	if err := rows.Scan(&user.ID); err != nil {
		return nil, fmt.Errorf("CreateUser: scan returned id: %w", err)
	}
	return &user, nil
}
