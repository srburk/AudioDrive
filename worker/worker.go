package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

// Worker processes URL jobs through the pipeline in goroutines.
type Worker struct {
	cfg     Config
	store   store.URLStore
	fetcher Fetcher
	tts     Client
	sem     chan struct{} // capacity = cfg.Concurrency
}

// New creates a Worker.
func New(cfg Config, s store.URLStore, f Fetcher, t Client) *Worker {
	return &Worker{
		cfg:     cfg,
		store:   s,
		fetcher: f,
		tts:     t,
		sem:     make(chan struct{}, cfg.Concurrency),
	}
}

// Submit spawns a goroutine that processes u. Blocks on the semaphore if
// Concurrency slots are full, so the HTTP handler always returns immediately
// (the goroutine itself blocks, not the caller).
func (w *Worker) Submit(u model.URL) {
	go func() {
		w.sem <- struct{}{}
		defer func() { <-w.sem }()
		w.process(context.Background(), u)
	}()
}

func (w *Worker) process(ctx context.Context, u model.URL) {
	jlog := func(format string, args ...any) {
		log.Printf("[job %d] "+format, append([]any{u.ID}, args...)...)
	}

	jlog("started — %s", u.RawURL)
	w.store.UpdateStatus(ctx, u.ID, model.StatusProcessing, nil)

	html, err := w.fetcher.Fetch(ctx, u.RawURL)
	if err != nil {
		jlog("fetch failed: %v", err)
		w.store.UpdateStatus(ctx, u.ID, model.StatusFailed, nil)
		return
	}
	jlog("fetched — %s", formatSize(len(html)))

	text := ExtractText(html)
	title, desc := ExtractMeta(html)
	if len(text) > w.cfg.TTSMaxChars {
		text = text[:w.cfg.TTSMaxChars]
	}
	jlog("extracted — %d chars", len(text))

	if _, err := w.store.Update(ctx, u.ID, strPtr(title), strPtr(desc)); err != nil {
		jlog("meta update failed: %v", err) // non-fatal
	}

	audio, err := w.tts.Synthesize(ctx, text)
	if err != nil {
		jlog("tts failed: %v", err)
		w.store.UpdateStatus(ctx, u.ID, model.StatusFailed, nil)
		return
	}
	jlog("synthesized — %s %s", formatSize(len(audio)), w.cfg.TTSFormat)

	filename := fmt.Sprintf("%d.%s", u.ID, w.cfg.TTSFormat)
	filePath := filepath.Join(w.cfg.AudioDir, filename)
	if err := os.WriteFile(filePath, audio, 0644); err != nil {
		jlog("write audio file failed: %v", err)
		w.store.UpdateStatus(ctx, u.ID, model.StatusFailed, nil)
		return
	}

	if _, err = w.store.UpdateStatus(ctx, u.ID, model.StatusDone, &filePath); err != nil {
		jlog("final status update failed: %v", err)
		return
	}
	jlog("done — %s", filePath)
}

// strPtr returns nil for empty string, preventing overwriting user-edited fields.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
