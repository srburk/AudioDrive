package server

import (
	"net/http"

	"audiodrive/internal/handler"
	"audiodrive/internal/store"
)

func New(addr string, s store.URLStore) *http.Server {
	h := handler.New(s)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /urls", h.CreateURL)
	mux.HandleFunc("GET /urls/{id}", h.GetURL)
	mux.HandleFunc("GET /urls", h.ListURLs)
	mux.HandleFunc("PATCH /urls/{id}", h.UpdateURL)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
