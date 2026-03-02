package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audiodrive/internal/handler"
	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

// stubAudioStore satisfies store.AudioStore for testing.
type stubAudioStore struct {
	content  string
	mime     string
	notFound bool
}

func (s *stubAudioStore) Get(_ string) (io.ReadCloser, string, error) {
	if s.notFound {
		return nil, "", store.ErrNotFound
	}
	return io.NopCloser(strings.NewReader(s.content)), s.mime, nil
}

func (s *stubAudioStore) Delete(_ string) error { return nil }

func TestGetAudio_OK(t *testing.T) {
	audioPath := "/fake/1.mp3"
	s := newStub()
	s.saved = append(s.saved, model.URL{ID: 1, RawURL: "https://example.com", AudioPath: &audioPath})

	audio := &stubAudioStore{content: "mp3data", mime: "audio/mpeg"}
	h := handler.New(s, audio, nil)

	req := httptest.NewRequest(http.MethodGet, "/audio/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.GetAudio(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "audio/mpeg" {
		t.Errorf("Content-Type = %q, want %q", ct, "audio/mpeg")
	}
	if rr.Body.String() != "mp3data" {
		t.Errorf("body = %q, want %q", rr.Body.String(), "mp3data")
	}
}

func TestGetAudio_URLNotFound(t *testing.T) {
	h := handler.New(newStub(), &stubAudioStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/audio/999", nil)
	req.SetPathValue("id", "999")
	rr := httptest.NewRecorder()
	h.GetAudio(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetAudio_NoAudioPath(t *testing.T) {
	s := newStub()
	s.saved = append(s.saved, model.URL{ID: 2, RawURL: "https://example.com", AudioPath: nil})

	h := handler.New(s, &stubAudioStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/audio/2", nil)
	req.SetPathValue("id", "2")
	rr := httptest.NewRecorder()
	h.GetAudio(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetAudio_FileNotFound(t *testing.T) {
	audioPath := "/fake/1.mp3"
	s := newStub()
	s.saved = append(s.saved, model.URL{ID: 1, RawURL: "https://example.com", AudioPath: &audioPath})

	h := handler.New(s, &stubAudioStore{notFound: true}, nil)

	req := httptest.NewRequest(http.MethodGet, "/audio/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	h.GetAudio(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetAudio_BadID(t *testing.T) {
	h := handler.New(newStub(), &stubAudioStore{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/audio/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	h.GetAudio(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}
