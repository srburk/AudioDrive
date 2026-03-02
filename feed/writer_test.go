package feed_test

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"audiodrive/feed"
	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func TestWriter_Rebuild_WritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feed.xml")

	s := store.NewInMemory()
	audioPath := filepath.Join(dir, "audio.mp3")
	os.WriteFile(audioPath, []byte("fake audio data"), 0644)

	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com", CreatedAt: time.Now()})
	s.UpdateStatus(context.Background(), saved.ID, model.StatusDone, &audioPath)

	p := feed.New().Add(feed.SizeProcessor)
	ch := feed.Channel{Title: "Test", BaseURL: "https://host", Description: "Test feed", Language: "en"}
	w := feed.NewWriter(s, p, ch, path)

	if err := w.Rebuild(context.Background()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var rss struct {
		XMLName xml.Name `xml:"rss"`
	}
	if err := xml.Unmarshal(data, &rss); err != nil {
		t.Fatalf("invalid XML: %v\n%s", err, data)
	}
}

func TestWriter_Rebuild_ReplacesOnSecondCall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feed.xml")

	s := store.NewInMemory()
	p := feed.New()
	ch := feed.Channel{Title: "Test", BaseURL: "https://host", Description: "d", Language: "en"}
	w := feed.NewWriter(s, p, ch, path)

	if err := w.Rebuild(context.Background()); err != nil {
		t.Fatalf("first Rebuild: %v", err)
	}
	if err := w.Rebuild(context.Background()); err != nil {
		t.Fatalf("second Rebuild: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing after second rebuild: %v", err)
	}
}

func TestWriter_Rebuild_StoreError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "feed.xml")

	s := &failingListStore{}
	p := feed.New()
	ch := feed.Channel{Title: "Test", BaseURL: "https://host", Description: "d", Language: "en"}
	w := feed.NewWriter(s, p, ch, path)

	if err := w.Rebuild(context.Background()); err == nil {
		t.Fatal("expected error from store")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("no file should be written on store error")
	}
}

// failingListStore satisfies store.URLStore but always fails on List.
type failingListStore struct{ store.URLStore }

func (f *failingListStore) List(_ context.Context) ([]model.URL, error) {
	return nil, errors.New("store error")
}
