package store

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AudioStore retrieves synthesized audio files by their stored path.
type AudioStore interface {
	// Get opens the file at path and returns a ReadCloser and MIME type.
	// Returns ErrNotFound if the file does not exist.
	Get(path string) (io.ReadCloser, string, error)
}

var mimeTypes = map[string]string{
	"mp3":  "audio/mpeg",
	"opus": "audio/ogg",
	"aac":  "audio/aac",
	"flac": "audio/flac",
	"wav":  "audio/wav",
}

// DiskAudioStore opens audio files directly from the filesystem path stored in the DB.
type DiskAudioStore struct{}

func NewDiskAudioStore() *DiskAudioStore {
	return &DiskAudioStore{}
}

func (s *DiskAudioStore) Get(path string) (io.ReadCloser, string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	mime, ok := mimeTypes[ext]
	if !ok {
		mime = "application/octet-stream"
	}
	return f, mime, nil
}
