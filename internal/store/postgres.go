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
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id         BIGSERIAL PRIMARY KEY,
			raw_url    TEXT        NOT NULL,
			status     TEXT        NOT NULL DEFAULT 'pending'
			             CHECK (status IN ('pending','processing','done','failed')),
			audio_path TEXT        NULL,
			attempts   INT         NOT NULL DEFAULT 0,
			last_attempted_at TIMESTAMPTZ NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending'
			CHECK (status IN ('pending','processing','done','failed'))`,
		`ALTER TABLE urls DROP COLUMN IF EXISTS audio_id`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS audio_path TEXT NULL`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS last_attempted_at TIMESTAMPTZ NULL`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS title TEXT NULL`,
		`ALTER TABLE urls ADD COLUMN IF NOT EXISTS description TEXT NULL`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

const scanCols = `id, raw_url, status, audio_path, title, description, attempts, last_attempted_at, created_at`

func scanURL(row interface{ Scan(...any) error }) (model.URL, error) {
	var u model.URL
	err := row.Scan(&u.ID, &u.RawURL, &u.Status, &u.AudioPath, &u.Title, &u.Description, &u.Attempts, &u.LastAttemptedAt, &u.CreatedAt)
	return u, err
}

func (s *Postgres) Save(ctx context.Context, u model.URL) (model.URL, error) {
	q := `INSERT INTO urls (raw_url) VALUES ($1) RETURNING ` + scanCols
	row := s.db.QueryRowContext(ctx, q, u.RawURL)
	out, err := scanURL(row)
	if err != nil {
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) GetByID(ctx context.Context, id int64) (model.URL, error) {
	q := `SELECT ` + scanCols + ` FROM urls WHERE id = $1`
	row := s.db.QueryRowContext(ctx, q, id)
	out, err := scanURL(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) List(ctx context.Context) ([]model.URL, error) {
	q := `SELECT ` + scanCols + ` FROM urls ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.URL
	for rows.Next() {
		u, err := scanURL(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Postgres) UpdateStatus(ctx context.Context, id int64, status string, audioPath *string) (model.URL, error) {
	q := `UPDATE urls SET status = $2, audio_path = COALESCE($3, audio_path) WHERE id = $1 RETURNING ` + scanCols
	row := s.db.QueryRowContext(ctx, q, id, status, audioPath)
	out, err := scanURL(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) Update(ctx context.Context, id int64, title, description *string) (model.URL, error) {
	q := `UPDATE urls SET title = COALESCE($2, title), description = COALESCE($3, description) WHERE id = $1 RETURNING ` + scanCols
	row := s.db.QueryRowContext(ctx, q, id, title, description)
	out, err := scanURL(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM urls WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
