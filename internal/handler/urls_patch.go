package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"audiodrive/internal/model"
	"audiodrive/internal/store"
)

func (h *Handler) UpdateURL(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be an integer")
		return
	}

	var body struct {
		Status  string `json:"status"`
		AudioID *int64 `json:"audio_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	switch body.Status {
	case model.StatusPending, model.StatusProcessing, model.StatusDone, model.StatusFailed:
	default:
		writeError(w, http.StatusUnprocessableEntity, "status must be one of: pending, processing, done, failed")
		return
	}

	u, err := h.store.UpdateStatus(r.Context(), id, body.Status, body.AudioID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, u)
}
