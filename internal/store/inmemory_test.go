package store_test

import (
	"context"
	"testing"
	"time"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func newTestURL(raw string) model.URL {
	return model.URL{RawURL: raw, CreatedAt: time.Now()}
}

func TestInMemory_Save(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	u, err := s.Save(ctx, newTestURL("https://example.com"))
	if err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}
	if u.ID == 0 {
		t.Error("Save: expected non-zero ID")
	}
	if u.RawURL != "https://example.com" {
		t.Errorf("Save: RawURL = %q, want %q", u.RawURL, "https://example.com")
	}
}

func TestInMemory_GetByID(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	saved, _ := s.Save(ctx, newTestURL("https://example.com"))

	got, err := s.GetByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("GetByID: unexpected error: %v", err)
	}
	if got.ID != saved.ID {
		t.Errorf("GetByID: ID = %d, want %d", got.ID, saved.ID)
	}
}

func TestInMemory_GetByID_NotFound(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	_, err := s.GetByID(ctx, 9999)
	if err != store.ErrNotFound {
		t.Errorf("GetByID: err = %v, want store.ErrNotFound", err)
	}
}

func TestInMemory_List(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	urls, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List (empty): unexpected error: %v", err)
	}
	if len(urls) != 0 {
		t.Errorf("List (empty): got %d items, want 0", len(urls))
	}

	s.Save(ctx, newTestURL("https://first.com"))
	s.Save(ctx, newTestURL("https://second.com"))

	urls, err = s.List(ctx)
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("List: got %d items, want 2", len(urls))
	}
	// Insertion order preserved
	if urls[0].RawURL != "https://first.com" {
		t.Errorf("List: urls[0].RawURL = %q, want %q", urls[0].RawURL, "https://first.com")
	}
}
