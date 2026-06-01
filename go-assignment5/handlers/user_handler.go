package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/yertaypert/go-assignment5/models"
	"github.com/yertaypert/go-assignment5/repository"
)

// Handler struct contains repository
type Handler struct {
	Repo *repository.Repository
}

// NewHandler constructor
func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// GetPaginatedUsers GET /users?page=1&pageSize=10&name=Alice&order_by=name
func (h *Handler) GetPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}

	// Collect filters dynamically
	filters := make(map[string]string)
	for _, field := range []string{"id", "name", "email", "gender", "birth_date"} {
		if val := r.URL.Query().Get(field); val != "" {
			filters[field] = val
		}
	}

	// Order by
	orderBy := r.URL.Query().Get("order_by")

	resp, err := h.Repo.GetPaginatedUsers(page, pageSize, filters, orderBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetCommonFriends GET /users/common-friends?user1=<uuid>&user2=<uuid>
func (h *Handler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	user1Str := r.URL.Query().Get("user1")
	user2Str := r.URL.Query().Get("user2")

	user1, err := uuid.Parse(user1Str)
	if err != nil {
		http.Error(w, "invalid user1 UUID", http.StatusBadRequest)
		return
	}

	user2, err := uuid.Parse(user2Str)
	if err != nil {
		http.Error(w, "invalid user2 UUID", http.StatusBadRequest)
		return
	}

	users, err := h.Repo.GetCommonFriends(user1, user2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Data []models.User `json:"data"`
	}{
		Data: users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
