package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/api/handlers"
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
	// Create handlers
	issueHandler := handlers.NewIssueHandler(s.trackingService)
	prHandler := handlers.NewPullRequestHandler(s.trackingService)
	subHandler := handlers.NewSubscriptionHandler(s.trackingService)

	// API v1 routes
	v1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Issue routes
	v1.HandleFunc("/issues", issueHandler.ListIssues).Methods("GET")
	v1.HandleFunc("/issues", issueHandler.TrackIssue).Methods("POST")
	v1.HandleFunc("/issues/{key}", issueHandler.GetIssue).Methods("GET")
	v1.HandleFunc("/issues/{key}/subscribe", issueHandler.SubscribeToIssue).Methods("POST")
	v1.HandleFunc("/issues/{key}/unsubscribe", issueHandler.UnsubscribeFromIssue).Methods("DELETE")

	// Pull request routes
	v1.HandleFunc("/issues/{key}/prs", prHandler.ListPullRequests).Methods("GET")
	v1.HandleFunc("/issues/{key}/prs", prHandler.AddPullRequest).Methods("POST")
	v1.HandleFunc("/issues/{key}/prs/{prNumber}", prHandler.UpdatePullRequest).Methods("PUT")

	// Subscription routes
	v1.HandleFunc("/subscriptions", subHandler.ListSubscriptions).Methods("GET")
	v1.HandleFunc("/subscriptions/{id}", subHandler.GetSubscription).Methods("GET")
	v1.HandleFunc("/subscriptions/{id}", subHandler.UpdateSubscription).Methods("PUT")
	v1.HandleFunc("/subscriptions/{id}", subHandler.DeleteSubscription).Methods("DELETE")
}

// Start starts the API server
func (s *Server) Start(addr string) error {
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
