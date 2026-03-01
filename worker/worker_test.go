package worker_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
	"audiodrive/worker"
)

// --- stub store ---

type stubStore struct {
	pending  []model.URL
	updates  []statusUpdate
	claimErr error
}

type statusUpdate struct {
	id        int64
	status    string
	audioPath *string
}

func (s *stubStore) ClaimPending(_ context.Context) (model.URL, error) {
	if s.claimErr != nil {
		return model.URL{}, s.claimErr
	}
	if len(s.pending) == 0 {
		return model.URL{}, store.ErrNotFound
	}
	u := s.pending[0]
	s.pending = s.pending[1:]
	u.Status = model.StatusProcessing
	u.Attempts++
	return u, nil
}

func (s *stubStore) UpdateStatus(_ context.Context, id int64, status string, audioPath *string) (model.URL, error) {
	s.updates = append(s.updates, statusUpdate{id, status, audioPath})
	return model.URL{ID: id, Status: status, AudioPath: audioPath}, nil
}

func (s *stubStore) Save(_ context.Context, u model.URL) (model.URL, error) { return u, nil }
func (s *stubStore) GetByID(_ context.Context, _ int64) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *stubStore) List(_ context.Context) ([]model.URL, error)              { return nil, nil }
func (s *stubStore) ListByStatus(_ context.Context, _ string) ([]model.URL, error) {
	return nil, nil
}
func (s *stubStore) ReapStuck(_ context.Context, _ time.Duration, _ int) (int, error) {
	return 0, nil
}

// --- stub fetcher ---

type stubFetcher struct {
	html string
	err  error
}

func (f *stubFetcher) Fetch(_ context.Context, _ string) (string, error) {
	return f.html, f.err
}

// --- stub TTS ---

type stubTTS struct {
	audio []byte
	err   error
}

func (t *stubTTS) Synthesize(_ context.Context, _ string) ([]byte, error) {
	return t.audio, t.err
}

// --- helpers ---

func newTestWorker(s *stubStore, f worker.Fetcher, t worker.Client, audioDir string) *worker.Worker {
	cfg := worker.Config{
		TTSFormat:   "mp3",
		TTSMaxChars: 4096,
		MaxAttempts: 3,
		AudioDir:    audioDir,
	}
	return worker.New(cfg, s, f, t)
}

// --- tests ---

func TestProcessOne_HappyPath(t *testing.T) {
	dir := t.TempDir()

	ttsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-audio"))
	}))
	defer ttsServer.Close()

	s := &stubStore{
		pending: []model.URL{{ID: 1, RawURL: "https://example.com", Status: model.StatusPending, Attempts: 0}},
	}
	f := &stubFetcher{html: "<html><body><p>Article text</p></body></html>"}
	cfg := worker.Config{
		TTSEndpoint: ttsServer.URL,
		TTSAPIKey:   "key",
		TTSModel:    "tts-1",
		TTSVoice:    "alloy",
		TTSFormat:   "mp3",
		TTSMaxChars: 4096,
		MaxAttempts: 3,
		AudioDir:    dir,
	}
	tts := worker.NewOpenAIClient(cfg)
	w := worker.New(cfg, s, f, tts)

	err := w.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne: unexpected error: %v", err)
	}
	if len(s.updates) == 0 {
		t.Fatal("expected UpdateStatus to be called")
	}
	last := s.updates[len(s.updates)-1]
	if last.status != model.StatusDone {
		t.Errorf("status = %q, want done", last.status)
	}
	if last.audioPath == nil {
		t.Error("audioPath should not be nil on success")
	}
}

func TestProcessOne_EmptyQueue_ReturnsNotFound(t *testing.T) {
	dir := t.TempDir()
	s := &stubStore{}
	w := newTestWorker(s, &stubFetcher{}, &stubTTS{}, dir)

	err := w.ProcessOne(context.Background())
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("ProcessOne: err = %v, want ErrNotFound", err)
	}
}

func TestProcessOne_FetchError_Retry(t *testing.T) {
	dir := t.TempDir()
	s := &stubStore{
		pending: []model.URL{{ID: 1, RawURL: "https://example.com", Status: model.StatusPending, Attempts: 0}},
	}
	f := &stubFetcher{err: errors.New("connection refused")}

	w := newTestWorker(s, f, &stubTTS{}, dir)
	err := w.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne: unexpected error: %v", err)
	}
	if len(s.updates) == 0 {
		t.Fatal("expected UpdateStatus to be called after fetch error")
	}
	// attempts=1, maxAttempts=3 → should requeue as pending
	if s.updates[0].status != model.StatusPending {
		t.Errorf("status = %q, want pending (retry)", s.updates[0].status)
	}
}

func TestProcessOne_FetchError_MaxAttempts_Fail(t *testing.T) {
	dir := t.TempDir()
	s := &stubStore{
		pending: []model.URL{{ID: 1, RawURL: "https://example.com", Status: model.StatusProcessing, Attempts: 3}},
	}
	f := &stubFetcher{err: errors.New("timeout")}

	w := newTestWorker(s, f, &stubTTS{}, dir)
	err := w.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne: unexpected error: %v", err)
	}
	if len(s.updates) == 0 {
		t.Fatal("expected UpdateStatus to be called")
	}
	if s.updates[0].status != model.StatusFailed {
		t.Errorf("status = %q, want failed (exhausted)", s.updates[0].status)
	}
}

func TestProcessOne_TTSError_Retry(t *testing.T) {
	dir := t.TempDir()
	s := &stubStore{
		pending: []model.URL{{ID: 2, RawURL: "https://example.com", Status: model.StatusProcessing, Attempts: 1}},
	}
	f := &stubFetcher{html: "<html><body>text</body></html>"}
	tts := &stubTTS{err: errors.New("tts unavailable")}

	w := newTestWorker(s, f, tts, dir)
	err := w.ProcessOne(context.Background())
	if err != nil {
		t.Fatalf("ProcessOne: unexpected error: %v", err)
	}
	if len(s.updates) == 0 {
		t.Fatal("expected UpdateStatus to be called after TTS error")
	}
	// attempts=1, maxAttempts=3 → pending
	if s.updates[0].status != model.StatusPending {
		t.Errorf("status = %q, want pending (retry after TTS error)", s.updates[0].status)
	}
}

func TestProcessOne_AudioFileWritten(t *testing.T) {
	dir := t.TempDir()

	ttsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("audio-content"))
	}))
	defer ttsServer.Close()

	s := &stubStore{
		pending: []model.URL{{ID: 5, RawURL: "https://example.com", Status: model.StatusPending, Attempts: 0}},
	}
	f := &stubFetcher{html: "<html><body>some text</body></html>"}
	cfg := worker.Config{
		TTSEndpoint: ttsServer.URL,
		TTSAPIKey:   "key",
		TTSModel:    "tts-1",
		TTSVoice:    "alloy",
		TTSFormat:   "mp3",
		TTSMaxChars: 4096,
		MaxAttempts: 3,
		AudioDir:    dir,
	}
	tts := worker.NewOpenAIClient(cfg)
	w := worker.New(cfg, s, f, tts)

	if err := w.ProcessOne(context.Background()); err != nil {
		t.Fatalf("ProcessOne: unexpected error: %v", err)
	}

	expected := dir + "/5.mp3"
	data, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("audio file not written to %s: %v", expected, err)
	}
	if string(data) != "audio-content" {
		t.Errorf("audio content = %q, want audio-content", string(data))
	}
}
