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

func TestInMemory_Save_DefaultStatus(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	u, _ := s.Save(ctx, newTestURL("https://example.com"))
	if u.Status != "pending" {
		t.Errorf("Save: Status = %q, want %q", u.Status, "pending")
	}
}

func TestInMemory_UpdateStatus(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	saved, _ := s.Save(ctx, newTestURL("https://example.com"))

	audioID := int64(42)
	updated, err := s.UpdateStatus(ctx, saved.ID, "done", &audioID)
	if err != nil {
		t.Fatalf("UpdateStatus: unexpected error: %v", err)
	}
	if updated.Status != "done" {
		t.Errorf("UpdateStatus: Status = %q, want %q", updated.Status, "done")
	}
	if updated.AudioID == nil || *updated.AudioID != 42 {
		t.Errorf("UpdateStatus: AudioID = %v, want 42", updated.AudioID)
	}
}

func TestInMemory_UpdateStatus_NotFound(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	_, err := s.UpdateStatus(ctx, 9999, "done", nil)
	if err != store.ErrNotFound {
		t.Errorf("UpdateStatus: err = %v, want store.ErrNotFound", err)
	}
}

func TestInMemory_ListByStatus(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	u1, _ := s.Save(ctx, newTestURL("https://a.com"))
	u2, _ := s.Save(ctx, newTestURL("https://b.com"))
	s.Save(ctx, newTestURL("https://c.com"))

	s.UpdateStatus(ctx, u1.ID, "processing", nil)
	s.UpdateStatus(ctx, u2.ID, "processing", nil)

	results, err := s.ListByStatus(ctx, "processing")
	if err != nil {
		t.Fatalf("ListByStatus: unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ListByStatus: got %d items, want 2", len(results))
	}

	pending, _ := s.ListByStatus(ctx, "pending")
	if len(pending) != 1 {
		t.Errorf("ListByStatus pending: got %d items, want 1", len(pending))
	}
}
