package server

import (
	"net/http"

	"audiodrive/internal/handler"
	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func New(addr string, s store.URLStore, a store.AudioStore, submit func(model.URL)) *http.Server {
	h := handler.New(s, a, submit)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /urls", h.CreateURL)
	mux.HandleFunc("GET /urls/{id}", h.GetURL)
	mux.HandleFunc("GET /urls", h.ListURLs)
	mux.HandleFunc("GET /audio/{id}", h.GetAudio)
	mux.HandleFunc("PATCH /urls/{id}", h.PatchURL)
	mux.HandleFunc("DELETE /urls/{id}", h.DeleteURL)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
