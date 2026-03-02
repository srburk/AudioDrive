package feed_test

import (
	"context"
	"testing"
	"time"

	"audiodrive/feed"
	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func TestNotifyingStore_UpdateStatus_Notifies(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	path := "/audio/1.mp3"
	if _, err := ns.UpdateStatus(context.Background(), saved.ID, model.StatusDone, &path); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("notify was not called within timeout")
	}
}

func TestNotifyingStore_Update_Notifies(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	title := "New Title"
	if _, err := ns.Update(context.Background(), saved.ID, &title, nil); err != nil {
		t.Fatalf("Update: %v", err)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("notify was not called within timeout")
	}
}

func TestNotifyingStore_Delete_Notifies(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	if err := ns.Delete(context.Background(), saved.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("notify was not called within timeout")
	}
}

func TestNotifyingStore_UpdateStatus_ErrorNoNotify(t *testing.T) {
	s := store.NewInMemory()

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	_, err := ns.UpdateStatus(context.Background(), 9999, model.StatusDone, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	select {
	case <-done:
		t.Error("notify should not be called on error")
	case <-time.After(20 * time.Millisecond):
		// good — notify was not called
	}
}

func TestNotifyingStore_Update_ErrorNoNotify(t *testing.T) {
	s := store.NewInMemory()

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	title := "x"
	_, err := ns.Update(context.Background(), 9999, &title, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	select {
	case <-done:
		t.Error("notify should not be called on error")
	case <-time.After(20 * time.Millisecond):
		// good
	}
}

func TestNotifyingStore_Delete_ErrorNoNotify(t *testing.T) {
	s := store.NewInMemory()

	done := make(chan struct{}, 1)
	ns := feed.NewNotifyingStore(s, func() { done <- struct{}{} })

	err := ns.Delete(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error")
	}

	select {
	case <-done:
		t.Error("notify should not be called on error")
	case <-time.After(20 * time.Millisecond):
		// good
	}
}
