package store

import (
	"context"
	"database/sql"
	"errors"

	"audiodrive/internal/model"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) (*Postgres, error) {
	s := &Postgres{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Postgres) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id         BIGSERIAL PRIMARY KEY,
			raw_url    TEXT      NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (s *Postgres) Save(ctx context.Context, u model.URL) (model.URL, error) {
	const q = `INSERT INTO urls (raw_url) VALUES ($1) RETURNING id, raw_url, created_at`
	row := s.db.QueryRowContext(ctx, q, u.RawURL)
	var out model.URL
	if err := row.Scan(&out.ID, &out.RawURL, &out.CreatedAt); err != nil {
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) GetByID(ctx context.Context, id int64) (model.URL, error) {
	const q = `SELECT id, raw_url, created_at FROM urls WHERE id = $1`
	row := s.db.QueryRowContext(ctx, q, id)
	var out model.URL
	if err := row.Scan(&out.ID, &out.RawURL, &out.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) List(ctx context.Context) ([]model.URL, error) {
	const q = `SELECT id, raw_url, created_at FROM urls ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.URL
	for rows.Next() {
		var u model.URL
		if err := rows.Scan(&u.ID, &u.RawURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
