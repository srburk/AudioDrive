package store

import (
	"context"
	"sync"
	"time"

	"audiodrive/internal/model"
)

type InMemory struct {
	mu      sync.Mutex
	records []model.URL
	nextID  int64
}

func NewInMemory() *InMemory {
	return &InMemory{nextID: 1}
}

func (s *InMemory) Save(_ context.Context, u model.URL) (model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u.ID = s.nextID
	u.Status = model.StatusPending
	s.nextID++
	s.records = append(s.records, u)
	return u, nil
}

func (s *InMemory) GetByID(_ context.Context, id int64) (model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.records {
		if u.ID == id {
			return u, nil
		}
	}
	return model.URL{}, ErrNotFound
}

func (s *InMemory) List(_ context.Context) ([]model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]model.URL, len(s.records))
	copy(out, s.records)
	return out, nil
}

func (s *InMemory) UpdateStatus(_ context.Context, id int64, status string, audioPath *string) (model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, u := range s.records {
		if u.ID == id {
			s.records[i].Status = status
			s.records[i].AudioPath = audioPath
			return s.records[i], nil
		}
	}
	return model.URL{}, ErrNotFound
}

func (s *InMemory) ListByStatus(_ context.Context, status string) ([]model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []model.URL
	for _, u := range s.records {
		if u.Status == status {
			out = append(out, u)
		}
	}
	return out, nil
}

func (s *InMemory) ClaimPending(_ context.Context) (model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for i, u := range s.records {
		if u.Status == model.StatusPending {
			s.records[i].Status = model.StatusProcessing
			s.records[i].Attempts++
			s.records[i].LastAttemptedAt = &now
			return s.records[i], nil
		}
	}
	return model.URL{}, ErrNotFound
}

func (s *InMemory) ReapStuck(_ context.Context, threshold time.Duration, maxAttempts int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-threshold)
	n := 0
	for i, u := range s.records {
		if u.Status == model.StatusProcessing &&
			u.LastAttemptedAt != nil &&
			u.LastAttemptedAt.Before(cutoff) {
			if u.Attempts >= maxAttempts {
				s.records[i].Status = model.StatusFailed
			} else {
				s.records[i].Status = model.StatusPending
			}
			n++
		}
	}
	return n, nil
}
