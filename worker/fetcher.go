package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Fetcher fetches the raw HTML body of a URL.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) (string, error)
}

// HTTPFetcher is a Fetcher backed by a real HTTP client.
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher creates an HTTPFetcher with a timeout set by cfg.FetchTimeout.
func NewHTTPFetcher(cfg Config) *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{Timeout: cfg.FetchTimeout},
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetcher: build request: %w", err)
	}
	req.Header.Set("User-Agent", "AudioDrive/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetcher: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetcher: unexpected status %d for %s", resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetcher: read body: %w", err)
	}
	return string(body), nil
}
