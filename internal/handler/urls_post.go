package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"audiodrive/internal/model"
)

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	u := model.URL{RawURL: input.URL}
	if err := u.Validate(); err != nil {
		if errors.Is(err, model.ErrInvalidURL) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	saved, err := h.store.Save(r.Context(), u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if h.submit != nil {
		h.submit(saved)
	}

	writeJSON(w, http.StatusCreated, saved)
}
