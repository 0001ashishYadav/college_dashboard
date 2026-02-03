package pgdb

import "github.com/jackc/pgx/v5/pgxpool"

type Store interface {
	Querier
}

type SqlStore struct {
	Querier
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) Store {
	return &SqlStore{
		db:      db,
		Querier: New(db),
	}
}
