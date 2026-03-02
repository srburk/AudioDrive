package worker

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the worker.
type Config struct {
	DatabaseURL  string
	Concurrency  int
	FetchTimeout time.Duration
	TTSEndpoint  string
	TTSAPIKey    string
	TTSModel     string
	TTSVoice     string
	TTSFormat    string
	TTSMaxChars  int
	AudioDir     string
}

// FromEnv builds a Config from environment variables, applying defaults and
// fatally logging on missing required values.
func FromEnv() Config {
	cfg := Config{
		DatabaseURL:  requireEnv("DATABASE_URL"),
		TTSEndpoint:  requireEnv("TTS_ENDPOINT"),
		TTSAPIKey:    requireEnv("TTS_API_KEY"),
		AudioDir:     requireEnv("AUDIO_DIR"),
		Concurrency:  parseInt("CONCURRENCY", 1),
		FetchTimeout: parseDuration("FETCH_TIMEOUT", 30*time.Second),
		TTSModel:     envOr("TTS_MODEL", "tts-1"),
		TTSVoice:     envOr("TTS_VOICE", "alloy"),
		TTSFormat:    envOr("TTS_FORMAT", "mp3"),
		TTSMaxChars:  parseInt("TTS_MAX_CHARS", 4096),
	}
	return cfg
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("worker: required env var %s is not set", key)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("worker: invalid duration for %s=%q: %v", key, v, err)
	}
	return d
}

func parseInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("worker: invalid int for %s=%q: %v", key, v, err)
	}
	return n
}
