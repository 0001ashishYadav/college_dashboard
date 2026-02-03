package pgdb

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	ErrorDuplicateKey = "23505"
	ErrorUniqueKey    = "23514"
	ErrorNotNull      = "23502"
	ErrorForeignKey   = "23503"
	ErrorRelation     = "42P01"
	ErrorNoRow        = "no rows in result set"
)

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return err.Error()
}
