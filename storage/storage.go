package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Metadata represents the editable fields for an audio file.
type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Store defines the interface for metadata storage.
type Store interface {
	// GetMetadata returns the metadata for a given filename.
	// If no metadata exists, it returns nil and no error.
	GetMetadata(filename string) (*Metadata, error)

	// SetMetadata saves the metadata for a given filename.
	SetMetadata(filename string, meta Metadata) error
}

// JSONStore implements Store using a single JSON file.
type JSONStore struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]Metadata
}

// NewJSONStore creates a new JSONStore.
// It loads existing data from the file if it exists.
func NewJSONStore(folder string) (*JSONStore, error) {
	filePath := filepath.Join(folder, "metadata.json")
	store := &JSONStore{
		filePath: filePath,
		data:     make(map[string]Metadata),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// load reads the JSON file into memory.
func (s *JSONStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if os.IsNotExist(err) {
		// File doesn't exist yet, that's fine.
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&s.data); err != nil {
		// If the file is empty or invalid, we might want to start fresh or return error.
		// For now, let's return error to be safe, unless it's EOF.
		if err.Error() == "EOF" {
			return nil
		}
		return fmt.Errorf("failed to decode metadata file: %w", err)
	}

	return nil
}

// save writes the memory cache to the JSON file.
func (s *JSONStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(s.data); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata for a file.
func (s *JSONStore) GetMetadata(filename string) (*Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	meta, ok := s.data[filename]
	if !ok {
		return nil, nil
	}
	return &meta, nil
}

// SetMetadata saves metadata for a file.
func (s *JSONStore) SetMetadata(filename string, meta Metadata) error {
	// Update memory
	s.mu.Lock()
	s.data[filename] = meta
	s.mu.Unlock()

	// Persist to disk
	return s.save()
}
