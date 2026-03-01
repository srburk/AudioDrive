package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client synthesizes text into audio bytes.
type Client interface {
	Synthesize(ctx context.Context, text string) ([]byte, error)
}

// OpenAIClient implements Client against an OpenAI-compatible TTS API.
type OpenAIClient struct {
	endpoint   string
	apiKey     string
	model      string
	voice      string
	format     string
	httpClient *http.Client
}

// NewOpenAIClient creates an OpenAIClient from Config.
func NewOpenAIClient(cfg Config) *OpenAIClient {
	return &OpenAIClient{
		endpoint:   cfg.TTSEndpoint,
		apiKey:     cfg.TTSAPIKey,
		model:      cfg.TTSModel,
		voice:      cfg.TTSVoice,
		format:     cfg.TTSFormat,
		httpClient: &http.Client{},
	}
}

func (c *OpenAIClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	payload := struct {
		Model          string `json:"model"`
		Input          string `json:"input"`
		Voice          string `json:"voice"`
		ResponseFormat string `json:"response_format"`
	}{
		Model:          c.model,
		Input:          text,
		Voice:          c.voice,
		ResponseFormat: c.format,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("tts: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("tts: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tts: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts: API returned status %d: %s", resp.StatusCode, string(errBody))
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "audio/") {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts: unexpected content-type %q: %s", ct, string(errBody))
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tts: read response: %w", err)
	}
	return audio, nil
}
