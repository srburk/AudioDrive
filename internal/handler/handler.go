package handler

import (
	"encoding/json"
	"net/http"

	"audiodrive/internal/store"
)

type Handler struct {
	store      store.URLStore
	audioStore store.AudioStore
}

func New(s store.URLStore, a store.AudioStore) *Handler {
	return &Handler{store: s, audioStore: a}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
