package model

import (
	"errors"
	"net/url"
	"time"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

type URL struct {
	ID        int64     `json:"id"`
	RawURL    string    `json:"url"`
	Status    string    `json:"status"`
	AudioID   *int64    `json:"audio_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrInvalidURL = errors.New("invalid url: must be absolute with http or https scheme")

func (u *URL) Validate() error {
	parsed, err := url.ParseRequestURI(u.RawURL)
	if err != nil {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	if parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
