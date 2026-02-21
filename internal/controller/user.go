package controller

import (
	"audiodrive/internal/repo"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type UserController struct {
	repo *repo.UserRepo
}

func NewUserController(repo *repo.UserRepo) *UserController {
	return &UserController{repo: repo}
}

func (c UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/api/users/{id}", c.Get)
	mux.HandleFunc("POST /v1/api/users", c.Create)
}

func (c *UserController) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid id")
		return
	}
	user, err := c.repo.GetById(r.Context(), id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

type CreateUserRequest struct {
	Email string `json:"email"`
}

func (req CreateUserRequest) validate() error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("invalid JSON")
		return
	}

	if err := req.validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	user, err := c.repo.Create(r.Context(), req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to create user")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
