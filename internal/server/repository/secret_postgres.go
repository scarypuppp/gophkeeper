package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophkeeper/internal/entities"
)

// SecretRepositoryPostgres реализует SecretRepository поверх PostgreSQL.
type SecretRepositoryPostgres struct {
	exec sqlx.ExtContext
}

// NewSecretRepositoryPostgres создаёт SecretRepositoryPostgres, работающий без транзакции.
func NewSecretRepositoryPostgres(db *sqlx.DB) *SecretRepositoryPostgres {
	return &SecretRepositoryPostgres{exec: db}
}

// NewSecretRepositoryPostgresTx создаёт SecretRepositoryPostgres, работающий внутри транзакции.
func NewSecretRepositoryPostgresTx(tx *sqlx.Tx) *SecretRepositoryPostgres {
	return &SecretRepositoryPostgres{exec: tx}
}

func (r *SecretRepositoryPostgres) GetByName(ctx context.Context, owner int64, name string) (*entities.Secret, error) {
	var secret entities.Secret
	err := sqlx.GetContext(ctx, r.exec, &secret,
		`SELECT id, owner, name, login, password, metadata, checksum, created_at, updated_at FROM secrets WHERE owner = $1 AND name = $2`,
		owner, name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	return &secret, nil
}

func (r *SecretRepositoryPostgres) GetSecretsByOwner(ctx context.Context, owner int64) ([]entities.Secret, error) {
	var secrets []entities.Secret
	err := sqlx.SelectContext(ctx, r.exec, &secrets,
		`SELECT id, owner, name, login, password, metadata, checksum, created_at, updated_at FROM secrets WHERE owner = $1 ORDER BY name`,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSecretsByOwner: %w", err)
	}
	return secrets, nil
}

func (r *SecretRepositoryPostgres) CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	const insertQuery = `
    INSERT INTO secrets (owner, name, login, password, metadata, checksum)
    VALUES (:owner, :name, :login, :password, :metadata, :checksum)
    RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, insertQuery, secret)
	if err != nil {
		return nil, wrapError("CreateSecret", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, wrapError("CreateSecret: rows iteration", err)
		}
		return nil, fmt.Errorf("CreateSecret: no id returned after insert")
	}
	if err := rows.Scan(&secret.Id, &secret.CreatedAt, &secret.UpdatedAt); err != nil {
		return nil, fmt.Errorf("CreateSecret: scan returned row: %w", err)
	}
	return &secret, nil
}

// UpdateSecret перезаписывает содержимое записи и возвращает новое время изменения.
func (r *SecretRepositoryPostgres) UpdateSecret(ctx context.Context, secret entities.Secret) (time.Time, error) {
	var updatedAt time.Time
	err := sqlx.GetContext(ctx, r.exec, &updatedAt,
		`UPDATE secrets SET login = $1, password = $2, metadata = $3, checksum = $4, updated_at = now()
         WHERE owner = $5 AND name = $6
         RETURNING updated_at`,
		secret.Login, secret.Password, secret.Metadata, secret.Checksum, secret.Owner, secret.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNoRows
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("UpdateSecret: %w", err)
	}
	return updatedAt, nil
}

func (r *SecretRepositoryPostgres) DeleteSecret(ctx context.Context, owner int64, name string) error {
	result, err := r.exec.ExecContext(ctx,
		`DELETE FROM secrets WHERE owner = $1 AND name = $2`, owner, name,
	)
	if err != nil {
		return fmt.Errorf("DeleteSecret: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteSecret: %w", err)
	}
	if affected == 0 {
		return ErrNoRows
	}
	return nil
}
