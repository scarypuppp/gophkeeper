package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// UnitOfWorkPostgres реализует UnitOfWork поверх PostgreSQL.
// Может работать как без транзакции (db), так и внутри открытой транзакции (tx).
type UnitOfWorkPostgres struct {
	db      *sqlx.DB
	tx      *sqlx.Tx
	users   *UserRepositoryPostgres
	files   *FileRepositoryPostgres
	secrets *SecretRepositoryPostgres
	cards   *CardRepositoryPostgres
	texts   *TextRepositoryPostgres
}

// NewUnitOfWorkPostgres создаёт UnitOfWorkPostgres с подключением к базе данных.
func NewUnitOfWorkPostgres(db *sqlx.DB) *UnitOfWorkPostgres {
	return &UnitOfWorkPostgres{
		db:      db,
		users:   NewUserRepositoryPostgres(db),
		files:   NewFileRepositoryPostgres(db),
		secrets: NewSecretRepositoryPostgres(db),
		cards:   NewCardRepositoryPostgres(db),
		texts:   NewTextRepositoryPostgres(db),
	}
}

func (u *UnitOfWorkPostgres) Users() UserRepository     { return u.users }
func (u *UnitOfWorkPostgres) Files() FileRepository     { return u.files }
func (u *UnitOfWorkPostgres) Secrets() SecretRepository { return u.secrets }
func (u *UnitOfWorkPostgres) Cards() CardRepository     { return u.cards }
func (u *UnitOfWorkPostgres) Texts() TextRepository     { return u.texts }

func (u *UnitOfWorkPostgres) BeginTx(ctx context.Context) (UnitOfWork, error) {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &UnitOfWorkPostgres{
		db:      u.db,
		tx:      tx,
		users:   NewUserRepositoryPostgresTx(tx),
		files:   NewFileRepositoryPostgresTx(tx),
		secrets: NewSecretRepositoryPostgresTx(tx),
		cards:   NewCardRepositoryPostgresTx(tx),
		texts:   NewTextRepositoryPostgresTx(tx),
	}, nil
}

func (u *UnitOfWorkPostgres) Commit(_ context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Commit()
}

func (u *UnitOfWorkPostgres) Rollback(_ context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Rollback()
}
