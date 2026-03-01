package store_test

import (
	"context"
	"sync"
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

func TestInMemory_ClaimPending_ReturnFirstPending(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	s.Save(ctx, newTestURL("https://a.com"))
	s.Save(ctx, newTestURL("https://b.com"))

	u, err := s.ClaimPending(ctx)
	if err != nil {
		t.Fatalf("ClaimPending: unexpected error: %v", err)
	}
	if u.Status != model.StatusProcessing {
		t.Errorf("ClaimPending: Status = %q, want %q", u.Status, model.StatusProcessing)
	}
	if u.Attempts != 1 {
		t.Errorf("ClaimPending: Attempts = %d, want 1", u.Attempts)
	}
	if u.LastAttemptedAt == nil {
		t.Error("ClaimPending: LastAttemptedAt should not be nil")
	}
	// Should claim the first one (lowest ID)
	if u.RawURL != "https://a.com" {
		t.Errorf("ClaimPending: RawURL = %q, want https://a.com", u.RawURL)
	}
}

func TestInMemory_ClaimPending_SkipsProcessing(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	u1, _ := s.Save(ctx, newTestURL("https://a.com"))
	s.Save(ctx, newTestURL("https://b.com"))

	// Mark first as processing manually
	s.UpdateStatus(ctx, u1.ID, model.StatusProcessing, nil)

	// ClaimPending should return the second one
	u, err := s.ClaimPending(ctx)
	if err != nil {
		t.Fatalf("ClaimPending: unexpected error: %v", err)
	}
	if u.RawURL != "https://b.com" {
		t.Errorf("ClaimPending: RawURL = %q, want https://b.com", u.RawURL)
	}
}

func TestInMemory_ClaimPending_EmptyQueueReturnsNotFound(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	_, err := s.ClaimPending(ctx)
	if err != store.ErrNotFound {
		t.Errorf("ClaimPending: err = %v, want store.ErrNotFound", err)
	}
}

func TestInMemory_ClaimPending_Concurrent(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	s.Save(ctx, newTestURL("https://a.com"))
	s.Save(ctx, newTestURL("https://b.com"))

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		claimed []int64
	)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := s.ClaimPending(ctx)
			if err == nil {
				mu.Lock()
				claimed = append(claimed, u.ID)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(claimed) != 2 {
		t.Errorf("concurrent ClaimPending: got %d claims, want 2", len(claimed))
	}
	// No duplicates
	seen := map[int64]bool{}
	for _, id := range claimed {
		if seen[id] {
			t.Errorf("duplicate claim for ID %d", id)
		}
		seen[id] = true
	}
}

func TestInMemory_ReapStuck_ResetsOldProcessing(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	s.Save(ctx, newTestURL("https://a.com"))
	u, _ := s.ClaimPending(ctx)

	// Backdate last_attempted_at by manipulating via a second claim + the returned value's age
	// Reap with threshold=0 to catch everything
	n, err := s.ReapStuck(ctx, 0, 3)
	if err != nil {
		t.Fatalf("ReapStuck: unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("ReapStuck: n = %d, want 1", n)
	}

	got, _ := s.GetByID(ctx, u.ID)
	if got.Status != model.StatusPending {
		t.Errorf("ReapStuck: Status = %q, want pending", got.Status)
	}
}

func TestInMemory_ReapStuck_MarksFailedWhenExhausted(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	s.Save(ctx, newTestURL("https://a.com"))

	// Claim 3 times to exhaust maxAttempts=3, then reap
	for i := 0; i < 3; i++ {
		u, _ := s.ClaimPending(ctx)
		// Reset back to pending except last iteration
		if i < 2 {
			s.UpdateStatus(ctx, u.ID, model.StatusPending, nil)
		}
	}

	// Now it's processing with attempts=3; reap with threshold=0, maxAttempts=3
	n, err := s.ReapStuck(ctx, 0, 3)
	if err != nil {
		t.Fatalf("ReapStuck: unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("ReapStuck: n = %d, want 1", n)
	}

	got, _ := s.GetByID(ctx, 1)
	if got.Status != model.StatusFailed {
		t.Errorf("ReapStuck: Status = %q, want failed", got.Status)
	}
}

func TestInMemory_ReapStuck_IgnoresRecentProcessing(t *testing.T) {
	s := store.NewInMemory()
	ctx := context.Background()

	s.Save(ctx, newTestURL("https://a.com"))
	s.ClaimPending(ctx) // just claimed, recent

	// threshold=5min: should NOT reap a job just claimed
	n, err := s.ReapStuck(ctx, 5*time.Minute, 3)
	if err != nil {
		t.Fatalf("ReapStuck: unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("ReapStuck: n = %d, want 0 (job is recent)", n)
	}
}
