package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"audiodrive/feed"
	"audiodrive/internal/server"
	"audiodrive/internal/store"
	"audiodrive/worker"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatalf("DATABASE_URL environment variable is required")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatalf("BASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	s, err := store.NewPostgres(db)
	if err != nil {
		log.Fatalf("store.NewPostgres: %v", err)
	}

	// Reset any jobs stuck in "processing" from a previous run.
	if res, err := db.ExecContext(context.Background(), `UPDATE urls SET status='failed' WHERE status='processing'`); err == nil {
		if n, _ := res.RowsAffected(); n > 0 {
			log.Printf("startup: marked %d stuck job(s) as failed", n)
		}
	}

	// Build the feed pipeline and writer.
	p := feed.New().Add(feed.SizeProcessor)
	ch := feed.Channel{
		Title:       "AudioDrive",
		BaseURL:     baseURL,
		Description: "Web pages converted to audio",
		Language:    "en",
	}
	fw := feed.NewWriter(s, p, ch, "web/feed.xml")

	// Initial feed build at startup.
	if err := fw.Rebuild(context.Background()); err != nil {
		log.Printf("feed: initial rebuild failed: %v", err)
	}

	// Wrap the store so mutations trigger a feed rebuild.
	wrappedStore := feed.NewNotifyingStore(s, func() {
		if err := fw.Rebuild(context.Background()); err != nil {
			log.Printf("feed: rebuild failed: %v", err)
		}
	})

	cfg := worker.FromEnv()
	w := worker.New(cfg, wrappedStore, worker.NewHTTPFetcher(cfg), worker.NewOpenAIClient(cfg))

	audioStore := store.NewDiskAudioStore()

	srv := server.New(":"+port, wrappedStore, audioStore, w.Submit)
	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
