package models

import "errors"

type Object struct {
	Id              int64
	UserId          int64
	Name            string
	URL             string
	DurationSeconds int
}

type CreateObjectRequest struct {
	UserId          int64  `json:"user_id"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}

func (req CreateObjectRequest) ValidateRequest() error {
	if req.Name == "" {
		return errors.New("email is required")
	}
	if req.URL == "" {
		return errors.New("URL is required")
	}
	return nil
}
