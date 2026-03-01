package main

import (
	"database/sql"
	"log"
	"os"

	"audiodrive/internal/server"
	"audiodrive/internal/store"

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

	srv := server.New(":"+port, s)
	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
