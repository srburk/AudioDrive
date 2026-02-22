package repo

import (
	"audiodrive/internal/models"
	"context"
	"database/sql"
)

type AudioObjectRepo struct {
	db *sql.DB
}

func NewAudioObjectRepo(db *sql.DB) *AudioObjectRepo {
	return &AudioObjectRepo{db: db}
}

func (r *AudioObjectRepo) Create(ctx context.Context, req models.CreateAudioObjectRequest) (*models.AudioObject, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO objects (user_id, name, url, duration_seconds)
         VALUES ($1, $2, $3, $4)
         RETURNING id, user_id, name, url, duration_seconds`,
		req.UserId, req.Name, req.URL, req.DurationSeconds,
	)
	var object models.AudioObject
	if err := row.Scan(&object.Id, &object.UserId, &object.Name, &object.URL, &object.DurationSeconds); err != nil {
		return nil, err
	}

	return &object, nil
}

func (r *AudioObjectRepo) GetById(ctx context.Context, id int64) (*models.AudioObject, error) {
	var object models.AudioObject
	err := r.db.QueryRowContext(ctx,
		`SELECT * FROM objects WHERE id = $1`,
		id,
	).Scan(&object.Id, &object.UserId, &object.Name, &object.URL, &object.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return &object, nil
}
