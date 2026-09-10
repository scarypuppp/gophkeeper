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

// CardRepositoryPostgres реализует CardRepository поверх PostgreSQL.
type CardRepositoryPostgres struct {
	exec sqlx.ExtContext
}

// NewCardRepositoryPostgres создаёт CardRepositoryPostgres, работающий без транзакции.
func NewCardRepositoryPostgres(db *sqlx.DB) *CardRepositoryPostgres {
	return &CardRepositoryPostgres{exec: db}
}

// NewCardRepositoryPostgresTx создаёт CardRepositoryPostgres, работающий внутри транзакции.
func NewCardRepositoryPostgresTx(tx *sqlx.Tx) *CardRepositoryPostgres {
	return &CardRepositoryPostgres{exec: tx}
}

func (r *CardRepositoryPostgres) GetByName(ctx context.Context, owner int64, name string) (*entities.Card, error) {
	var card entities.Card
	err := sqlx.GetContext(ctx, r.exec, &card,
		`SELECT id, owner, name, number, holder, expires_at, cvv, metadata, checksum, created_at, updated_at FROM cards WHERE owner = $1 AND name = $2`,
		owner, name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("GetByName: %w", err)
	}
	return &card, nil
}

func (r *CardRepositoryPostgres) GetCardsByOwner(ctx context.Context, owner int64) ([]entities.Card, error) {
	var cards []entities.Card
	err := sqlx.SelectContext(ctx, r.exec, &cards,
		`SELECT id, owner, name, number, holder, expires_at, cvv, metadata, checksum, created_at, updated_at FROM cards WHERE owner = $1 ORDER BY name`,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("GetCardsByOwner: %w", err)
	}
	return cards, nil
}

func (r *CardRepositoryPostgres) CreateCard(ctx context.Context, card entities.Card) (*entities.Card, error) {
	const insertQuery = `
    INSERT INTO cards (owner, name, number, holder, expires_at, cvv, metadata, checksum)
    VALUES (:owner, :name, :number, :holder, :expires_at, :cvv, :metadata, :checksum)
    RETURNING id, created_at, updated_at`
	rows, err := sqlx.NamedQueryContext(ctx, r.exec, insertQuery, card)
	if err != nil {
		return nil, wrapError("CreateCard", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, wrapError("CreateCard: rows iteration", err)
		}
		return nil, fmt.Errorf("CreateCard: no id returned after insert")
	}
	if err := rows.Scan(&card.Id, &card.CreatedAt, &card.UpdatedAt); err != nil {
		return nil, fmt.Errorf("CreateCard: scan returned row: %w", err)
	}
	return &card, nil
}

// UpdateCard перезаписывает содержимое записи и возвращает новое время изменения.
func (r *CardRepositoryPostgres) UpdateCard(ctx context.Context, card entities.Card) (time.Time, error) {
	var updatedAt time.Time
	err := sqlx.GetContext(ctx, r.exec, &updatedAt,
		`UPDATE cards SET number = $1, holder = $2, expires_at = $3, cvv = $4, metadata = $5, checksum = $6, updated_at = now()
         WHERE owner = $7 AND name = $8
         RETURNING updated_at`,
		card.Number, card.Holder, card.ExpiresAt, card.CVV, card.Metadata, card.Checksum, card.Owner, card.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, ErrNoRows
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("UpdateCard: %w", err)
	}
	return updatedAt, nil
}

func (r *CardRepositoryPostgres) DeleteCard(ctx context.Context, owner int64, name string) error {
	result, err := r.exec.ExecContext(ctx,
		`DELETE FROM cards WHERE owner = $1 AND name = $2`, owner, name,
	)
	if err != nil {
		return fmt.Errorf("DeleteCard: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteCard: %w", err)
	}
	if affected == 0 {
		return ErrNoRows
	}
	return nil
}
