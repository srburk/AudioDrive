package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

const scanCols = `id, raw_url, status, audio_path, attempts, last_attempted_at, created_at`

func scanURL(row interface{ Scan(...any) error }) (model.URL, error) {
	var u model.URL
	err := row.Scan(&u.ID, &u.RawURL, &u.Status, &u.AudioPath, &u.Attempts, &u.LastAttemptedAt, &u.CreatedAt)
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
	q := `UPDATE urls SET status = $2, audio_path = $3 WHERE id = $1 RETURNING ` + scanCols
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

func (s *Postgres) ListByStatus(ctx context.Context, status string) ([]model.URL, error) {
	q := `SELECT ` + scanCols + ` FROM urls WHERE status = $1 ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, q, status)
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

func (s *Postgres) ClaimPending(ctx context.Context) (model.URL, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.URL{}, err
	}
	defer tx.Rollback()

	selectQ := `SELECT id FROM urls WHERE status = 'pending' ORDER BY id ASC LIMIT 1 FOR UPDATE SKIP LOCKED`
	row := tx.QueryRowContext(ctx, selectQ)
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}
		return model.URL{}, err
	}

	updateQ := `UPDATE urls SET status = 'processing', attempts = attempts + 1, last_attempted_at = NOW()
		WHERE id = $1 RETURNING ` + scanCols
	out, err := scanURL(tx.QueryRowContext(ctx, updateQ, id))
	if err != nil {
		return model.URL{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.URL{}, err
	}
	return out, nil
}

func (s *Postgres) ReapStuck(ctx context.Context, threshold time.Duration, maxAttempts int) (int, error) {
	q := `UPDATE urls
		SET status = CASE WHEN attempts >= $2 THEN 'failed' ELSE 'pending' END
		WHERE status = 'processing'
		  AND last_attempted_at < NOW() - make_interval(secs => $1)`
	res, err := s.db.ExecContext(ctx, q, threshold.Seconds(), maxAttempts)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}
