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
	UpdateStatus(ctx context.Context, id int64, status string, audioID *int64) (model.URL, error)
	ListByStatus(ctx context.Context, status string) ([]model.URL, error)
}
