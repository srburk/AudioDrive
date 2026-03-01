package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

// Worker claims pending URL jobs and processes them through the pipeline.
type Worker struct {
	cfg     Config
	store   store.URLStore
	fetcher Fetcher
	tts     Client
}

// New creates a Worker.
func New(cfg Config, s store.URLStore, f Fetcher, t Client) *Worker {
	return &Worker{cfg: cfg, store: s, fetcher: f, tts: t}
}

// Run starts Concurrency worker goroutines and a reaper, then blocks until ctx is done.
func (w *Worker) Run(ctx context.Context) {
	for i := 0; i < w.cfg.Concurrency; i++ {
		go w.loop(ctx)
	}
	if w.cfg.ReaperInterval > 0 {
		reaper := NewReaper(w.store, w.cfg)
		go reaper.Run(ctx)
	}
	<-ctx.Done()
}

func (w *Worker) loop(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		err := w.processOne(ctx)
		if errors.Is(err, store.ErrNotFound) {
			select {
			case <-time.After(w.cfg.PollInterval):
			case <-ctx.Done():
				return
			}
		} else if err != nil {
			log.Printf("worker: unexpected error: %v", err)
			time.Sleep(time.Second)
		}
	}
}

// ProcessOne claims one pending job and runs it through the full pipeline.
// Returns store.ErrNotFound when the queue is empty.
func (w *Worker) ProcessOne(ctx context.Context) error {
	return w.processOne(ctx)
}

func (w *Worker) processOne(ctx context.Context) error {
	u, err := w.store.ClaimPending(ctx)
	if err != nil {
		return err
	}

	jlog := func(format string, args ...any) {
		log.Printf("[job %d] "+format, append([]any{u.ID}, args...)...)
	}

	jlog("claimed — %s (attempt %d/%d)", u.RawURL, u.Attempts, w.cfg.MaxAttempts)

	html, err := w.fetcher.Fetch(ctx, u.RawURL)
	if err != nil {
		jlog("fetch failed: %v", err)
		return w.markFailed(ctx, u, jlog)
	}
	jlog("fetched — %s", formatSize(len(html)))

	text := ExtractText(html)
	if len(text) > w.cfg.TTSMaxChars {
		text = text[:w.cfg.TTSMaxChars]
	}
	jlog("extracted — %d chars", len(text))

	audio, err := w.tts.Synthesize(ctx, text)
	if err != nil {
		jlog("tts failed: %v", err)
		return w.markFailed(ctx, u, jlog)
	}
	jlog("synthesized — %s %s", formatSize(len(audio)), w.cfg.TTSFormat)

	filename := fmt.Sprintf("%d.%s", u.ID, w.cfg.TTSFormat)
	filePath := filepath.Join(w.cfg.AudioDir, filename)
	if err := os.WriteFile(filePath, audio, 0644); err != nil {
		return fmt.Errorf("write audio file: %w", err)
	}

	if _, err = w.store.UpdateStatus(ctx, u.ID, model.StatusDone, &filePath); err != nil {
		return err
	}
	jlog("done — %s", filePath)
	return nil
}

func (w *Worker) markFailed(ctx context.Context, u model.URL, jlog func(string, ...any)) error {
	if u.Attempts >= w.cfg.MaxAttempts {
		jlog("giving up after %d attempts", u.Attempts)
		_, err := w.store.UpdateStatus(ctx, u.ID, model.StatusFailed, nil)
		return err
	}
	jlog("will retry (attempt %d/%d)", u.Attempts, w.cfg.MaxAttempts)
	_, err := w.store.UpdateStatus(ctx, u.ID, model.StatusPending, nil)
	return err
}

func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
