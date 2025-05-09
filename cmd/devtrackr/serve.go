package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/api/handlers"
	"github.com/jparrill/devtrackr/internal/jira"
	"github.com/jparrill/devtrackr/internal/services"
	"github.com/jparrill/devtrackr/internal/storage"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the DevTrackr API server",
	Long:  `Start the DevTrackr API server on the specified port`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := serve(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() error {
	// Initialize storage
	db, err := storage.NewStorage("devtrackr.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize Jira client
	jiraClient := jira.NewClient("https://issues.redhat.com")

	// Initialize services
	trackingService := services.NewTrackingService(db, jiraClient)

	// Initialize handlers
	issueHandler := handlers.NewIssueHandler(trackingService)
	prHandler := handlers.NewPullRequestHandler(trackingService)
	subHandler := handlers.NewSubscriptionHandler(trackingService)

	// Set up router
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Issue routes
	api.HandleFunc("/issues", issueHandler.ListIssues).Methods("GET")
	api.HandleFunc("/issues", issueHandler.TrackIssue).Methods("POST")
	api.HandleFunc("/issues/{key}", issueHandler.GetIssue).Methods("GET")
	api.HandleFunc("/issues/{key}", issueHandler.DeleteIssue).Methods("DELETE")
	api.HandleFunc("/issues/{key}/subscribe", issueHandler.SubscribeToIssue).Methods("POST")
	api.HandleFunc("/issues/{key}/unsubscribe", issueHandler.UnsubscribeFromIssue).Methods("DELETE")

	// Pull request routes
	api.HandleFunc("/issues/{key}/pull-requests", prHandler.ListPullRequests).Methods("GET")
	api.HandleFunc("/issues/{key}/pull-requests", prHandler.AddPullRequest).Methods("POST")
	api.HandleFunc("/issues/{key}/pull-requests/{number}", prHandler.UpdatePullRequest).Methods("PUT")

	// Subscription routes
	api.HandleFunc("/subscriptions", subHandler.ListSubscriptions).Methods("GET")
	api.HandleFunc("/subscriptions/{id}", subHandler.GetSubscription).Methods("GET")
	api.HandleFunc("/subscriptions/{id}", subHandler.UpdateSubscription).Methods("PUT")
	api.HandleFunc("/subscriptions/{id}", subHandler.DeleteSubscription).Methods("DELETE")

	// Start server
	log.Printf("Server starting on :8080")
	return http.ListenAndServe(":8080", r)
}
