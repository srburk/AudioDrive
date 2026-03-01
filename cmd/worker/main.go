package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"audiodrive/internal/store"
	"audiodrive/worker"
)

func main() {
	cfg := worker.FromEnv()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("worker: open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("worker: ping db: %v", err)
	}

	if err := os.MkdirAll(cfg.AudioDir, 0755); err != nil {
		log.Fatalf("worker: create audio dir: %v", err)
	}

	pg, err := store.NewPostgres(db)
	if err != nil {
		log.Fatalf("worker: migrate: %v", err)
	}

	w := worker.New(cfg, pg, worker.NewHTTPFetcher(cfg), worker.NewOpenAIClient(cfg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker: starting (concurrency=%d, poll=%s)", cfg.Concurrency, cfg.PollInterval)
	w.Run(ctx)
	log.Printf("worker: shutdown complete")
}
