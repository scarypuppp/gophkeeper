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

// TextRepositoryPostgres реализует TextRepository поверх PostgreSQL.
type TextRepositoryPostgres struct {
	exec sqlx.ExtContext
}

// NewTextRepositoryPostgres создаёт TextRepositoryPostgres, работающий без транзакции.
func NewTextRepositoryPostgres(db *sqlx.DB) *TextRepositoryPostgres {
	return &TextRepositoryPostgres{exec: db}
}

// NewTextRepositoryPostgresTx создаёт TextRepositoryPostgres, работающий внутри транзакции.
func NewTextRepositoryPostgresTx(tx *sqlx.Tx) *TextRepositoryPostgres {
	return &TextRepositoryPostgres{exec: tx}
}

func (r *TextRepositoryPostgres) GetByName(ctx context.Context, owner int64, name string) (*entities.Text, error) {
	var text entities.Text
	err := sqlx.GetContext(ctx, r.exec, &text,
		`SELECT id, owner, name, text, metadata, checksum, created_at, updated_at FROM texts WHERE owner = $1 AND name = $2`,
		owner, name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	return &text, nil
}

func (r *TextRepositoryPostgres) GetTextsByOwner(ctx context.Context, owner int64) ([]entities.Text, error) {
	var texts []entities.Text
	err := sqlx.SelectContext(ctx, r.exec, &texts,
		`SELECT id, owner, name, text, metadata, checksum, created_at, updated_at FROM texts WHERE owner = $1 ORDER BY name`,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTextsByOwner: %w", err)
	}
	return texts, nil
}

func (r *TextRepositoryPostgres) CreateText(ctx context.Context, text entities.Text) (*entities.Text, error) {
	const insertQuery = `
    INSERT INTO texts (owner, name, text, metadata, checksum)
    VALUES (:owner, :name, :text, :metadata, :checksum)
    RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, insertQuery, text)
	if err != nil {
		return nil, wrapError("CreateText", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, wrapError("CreateText: rows iteration", err)
		}
		return nil, fmt.Errorf("CreateText: no id returned after insert")
	}
	if err := rows.Scan(&text.Id, &text.CreatedAt, &text.UpdatedAt); err != nil {
		return nil, fmt.Errorf("CreateText: scan returned row: %w", err)
	}
	return &text, nil
}

// UpdateText перезаписывает содержимое записи и возвращает новое время изменения.
func (r *TextRepositoryPostgres) UpdateText(ctx context.Context, text entities.Text) (time.Time, error) {
	var updatedAt time.Time
	err := sqlx.GetContext(ctx, r.exec, &updatedAt,
		`UPDATE texts SET text = $1, metadata = $2, checksum = $3, updated_at = now()
         WHERE owner = $4 AND name = $5
         RETURNING updated_at`,
		text.Text, text.Metadata, text.Checksum, text.Owner, text.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNoRows
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("UpdateText: %w", err)
	}
	return updatedAt, nil
}

func (r *TextRepositoryPostgres) DeleteText(ctx context.Context, owner int64, name string) error {
	result, err := r.exec.ExecContext(ctx,
		`DELETE FROM texts WHERE owner = $1 AND name = $2`, owner, name,
	)
	if err != nil {
		return fmt.Errorf("DeleteText: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteText: %w", err)
	}
	if affected == 0 {
		return ErrNoRows
	}
	return nil
}
