package handler

import (
	"io"
	"net/http"
	"strconv"

	"audiodrive/internal/store"
)

// GetAudio serves the synthesized audio file for a URL by its ID.
// GET /audio/{id}
func (h *Handler) GetAudio(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	u, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if u.AudioPath == nil {
		writeError(w, http.StatusNotFound, "audio not available")
		return
	}

	rc, mimeType, err := h.audioStore.Get(*u.AudioPath)
	if err != nil {
		if err == store.ErrNotFound {
			writeError(w, http.StatusNotFound, "audio file not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", mimeType)
	io.Copy(w, rc)
}
