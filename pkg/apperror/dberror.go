package apperror

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func MapDBError(op string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: %w", op, err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", op, ErrNotFound)
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505": // unique_violation (SKU 	ซ้ำ)
			return fmt.Errorf("%s: %w", op, ErrConflict)
		case "23503", "23514":
			return fmt.Errorf("%s: %w", op, ErrInvalidInput)
		}
	}
	return fmt.Errorf("%s: %w", op, ErrInternalServer)
}
