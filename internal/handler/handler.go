package handler

import (
	"encoding/json"
	"net/http"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

type Handler struct {
	store      store.URLStore
	audioStore store.AudioStore
	submit     func(model.URL) // nil-safe; nil in tests
}

func New(s store.URLStore, a store.AudioStore, submit func(model.URL)) *Handler {
	return &Handler{store: s, audioStore: a, submit: submit}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
