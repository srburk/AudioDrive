package store_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"audiodrive/internal/store"
)

func TestDiskAudioStore_Get_Found(t *testing.T) {
	dir := t.TempDir()
	content := []byte("fake mp3 bytes")
	p := filepath.Join(dir, "42.mp3")
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}

	s := store.NewDiskAudioStore()
	rc, mime, err := s.Get(p)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer rc.Close()

	if mime != "audio/mpeg" {
		t.Errorf("mime = %q, want %q", mime, "audio/mpeg")
	}
	got, _ := io.ReadAll(rc)
	if string(got) != string(content) {
		t.Errorf("body = %q, want %q", got, content)
	}
}

func TestDiskAudioStore_Get_NotFound(t *testing.T) {
	s := store.NewDiskAudioStore()
	_, _, err := s.Get(filepath.Join(t.TempDir(), "missing.mp3"))
	if err != store.ErrNotFound {
		t.Errorf("err = %v, want store.ErrNotFound", err)
	}
}

func TestDiskAudioStore_Get_MimeType(t *testing.T) {
	cases := []struct {
		format string
		want   string
	}{
		{"mp3", "audio/mpeg"},
		{"opus", "audio/ogg"},
		{"aac", "audio/aac"},
		{"flac", "audio/flac"},
		{"wav", "audio/wav"},
		{"xyz", "application/octet-stream"},
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "1."+tc.format)
			if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
			s := store.NewDiskAudioStore()
			rc, mime, err := s.Get(p)
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			rc.Close()
			if mime != tc.want {
				t.Errorf("format=%q: mime = %q, want %q", tc.format, mime, tc.want)
			}
		})
	}
}
