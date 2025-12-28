package storage

import (
	"os"
	"testing"
)

func TestJSONStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewJSONStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Test Get on empty
	meta, err := store.GetMetadata("foo.mp3")
	if err != nil {
		t.Errorf("GetMetadata returned error on empty: %v", err)
	}
	if meta != nil {
		t.Errorf("GetMetadata returned non-nil on empty")
	}

	// Test Set
	newMeta := Metadata{Title: "My Title", Description: "Desc"}
	if err := store.SetMetadata("foo.mp3", newMeta); err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	// Test Get
	meta, err = store.GetMetadata("foo.mp3")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if meta == nil {
		t.Fatal("GetMetadata returned nil")
	}
	if meta.Title != "My Title" {
		t.Errorf("Expected title 'My Title', got '%s'", meta.Title)
	}

	// Test Persistence (Reload Store)
	store2, err := NewJSONStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to reload store: %v", err)
	}
	meta2, err := store2.GetMetadata("foo.mp3")
	if err != nil || meta2 == nil {
		t.Fatalf("Failed to retrieve persistent data")
	}
	if meta2.Title != "My Title" {
		t.Errorf("Expected persisted title 'My Title', got '%s'", meta2.Title)
	}

	// Test Delete
	if err := store2.DeleteMetadata("foo.mp3"); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}

	meta3, err := store2.GetMetadata("foo.mp3")
	if err != nil {
		t.Errorf("GetMetadata returned error after delete: %v", err)
	}
	if meta3 != nil {
		t.Errorf("Metadata should be nil after delete")
	}

	// Verify persistence of deletion
	store3, err := NewJSONStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to reload store 3: %v", err)
	}
	meta4, err := store3.GetMetadata("foo.mp3")
	if meta4 != nil {
		t.Errorf("Metadata should be nil after delete and reload")
	}
}
