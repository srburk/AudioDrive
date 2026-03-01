package store

import (
	"context"
	"errors"
	"time"

	"audiodrive/internal/model"
)

var ErrNotFound = errors.New("record not found")

type URLStore interface {
	Save(ctx context.Context, u model.URL) (model.URL, error)
	GetByID(ctx context.Context, id int64) (model.URL, error)
	List(ctx context.Context) ([]model.URL, error)
	UpdateStatus(ctx context.Context, id int64, status string, audioPath *string) (model.URL, error)
	ListByStatus(ctx context.Context, status string) ([]model.URL, error)
	// ClaimPending atomically selects the oldest pending job, marks it processing,
	// increments attempts, and sets last_attempted_at. Returns ErrNotFound when queue empty.
	ClaimPending(ctx context.Context) (model.URL, error)
	// ReapStuck resets processing jobs idle longer than threshold back to pending,
	// or to failed if attempts >= maxAttempts. Returns the count of rows updated.
	ReapStuck(ctx context.Context, threshold time.Duration, maxAttempts int) (int, error)
}
