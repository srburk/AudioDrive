package controller

import (
	"audiodrive/internal/models"
	"audiodrive/internal/repo"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type AudioObjectController struct {
	repo *repo.AudioObjectRepo
}

func NewAudioObjectController(repo *repo.AudioObjectRepo) *AudioObjectController {
	return &AudioObjectController{repo: repo}
}

func (c AudioObjectController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/api/objects/{id}", c.Get)
	mux.HandleFunc("POST /v1/api/objects", c.Create)
}

func (c *AudioObjectController) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid id")
		return
	}
	object, err := c.repo.GetById(r.Context(), id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(object)
}

func (c *AudioObjectController) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAudioObjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid JSON")
		return
	}

	if err := req.ValidateRequest(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	user, err := c.repo.Create(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to create user")

		log.Printf("Found error while creating user: %s", err.Error())

		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
