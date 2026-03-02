package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"audiodrive/internal/handler"
	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

// stubStore satisfies store.URLStore for testing.
type stubStore struct {
	saved  []model.URL
	nextID int64
}

func newStub() *stubStore { return &stubStore{nextID: 1} }

func (s *stubStore) Save(_ context.Context, u model.URL) (model.URL, error) {
	u.ID = s.nextID
	u.Status = model.StatusPending
	u.CreatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.nextID++
	s.saved = append(s.saved, u)
	return u, nil
}

func (s *stubStore) GetByID(_ context.Context, id int64) (model.URL, error) {
	for _, u := range s.saved {
		if u.ID == id {
			return u, nil
		}
	}
	return model.URL{}, store.ErrNotFound
}

func (s *stubStore) List(_ context.Context) ([]model.URL, error) {
	out := make([]model.URL, len(s.saved))
	copy(out, s.saved)
	return out, nil
}

func (s *stubStore) UpdateStatus(_ context.Context, id int64, status string, audioPath *string) (model.URL, error) {
	for i, u := range s.saved {
		if u.ID == id {
			s.saved[i].Status = status
			s.saved[i].AudioPath = audioPath
			return s.saved[i], nil
		}
	}
	return model.URL{}, store.ErrNotFound
}

func (s *stubStore) Update(_ context.Context, id int64, title, description *string) (model.URL, error) {
	for i, u := range s.saved {
		if u.ID != id {
			continue
		}
		if title != nil {
			s.saved[i].Title = title
		}
		if description != nil {
			s.saved[i].Description = description
		}
		return s.saved[i], nil
	}
	return model.URL{}, store.ErrNotFound
}

func (s *stubStore) Delete(_ context.Context, id int64) error {
	for i, u := range s.saved {
		if u.ID == id {
			s.saved = append(s.saved[:i], s.saved[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

// helpers

func postURL(h *handler.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/urls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateURL(rr, req)
	return rr
}

func getURL(h *handler.Handler, path string, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.GetURL(rr, req)
	return rr
}

// --- POST /urls ---

func TestCreateURL_Created(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	rr := postURL(h, `{"url":"https://example.com"}`)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	var got model.URL
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.RawURL != "https://example.com" {
		t.Errorf("RawURL = %q, want %q", got.RawURL, "https://example.com")
	}
	if got.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateURL_InvalidJSON(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	rr := postURL(h, `not json`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestCreateURL_InvalidURL(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	rr := postURL(h, `{"url":"not-a-url"}`)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
}

// --- GET /urls/{id} ---

func TestGetURL_OK(t *testing.T) {
	s := newStub()
	h := handler.New(s, nil, nil)
	postURL(h, `{"url":"https://example.com"}`)

	req := httptest.NewRequest(http.MethodGet, "/urls/1", nil)
	req.SetPathValue("id", "1")
	rr := getURL(h, "/urls/1", req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var got model.URL
	json.NewDecoder(rr.Body).Decode(&got)
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
}

func TestGetURL_NotFound(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/urls/999", nil)
	req.SetPathValue("id", "999")
	rr := getURL(h, "/urls/999", req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestGetURL_BadID(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/urls/abc", nil)
	req.SetPathValue("id", "abc")
	rr := getURL(h, "/urls/abc", req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// --- GET /urls ---

func TestListURLs_Empty(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/urls", nil)
	rr := httptest.NewRecorder()
	h.ListURLs(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var got []model.URL
	json.NewDecoder(rr.Body).Decode(&got)
	if got == nil {
		t.Error("expected empty array, not null")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestListURLs_NonEmpty(t *testing.T) {
	s := newStub()
	h := handler.New(s, nil, nil)
	postURL(h, `{"url":"https://a.com"}`)
	postURL(h, `{"url":"https://b.com"}`)

	req := httptest.NewRequest(http.MethodGet, "/urls", nil)
	rr := httptest.NewRecorder()
	h.ListURLs(rr, req)

	var got []model.URL
	json.NewDecoder(rr.Body).Decode(&got)
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// --- PATCH /urls/{id} ---

func TestPatchURL_UpdatesTitle(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})
	h := handler.New(s, nil, nil)

	body := fmt.Sprintf(`{"title":"New Title"}`)
	req := httptest.NewRequest(http.MethodPatch, "/urls/"+strconv.FormatInt(saved.ID, 10), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", strconv.FormatInt(saved.ID, 10))
	rr := httptest.NewRecorder()
	h.PatchURL(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var got model.URL
	json.NewDecoder(rr.Body).Decode(&got)
	if got.Title == nil || *got.Title != "New Title" {
		t.Errorf("title = %v", got.Title)
	}
}

func TestPatchURL_NotFound(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	req := httptest.NewRequest(http.MethodPatch, "/urls/999", strings.NewReader(`{"title":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "999")
	rr := httptest.NewRecorder()
	h.PatchURL(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

// --- DELETE /urls/{id} ---

func TestDeleteURL_NoContent(t *testing.T) {
	s := store.NewInMemory()
	saved, _ := s.Save(context.Background(), model.URL{RawURL: "https://a.com"})
	h := handler.New(s, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/urls/"+strconv.FormatInt(saved.ID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(saved.ID, 10))
	rr := httptest.NewRecorder()
	h.DeleteURL(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
	_, err := s.GetByID(context.Background(), saved.ID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Error("row should be gone")
	}
}

func TestDeleteURL_NotFound(t *testing.T) {
	h := handler.New(newStub(), nil, nil)
	req := httptest.NewRequest(http.MethodDelete, "/urls/999", nil)
	req.SetPathValue("id", "999")
	rr := httptest.NewRecorder()
	h.DeleteURL(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}
