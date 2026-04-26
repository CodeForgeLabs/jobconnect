package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
