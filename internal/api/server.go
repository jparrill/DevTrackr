package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/services"
)

// Server represents the API server
type Server struct {
	router          *mux.Router
	trackingService *services.TrackingService
}

// NewServer creates a new API server
func NewServer(trackingService *services.TrackingService) *Server {
	s := &Server{
		router:          mux.NewRouter(),
		trackingService: trackingService,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	// API v1 routes
	v1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Issue routes
	v1.HandleFunc("/issues", s.listIssues).Methods("GET")
	v1.HandleFunc("/issues", s.createIssue).Methods("POST")
	v1.HandleFunc("/issues/{key}", s.getIssue).Methods("GET")
	v1.HandleFunc("/issues/{key}", s.deleteIssue).Methods("DELETE")
	v1.HandleFunc("/issues/{key}/polling-interval", s.updatePollingInterval).Methods("PUT")
}

// Start starts the API server
func (s *Server) Start(addr string) error {
	fmt.Printf("API server starting on %s\n", addr)
	return http.ListenAndServe(addr, s.router)
}

// listIssues returns all tracked issues
func (s *Server) listIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := s.trackingService.ListIssues(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issues)
}

// createIssue creates a new tracked issue
func (s *Server) createIssue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JiraURL string `json:"jira_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	issue, err := s.trackingService.TrackIssue(r.Context(), req.JiraURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(issue)
}

// getIssue returns a specific issue
func (s *Server) getIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	issue, err := s.trackingService.GetIssue(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

// deleteIssue deletes a tracked issue
func (s *Server) deleteIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if err := s.trackingService.DeleteIssue(r.Context(), key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updatePollingInterval updates the polling interval for an issue
func (s *Server) updatePollingInterval(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req struct {
		PollingInterval int `json:"polling_interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PollingInterval < 0 {
		http.Error(w, "Polling interval must be non-negative", http.StatusBadRequest)
		return
	}

	if err := s.trackingService.UpdateIssuePollingInterval(r.Context(), key, req.PollingInterval); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
