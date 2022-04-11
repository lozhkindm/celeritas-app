package celeritas

import (
	"database/sql"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func (c *Celeritas) OpenDB(dbType, dsn string) (*sql.DB, error) {
	if dbType == "postgres" || dbType == "postgresql" {
		dbType = "pgx"
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
