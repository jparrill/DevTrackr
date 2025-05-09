package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/models"
	"github.com/jparrill/devtrackr/internal/services"
)

// PullRequestHandler handles pull request-related HTTP requests
type PullRequestHandler struct {
	trackingService *services.TrackingService
}

// NewPullRequestHandler creates a new pull request handler
func NewPullRequestHandler(trackingService *services.TrackingService) *PullRequestHandler {
	return &PullRequestHandler{
		trackingService: trackingService,
	}
}

// ListPullRequests handles GET /api/v1/issues/{key}/pull-requests
func (h *PullRequestHandler) ListPullRequests(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	prs, err := h.trackingService.ListPullRequests(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AddPullRequest handles POST /api/v1/issues/{key}/pull-requests
func (h *PullRequestHandler) AddPullRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req struct {
		Number       int    `json:"number"`
		Repository   string `json:"repository"`
		Title        string `json:"title"`
		URL          string `json:"url"`
		TargetBranch string `json:"target_branch"`
		IsBackport   bool   `json:"is_backport"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := h.trackingService.AddPullRequest(r.Context(), key, &models.PullRequest{
		Number:       req.Number,
		Repository:   req.Repository,
		Title:        req.Title,
		URL:          req.URL,
		TargetBranch: req.TargetBranch,
		IsBackport:   req.IsBackport,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pr); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdatePullRequest handles PUT /api/v1/issues/{key}/pull-requests/{number}
func (h *PullRequestHandler) UpdatePullRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		http.Error(w, "Invalid pull request number", http.StatusBadRequest)
		return
	}

	var pr models.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.trackingService.UpdatePullRequest(r.Context(), key, number, &pr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pr); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
