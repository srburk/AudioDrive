package server

import (
	"net/http"

	"audiodrive/internal/handler"
	"audiodrive/internal/store"
)

func New(addr string, s store.URLStore, a store.AudioStore) *http.Server {
	h := handler.New(s, a)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /urls", h.CreateURL)
	mux.HandleFunc("GET /urls/{id}", h.GetURL)
	mux.HandleFunc("GET /urls", h.ListURLs)
	mux.HandleFunc("GET /audio/{id}", h.GetAudio)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
