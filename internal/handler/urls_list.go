package handler

import (
	"net/http"

	"audiodrive/internal/model"
)

func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	urls, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Return empty array, not null
	if urls == nil {
		urls = []model.URL{}
	}

	writeJSON(w, http.StatusOK, urls)
}
