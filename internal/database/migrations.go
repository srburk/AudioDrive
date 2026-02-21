package database

import "database/sql"

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE schema_version (
			version text not null
		);

		CREATE TABLE users (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			email TEXT NOT NULL UNIQUE
		);

		CREATE TABLE objects (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id),
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			duration_seconds INT
		)
	`)
	return err
}

//			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
