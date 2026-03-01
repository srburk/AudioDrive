package worker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"audiodrive/worker"
)

func newFetcherCfg(timeout time.Duration) worker.Config {
	return worker.Config{FetchTimeout: timeout}
}

func TestHTTPFetcher_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "AudioDrive/1.0" {
			t.Errorf("User-Agent = %q, want AudioDrive/1.0", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Hello</body></html>"))
	}))
	defer ts.Close()

	f := worker.NewHTTPFetcher(newFetcherCfg(5 * time.Second))
	body, err := f.Fetch(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("Fetch: unexpected error: %v", err)
	}
	if body != "<html><body>Hello</body></html>" {
		t.Errorf("body = %q, want html content", body)
	}
}

func TestHTTPFetcher_NonSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	f := worker.NewHTTPFetcher(newFetcherCfg(5 * time.Second))
	_, err := f.Fetch(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("Fetch: expected error for 404, got nil")
	}
}

func TestHTTPFetcher_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// block forever
		<-r.Context().Done()
	}))
	defer ts.Close()

	f := worker.NewHTTPFetcher(newFetcherCfg(50 * time.Millisecond))
	ctx := context.Background()
	_, err := f.Fetch(ctx, ts.URL)
	if err == nil {
		t.Fatal("Fetch: expected timeout error, got nil")
	}
}
