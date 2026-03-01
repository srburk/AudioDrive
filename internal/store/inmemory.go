package store

import (
	"context"
	"sync"

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
