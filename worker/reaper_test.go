package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
	"audiodrive/worker"
)

type countingStore struct {
	reaped atomic.Int64
}

func (s *countingStore) Save(_ context.Context, u model.URL) (model.URL, error) {
	return u, nil
}
func (s *countingStore) GetByID(_ context.Context, _ int64) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *countingStore) List(_ context.Context) ([]model.URL, error) { return nil, nil }
func (s *countingStore) UpdateStatus(_ context.Context, _ int64, _ string, _ *string) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *countingStore) ListByStatus(_ context.Context, _ string) ([]model.URL, error) {
	return nil, nil
}
func (s *countingStore) ClaimPending(_ context.Context) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *countingStore) ReapStuck(_ context.Context, _ time.Duration, _ int) (int, error) {
	s.reaped.Add(1)
	return 0, nil
}

func TestReaper_CallsReapStuck(t *testing.T) {
	s := &countingStore{}
	cfg := worker.Config{
		ReaperInterval: 20 * time.Millisecond,
		StuckThreshold: 5 * time.Minute,
		MaxAttempts:    3,
	}
	r := worker.NewReaper(s, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r.Run(ctx)

	if s.reaped.Load() == 0 {
		t.Error("Reaper.Run: expected at least one ReapStuck call")
	}
}
