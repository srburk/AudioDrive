//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"errors"
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

func TestPostgres_Update_SetsFields(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	ctx := context.Background()
	saved, _ := s.Save(ctx, model.URL{RawURL: "https://update-test.example.com"})

	title := "My Title"
	got, err := s.Update(ctx, saved.ID, &title, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got.Title == nil || *got.Title != "My Title" {
		t.Errorf("title: got %v", got.Title)
	}
	if got.Description != nil {
		t.Error("description should remain nil")
	}

	desc := "My Desc"
	got2, _ := s.Update(ctx, saved.ID, nil, &desc)
	if got2.Title == nil || *got2.Title != "My Title" {
		t.Error("title was overwritten")
	}
	if got2.Description == nil || *got2.Description != "My Desc" {
		t.Error("desc not set")
	}
}

func TestPostgres_Update_NotFound(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	_, err = s.Update(context.Background(), -1, nil, nil)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestPostgres_Delete(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	ctx := context.Background()
	saved, _ := s.Save(ctx, model.URL{RawURL: "https://delete-test.example.com"})
	if err := s.Delete(ctx, saved.ID); err != nil {
		t.Fatal(err)
	}
	_, err = s.GetByID(ctx, saved.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestPostgres_Delete_NotFound(t *testing.T) {
	db := openTestDB(t)
	s, err := store.NewPostgres(db)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}

	if err := s.Delete(context.Background(), -1); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
