package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/services"
)

// IssueHandler handles HTTP requests for issues
type IssueHandler struct {
	trackingService *services.TrackingService
}

// NewIssueHandler creates a new issue handler
func NewIssueHandler(trackingService *services.TrackingService) *IssueHandler {
	return &IssueHandler{
		trackingService: trackingService,
	}
}

// ListIssues handles GET /api/v1/issues
func (h *IssueHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := h.trackingService.ListIssues(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(issues); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetIssue handles GET /api/v1/issues/{key}
func (h *IssueHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	issue, err := h.trackingService.GetIssue(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if issue == nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(issue); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SubscribeToIssue handles POST /api/v1/issues/{key}/subscribe
func (h *IssueHandler) SubscribeToIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Get user ID from request context or header
	userID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sub, err := h.trackingService.SubscribeToIssue(r.Context(), key, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UnsubscribeFromIssue handles DELETE /api/v1/issues/{key}/unsubscribe
func (h *IssueHandler) UnsubscribeFromIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Get user ID from request context or header
	userID, err := strconv.ParseInt(r.Header.Get("X-User-ID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check for unmerged pull requests
	hasUnmerged, err := h.trackingService.HasUnmergedPullRequests(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if hasUnmerged {
		http.Error(w, "Cannot unsubscribe: there are unmerged pull requests", http.StatusConflict)
		return
	}

	if err := h.trackingService.UnsubscribeFromIssue(r.Context(), key, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TrackIssue handles POST /api/v1/issues
func (h *IssueHandler) TrackIssue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JiraURL string `json:"jira_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.JiraURL == "" {
		http.Error(w, "Jira URL is required", http.StatusBadRequest)
		return
	}

	issue, err := h.trackingService.TrackIssue(r.Context(), req.JiraURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(issue); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteIssue handles DELETE requests to remove a tracked issue
func (h *IssueHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		http.Error(w, "Issue key is required", http.StatusBadRequest)
		return
	}

	if err := h.trackingService.DeleteIssue(r.Context(), key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateIssueStatus handles PUT /api/v1/issues/{key}/status
func (h *IssueHandler) UpdateIssueStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	// Get the issue first
	issue, err := h.trackingService.GetIssue(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if issue == nil {
		http.Error(w, "Issue not found", http.StatusNotFound)
		return
	}

	// Parse the request body
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	// Update the issue status
	if err := h.trackingService.UpdateIssueStatus(r.Context(), issue, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
