package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func SetupPostgres() (*sql.DB, error) {
	dsn := "host=localhost user=postgres dbname=audiodrive sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
