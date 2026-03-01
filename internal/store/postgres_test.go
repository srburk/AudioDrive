//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration tests")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestPostgres_SaveAndGet(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	ctx := context.Background()
	saved, err := s.Save(ctx, model.URL{RawURL: "https://integration-test.example.com"})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Error("expected non-zero ID")
	}

	got, err := s.GetByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.RawURL != saved.RawURL {
		t.Errorf("RawURL = %q, want %q", got.RawURL, saved.RawURL)
	}
}

func TestPostgres_GetByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	_, err = s.GetByID(context.Background(), -1)
	if err != store.ErrNotFound {
		t.Errorf("err = %v, want store.ErrNotFound", err)
	}
}

func TestPostgres_List(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	ctx := context.Background()
	s.Save(ctx, model.URL{RawURL: "https://list-test-a.example.com"})
	s.Save(ctx, model.URL{RawURL: "https://list-test-b.example.com"})

	urls, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(urls) < 2 {
		t.Errorf("List: got %d items, want >= 2", len(urls))
	}
}
