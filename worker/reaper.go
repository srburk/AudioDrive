package worker

import (
	"context"
	"log"
	"time"

	"audiodrive/internal/store"
)

// Reaper periodically resets stuck processing jobs back to pending (or failed).
type Reaper struct {
	store store.URLStore
	cfg   Config
}

// NewReaper creates a Reaper.
func NewReaper(s store.URLStore, cfg Config) *Reaper {
	return &Reaper{store: s, cfg: cfg}
}

// Run runs the reaper loop until ctx is cancelled.
func (r *Reaper) Run(ctx context.Context) {
	t := time.NewTicker(r.cfg.ReaperInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := r.store.ReapStuck(ctx, r.cfg.StuckThreshold, r.cfg.MaxAttempts)
			if err != nil {
				log.Printf("reaper: %v", err)
			}
			if n > 0 {
				log.Printf("reaper: reset %d stuck job(s)", n)
			}
		}
	}
}
