package handler

import (
	"errors"
	"net/http"
	"strconv"

	"audiodrive/internal/store"
)

func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	u, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if u.AudioPath != nil && h.audioStore != nil {
		h.audioStore.Delete(*u.AudioPath) //nolint:errcheck — best-effort cleanup
	}

	w.WriteHeader(http.StatusNoContent)
}
