package worker_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"audiodrive/worker"
)

func newTTSClient(endpoint string) *worker.OpenAIClient {
	cfg := worker.Config{
		TTSEndpoint: endpoint,
		TTSAPIKey:   "test-key",
		TTSModel:    "tts-1",
		TTSVoice:    "alloy",
		TTSFormat:   "mp3",
	}
	return worker.NewOpenAIClient(cfg)
}

func TestOpenAIClient_Synthesize_Success(t *testing.T) {
	fakeAudio := []byte("fake-mp3-bytes")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request shape
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", r.Header.Get("Authorization"))
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["model"] != "tts-1" {
			t.Errorf("model = %q, want tts-1", body["model"])
		}
		if body["voice"] != "alloy" {
			t.Errorf("voice = %q, want alloy", body["voice"])
		}
		if body["response_format"] != "mp3" {
			t.Errorf("response_format = %q, want mp3", body["response_format"])
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		w.Write(fakeAudio)
	}))
	defer ts.Close()

	client := newTTSClient(ts.URL)
	audio, err := client.Synthesize(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("Synthesize: unexpected error: %v", err)
	}
	if string(audio) != string(fakeAudio) {
		t.Errorf("audio = %q, want %q", audio, fakeAudio)
	}
}

func TestOpenAIClient_Synthesize_200WithJSONError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"error":{"message":"invalid api key","type":"invalid_request_error"}}`))
	}))
	defer ts.Close()

	client := newTTSClient(ts.URL)
	_, err := client.Synthesize(context.Background(), "Hello")
	if err == nil {
		t.Fatal("Synthesize: expected error for 200+JSON body, got nil")
	}
}

func TestOpenAIClient_Synthesize_NonOKError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer ts.Close()

	client := newTTSClient(ts.URL)
	_, err := client.Synthesize(context.Background(), "Hello")
	if err == nil {
		t.Fatal("Synthesize: expected error for 401, got nil")
	}
}

func TestOpenAIClient_Synthesize_RequestShape(t *testing.T) {
	var captured map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("audio"))
	}))
	defer ts.Close()

	client := newTTSClient(ts.URL)
	client.Synthesize(context.Background(), "test input")

	if captured["input"] != "test input" {
		t.Errorf("input = %q, want %q", captured["input"], "test input")
	}
}
