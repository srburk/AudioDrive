package store_test

import (
	"context"
	"errors"
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

	path := "/audio/1.mp3"
	updated, err := s.UpdateStatus(ctx, saved.ID, "done", &path)
	if err != nil {
		t.Fatalf("UpdateStatus: unexpected error: %v", err)
	}
	if updated.Status != "done" {
		t.Errorf("UpdateStatus: Status = %q, want %q", updated.Status, "done")
	}
	if updated.AudioPath == nil || *updated.AudioPath != path {
		t.Errorf("UpdateStatus: AudioPath = %v, want %q", updated.AudioPath, path)
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

func TestInMemory_Update_SetsFields(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})

	title := "My Title"
	got, err := s.Update(context.Background(), saved.ID, &title, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title == nil || *got.Title != "My Title" {
		t.Errorf("title: got %v", got.Title)
	}
	if got.Description != nil {
		t.Error("description should remain nil")
	}

	// second call: description only — title must not be cleared
	desc := "My Desc"
	got2, _ := s.Update(context.Background(), saved.ID, nil, &desc)
	if got2.Title == nil || *got2.Title != "My Title" {
		t.Error("title was overwritten")
	}
	if got2.Description == nil || *got2.Description != "My Desc" {
		t.Error("desc not set")
	}
}

func TestInMemory_Update_NotFound(t *testing.T) {
	s := store.NewInMemory()
	_, err := s.Update(context.Background(), 999, nil, nil)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestInMemory_Delete(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})
	if err := s.Delete(context.Background(), saved.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetByID(context.Background(), saved.ID); !errors.Is(err, store.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
	all, _ := s.List(context.Background())
	if len(all) != 0 {
		t.Errorf("want empty list, got %d", len(all))
	}
}

func TestInMemory_Delete_NotFound(t *testing.T) {
	s := store.NewInMemory()
	if err := s.Delete(context.Background(), 999); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
