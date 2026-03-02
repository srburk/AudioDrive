package feed_test

import (
	"context"
	"os"
	"testing"

	"audiodrive/feed"
	"audiodrive/internal/model"
)

func TestSizeProcessor_RealFile(t *testing.T) {
	f, err := os.CreateTemp("", "audio-*.mp3")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())

	content := []byte("hello world audio data")
	f.Write(content)
	f.Close()

	p := f.Name()
	item := &feed.Item{URL: model.URL{AudioPath: &p}}
	if err := feed.SizeProcessor(context.Background(), item); err != nil {
		t.Fatalf("SizeProcessor: %v", err)
	}
	if item.AudioSizeBytes != int64(len(content)) {
		t.Errorf("AudioSizeBytes = %d, want %d", item.AudioSizeBytes, len(content))
	}
}

func TestSizeProcessor_NilAudioPath(t *testing.T) {
	item := &feed.Item{URL: model.URL{AudioPath: nil}}
	if err := feed.SizeProcessor(context.Background(), item); err != nil {
		t.Fatalf("SizeProcessor: %v", err)
	}
	if item.AudioSizeBytes != 0 {
		t.Errorf("AudioSizeBytes = %d, want 0", item.AudioSizeBytes)
	}
}

func TestSizeProcessor_MissingFile(t *testing.T) {
	path := "/nonexistent/path/does-not-exist-audio.mp3"
	item := &feed.Item{URL: model.URL{AudioPath: &path}}
	if err := feed.SizeProcessor(context.Background(), item); err != nil {
		t.Fatalf("SizeProcessor: %v", err)
	}
	if item.AudioSizeBytes != 0 {
		t.Errorf("AudioSizeBytes = %d, want 0", item.AudioSizeBytes)
	}
}
