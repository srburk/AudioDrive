package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"audiodrive/internal/model"
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

	ext := strings.TrimPrefix(filepath.Ext(*u.AudioPath), ".")
	name := sanitizeFilename(titleOrID(u))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.%s"`, name, ext))
	w.Header().Set("Content-Type", mimeType)
	io.Copy(w, rc)
}

func titleOrID(u model.URL) string {
	if u.Title != nil && *u.Title != "" {
		return *u.Title
	}
	return strconv.FormatInt(u.ID, 10)
}

func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			b.WriteRune(r)
		}
	}
	result := strings.Join(strings.Fields(b.String()), "-")
	if len(result) > 100 {
		result = result[:100]
	}
	return result
}
