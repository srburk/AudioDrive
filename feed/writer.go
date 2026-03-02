package feed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"audiodrive/internal/store"
)

// Writer queries the store, runs the pipeline, and writes feed.xml atomically.
type Writer struct {
	store    store.URLStore
	pipeline *Pipeline
	channel  Channel
	path     string
	mu       sync.Mutex
}

// NewWriter constructs a Writer.
func NewWriter(s store.URLStore, p *Pipeline, ch Channel, path string) *Writer {
	return &Writer{store: s, pipeline: p, channel: ch, path: path}
}

// Rebuild queries the store, runs the pipeline, and writes the feed file
// atomically via a temp file + rename. If a rebuild is already in progress,
// this call returns nil immediately (de-duplicated via TryLock).
func (w *Writer) Rebuild(ctx context.Context) error {
	if !w.mu.TryLock() {
		return nil // another rebuild is already running
	}
	defer w.mu.Unlock()

	urls, err := w.store.List(ctx)
	if err != nil {
		return fmt.Errorf("feed: list urls: %w", err)
	}

	items, err := w.pipeline.Run(ctx, urls)
	if err != nil {
		return fmt.Errorf("feed: pipeline: %w", err)
	}

	data, err := Build(w.channel, items)
	if err != nil {
		return fmt.Errorf("feed: build: %w", err)
	}

	dir := filepath.Dir(w.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("feed: mkdir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "feed-*.xml")
	if err != nil {
		return fmt.Errorf("feed: create temp: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("feed: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("feed: close temp: %w", err)
	}

	if err := os.Rename(tmpName, w.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("feed: rename: %w", err)
	}

	return nil
}
