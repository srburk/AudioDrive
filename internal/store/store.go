package store

import (
	"context"
	"errors"

	"audiodrive/internal/model"
)

var ErrNotFound = errors.New("record not found")

type URLStore interface {
	Save(ctx context.Context, u model.URL) (model.URL, error)
	GetByID(ctx context.Context, id int64) (model.URL, error)
	List(ctx context.Context) ([]model.URL, error)
	UpdateStatus(ctx context.Context, id int64, status string, audioPath *string) (model.URL, error)
	// Update sets title and/or description; nil pointer = leave field unchanged.
	Update(ctx context.Context, id int64, title, description *string) (model.URL, error)
	// Delete permanently removes the row. Returns ErrNotFound if missing.
	Delete(ctx context.Context, id int64) error
}
