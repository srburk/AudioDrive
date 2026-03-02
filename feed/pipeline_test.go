package feed_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"audiodrive/feed"
	"audiodrive/internal/model"
)

// ptr is a generic helper to get a pointer to a value.
func ptr[T any](v T) *T { return &v }

// doneURL creates a URL with status=done and a non-nil AudioPath.
func doneURL(id int64, audioPath string) model.URL {
	return model.URL{
		ID:        id,
		RawURL:    "https://example.com",
		Status:    model.StatusDone,
		AudioPath: &audioPath,
		CreatedAt: time.Now(),
	}
}

func TestPipeline_FiltersByStatus(t *testing.T) {
	urls := []model.URL{
		doneURL(1, "/audio/1.mp3"),
		{ID: 2, RawURL: "https://b.com", Status: model.StatusPending, CreatedAt: time.Now()},
		{ID: 3, RawURL: "https://c.com", Status: model.StatusProcessing, CreatedAt: time.Now()},
		{ID: 4, RawURL: "https://d.com", Status: model.StatusFailed, CreatedAt: time.Now()},
	}
	p := feed.New()
	items, err := p.Run(context.Background(), urls)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len = %d, want 1", len(items))
	}
	if items[0].URL.ID != 1 {
		t.Errorf("ID = %d, want 1", items[0].URL.ID)
	}
}

func TestPipeline_FilterNilAudioPath(t *testing.T) {
	urls := []model.URL{
		{ID: 1, RawURL: "https://a.com", Status: model.StatusDone, AudioPath: nil, CreatedAt: time.Now()},
	}
	p := feed.New()
	items, err := p.Run(context.Background(), urls)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestPipeline_ProcessorsRunInOrder(t *testing.T) {
	var order []int
	proc1 := func(_ context.Context, _ *feed.Item) error { order = append(order, 1); return nil }
	proc2 := func(_ context.Context, _ *feed.Item) error { order = append(order, 2); return nil }

	urls := []model.URL{doneURL(1, "/audio/1.mp3")}
	p := feed.New().Add(proc1).Add(proc2)
	_, err := p.Run(context.Background(), urls)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("order = %v, want [1 2]", order)
	}
}

func TestPipeline_StepErrorSkipsItem(t *testing.T) {
	errBoom := errors.New("boom")
	failFirst := func(_ context.Context, item *feed.Item) error {
		if item.URL.ID == 1 {
			return errBoom
		}
		return nil
	}

	urls := []model.URL{
		doneURL(1, "/audio/1.mp3"),
		doneURL(2, "/audio/2.mp3"),
	}
	p := feed.New().Add(failFirst)
	items, err := p.Run(context.Background(), urls)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len = %d, want 1 (failed item should be skipped)", len(items))
	}
	if items[0].URL.ID != 2 {
		t.Errorf("ID = %d, want 2", items[0].URL.ID)
	}
}
