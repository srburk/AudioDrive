package repo

import (
	"audiodrive/internal/models"
	"context"
	"database/sql"
)

type ObjectRepo struct {
	db *sql.DB
}

func NewObjectRepo(db *sql.DB) *ObjectRepo {
	return &ObjectRepo{db: db}
}

func (r *ObjectRepo) Create(ctx context.Context, req models.CreateObjectRequest) (*models.Object, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO objects (user_id, name, url, duration_seconds)
         VALUES ($1, $2, $3, $4)
         RETURNING id, user_id, name, url, duration_seconds`,
		req.UserId, req.Name, req.URL, req.DurationSeconds,
	)
	var object models.Object
	if err := row.Scan(&object.Id, &object.UserId, &object.Name, &object.URL, &object.DurationSeconds); err != nil {
		return nil, err
	}

	return &object, nil
}

func (r *ObjectRepo) GetById(ctx context.Context, id int64) (*models.Object, error) {
	var object models.Object
	err := r.db.QueryRowContext(ctx,
		`SELECT * FROM objects WHERE id = $1`,
		id,
	).Scan(&object.Id, &object.UserId, &object.Name, &object.URL, &object.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return &object, nil
}
