package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// uniqueViolationCode — код ошибки PostgreSQL при нарушении уникального ограничения.
const uniqueViolationCode = "23505"

// ErrUniqueViolation возвращается репозиторием, когда вставка нарушила уникальное ограничение.
var ErrUniqueViolation = errors.New("unique constraint violation")

// wrapError оборачивает ошибку запроса, подменяя нарушение уникальности на ErrUniqueViolation.
func wrapError(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return fmt.Errorf("%s: %w", op, ErrUniqueViolation)
	}
	return fmt.Errorf("%s: %w", op, err)
}
