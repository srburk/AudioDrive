package worker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
	"audiodrive/worker"
)

// --- stub store ---

type stubStore struct {
	mu      sync.Mutex
	updates []string // status values received
	done    chan struct{}
	closed  bool
}

func newStubStore() *stubStore {
	return &stubStore{done: make(chan struct{})}
}

func (s *stubStore) UpdateStatus(_ context.Context, _ int64, status string, _ *string) (model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updates = append(s.updates, status)
	if !s.closed && (status == model.StatusDone || status == model.StatusFailed) {
		s.closed = true
		close(s.done)
	}
	return model.URL{}, nil
}

func (s *stubStore) Update(_ context.Context, _ int64, _, _ *string) (model.URL, error) {
	return model.URL{}, nil
}

func (s *stubStore) Save(_ context.Context, u model.URL) (model.URL, error) { return u, nil }
func (s *stubStore) GetByID(_ context.Context, _ int64) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *stubStore) List(_ context.Context) ([]model.URL, error)   { return nil, nil }
func (s *stubStore) Delete(_ context.Context, _ int64) error       { return nil }

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

// --- tests ---

func TestSubmit_HappyPath(t *testing.T) {
	dir := t.TempDir()

	ttsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("fake-audio"))
	}))
	defer ttsServer.Close()

	stub := newStubStore()
	cfg := worker.Config{
		TTSEndpoint: ttsServer.URL,
		TTSAPIKey:   "key",
		TTSModel:    "tts-1",
		TTSVoice:    "alloy",
		TTSFormat:   "mp3",
		TTSMaxChars: 4096,
		Concurrency: 1,
		AudioDir:    dir,
	}
	f := &stubFetcher{html: "<html><head><title>Test</title></head><body><p>Article text</p></body></html>"}
	w := worker.New(cfg, stub, f, worker.NewOpenAIClient(cfg))
	w.Submit(model.URL{ID: 1, RawURL: "https://example.com"})

	select {
	case <-stub.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if last := stub.updates[len(stub.updates)-1]; last != "done" {
		t.Errorf("final status = %q, want done", last)
	}
}

func TestSubmit_FetchError_MarksFailedImmediately(t *testing.T) {
	stub := newStubStore()
	cfg := worker.Config{Concurrency: 1, TTSFormat: "mp3", TTSMaxChars: 4096, AudioDir: t.TempDir()}
	f := &stubFetcher{err: errFetch}
	w := worker.New(cfg, stub, f, nil)
	w.Submit(model.URL{ID: 2, RawURL: "https://example.com"})

	select {
	case <-stub.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if last := stub.updates[len(stub.updates)-1]; last != "failed" {
		t.Errorf("final status = %q, want failed", last)
	}
}

var errFetch = &fetchError{"connection refused"}

type fetchError struct{ msg string }

func (e *fetchError) Error() string { return e.msg }

func TestSubmit_SemaphoreLimitsConcurrency(t *testing.T) {
	const jobs = 5
	const maxConcurrency = 2

	var active, peak int64
	var wg sync.WaitGroup

	// A fetcher that tracks concurrent calls
	blockCh := make(chan struct{})
	concurrentFetcher := &countingFetcher{
		active: &active,
		peak:   &peak,
		wg:     &wg,
		block:  blockCh,
	}
	// Unblock after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(blockCh)
	}()

	stub := &multiDoneStore{count: jobs, done: make(chan struct{})}
	cfg := worker.Config{Concurrency: maxConcurrency, TTSFormat: "mp3", TTSMaxChars: 4096, AudioDir: t.TempDir()}
	w := worker.New(cfg, stub, concurrentFetcher, &stubTTS{err: errFetch})

	wg.Add(jobs)
	for i := 0; i < jobs; i++ {
		w.Submit(model.URL{ID: int64(i + 1), RawURL: "https://example.com"})
	}

	select {
	case <-stub.done:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for all jobs")
	}

	if peak > maxConcurrency {
		t.Errorf("peak concurrency %d > %d", peak, maxConcurrency)
	}
}

type countingFetcher struct {
	active *int64
	peak   *int64
	wg     *sync.WaitGroup
	block  chan struct{}
}

func (f *countingFetcher) Fetch(_ context.Context, _ string) (string, error) {
	curr := atomic.AddInt64(f.active, 1)
	for {
		p := atomic.LoadInt64(f.peak)
		if curr <= p || atomic.CompareAndSwapInt64(f.peak, p, curr) {
			break
		}
	}
	<-f.block
	atomic.AddInt64(f.active, -1)
	f.wg.Done()
	return "", errFetch
}

type multiDoneStore struct {
	mu      sync.Mutex
	count   int
	done    chan struct{}
	closed  bool
}

func (s *multiDoneStore) UpdateStatus(_ context.Context, _ int64, status string, _ *string) (model.URL, error) {
	if status == model.StatusFailed || status == model.StatusDone {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.count--
		if s.count <= 0 && !s.closed {
			s.closed = true
			close(s.done)
		}
	}
	return model.URL{}, nil
}
func (s *multiDoneStore) Update(_ context.Context, _ int64, _, _ *string) (model.URL, error) {
	return model.URL{}, nil
}
func (s *multiDoneStore) Save(_ context.Context, u model.URL) (model.URL, error) { return u, nil }
func (s *multiDoneStore) GetByID(_ context.Context, _ int64) (model.URL, error) {
	return model.URL{}, store.ErrNotFound
}
func (s *multiDoneStore) List(_ context.Context) ([]model.URL, error) { return nil, nil }
func (s *multiDoneStore) Delete(_ context.Context, _ int64) error     { return nil }

func TestSubmit_AudioFileWritten(t *testing.T) {
	dir := t.TempDir()

	ttsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("audio-content"))
	}))
	defer ttsServer.Close()

	stub := newStubStore()
	cfg := worker.Config{
		TTSEndpoint: ttsServer.URL,
		TTSAPIKey:   "key",
		TTSModel:    "tts-1",
		TTSVoice:    "alloy",
		TTSFormat:   "mp3",
		TTSMaxChars: 4096,
		Concurrency: 1,
		AudioDir:    dir,
	}
	f := &stubFetcher{html: "<html><body>some text</body></html>"}
	w := worker.New(cfg, stub, f, worker.NewOpenAIClient(cfg))
	w.Submit(model.URL{ID: 5, RawURL: "https://example.com"})

	select {
	case <-stub.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
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
