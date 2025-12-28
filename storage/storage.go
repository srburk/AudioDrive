package storage

import (
	"crypto/rand"
	"encoding/hex"
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

// User represents a user with access to the feed.
type User struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

// Store defines the interface for metadata storage.
type Store interface {
	// Metadata Ops
	GetMetadata(filename string) (*Metadata, error)
	SetMetadata(filename string, meta Metadata) error
	DeleteMetadata(filename string) error

	// Token Ops
	AddUser(name string) (*User, error)
	RemoveUser(token string) error
	ListUsers() ([]User, error)
	ValidateToken(token string) bool
}

// JSONStore implements Store using JSON files.
type JSONStore struct {
	mu           sync.RWMutex
	metaFilePath string
	userFilePath string
	metadata     map[string]Metadata
	users        map[string]User // Keyed by Token
}

// NewJSONStore creates a new JSONStore.
func NewJSONStore(folder string) (*JSONStore, error) {
	store := &JSONStore{
		metaFilePath: filepath.Join(folder, "metadata.json"),
		userFilePath: filepath.Join(folder, "users.json"),
		metadata:     make(map[string]Metadata),
		users:        make(map[string]User),
	}

	if err := store.loadMetadata(); err != nil {
		return nil, err
	}
	if err := store.loadUsers(); err != nil {
		return nil, err
	}

	return store, nil
}

// --- Metadata Logic ---

func (s *JSONStore) loadMetadata() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.metaFilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&s.metadata); err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		return err
	}
	return nil
}

func (s *JSONStore) saveMetadata() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.metaFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.metadata)
}

func (s *JSONStore) GetMetadata(filename string) (*Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meta, ok := s.metadata[filename]
	if !ok {
		return nil, nil
	}
	return &meta, nil
}

func (s *JSONStore) SetMetadata(filename string, meta Metadata) error {
	s.mu.Lock()
	s.metadata[filename] = meta
	s.mu.Unlock()
	return s.saveMetadata()
}

func (s *JSONStore) DeleteMetadata(filename string) error {
	s.mu.Lock()
	delete(s.metadata, filename)
	s.mu.Unlock()
	return s.saveMetadata()
}

// --- User/Token Logic ---

func (s *JSONStore) loadUsers() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.userFilePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open users file: %w", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&s.users); err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		return err
	}
	return nil
}

func (s *JSONStore) saveUsers() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.userFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(s.users)
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *JSONStore) AddUser(name string) (*User, error) {
	token := generateToken()
	user := User{Name: name, Token: token}

	s.mu.Lock()
	s.users[token] = user
	s.mu.Unlock()

	if err := s.saveUsers(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *JSONStore) RemoveUser(token string) error {
	s.mu.Lock()
	delete(s.users, token)
	s.mu.Unlock()
	return s.saveUsers()
}

func (s *JSONStore) ListUsers() ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []User
	for _, u := range s.users {
		list = append(list, u)
	}
	return list, nil
}

func (s *JSONStore) ValidateToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.users[token]
	return exists
}
