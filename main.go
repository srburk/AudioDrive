package main

import (
	"context"
	"database/sql"
	"log"
	"os"

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

	cfg := worker.FromEnv()
	w := worker.New(cfg, s, worker.NewHTTPFetcher(cfg), worker.NewOpenAIClient(cfg))

	audioStore := store.NewDiskAudioStore()

	srv := server.New(":"+port, s, audioStore, w.Submit)
	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
