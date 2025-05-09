package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jparrill/devtrackr/internal/api"
	"github.com/jparrill/devtrackr/internal/jira"
	"github.com/jparrill/devtrackr/internal/services"
	"github.com/jparrill/devtrackr/internal/storage"
	"github.com/spf13/cobra"
)

var (
	pollingInterval int
	rootCmd         = &cobra.Command{
		Use:   "devtrackr",
		Short: "DevTrackr - Jira issue tracking service",
		Long:  `DevTrackr is a service that tracks Jira issues and their status changes.`,
	}
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the DevTrackr server",
		Long:  `Start the DevTrackr server and begin polling tracked issues.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Initialize services
			storage, err := initStorage()
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			jira, err := initJira()
			if err != nil {
				return fmt.Errorf("failed to initialize Jira client: %w", err)
			}

			// Create and start polling service
			pollingService := services.NewPollingService(storage, jira, time.Duration(pollingInterval)*time.Minute)
			if err := pollingService.Start(ctx); err != nil {
				return fmt.Errorf("failed to start polling service: %w", err)
			}

			// Create tracking service
			trackingService := services.NewTrackingService(storage, jira)

			// Start API server
			api := initAPI(trackingService)
			go func() {
				if err := api.Start(":8080"); err != nil {
					fmt.Printf("Error starting API server: %v\n", err)
					os.Exit(1)
				}
			}()

			fmt.Printf("Server started with polling interval of %d minutes\n", pollingInterval)

			// Wait for interrupt signal
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			// Stop polling service
			pollingService.Stop()

			return nil
		},
	}

	setPollingCmd = &cobra.Command{
		Use:   "set-polling [issue-key] [interval]",
		Short: "Set the polling interval for an issue",
		Long:  `Set the polling interval (in minutes) for a specific issue. Use 0 to use the default interval.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			interval, err := time.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("invalid interval format: %w", err)
			}

			// Initialize services
			storage, err := initStorage()
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			jira, err := initJira()
			if err != nil {
				return fmt.Errorf("failed to initialize Jira client: %w", err)
			}

			// Create tracking service
			trackingService := services.NewTrackingService(storage, jira)

			// Update polling interval
			if err := trackingService.UpdateIssuePollingInterval(context.Background(), key, int(interval.Seconds())); err != nil {
				return fmt.Errorf("failed to update polling interval: %w", err)
			}

			fmt.Printf("Updated polling interval for issue %s to %v\n", key, interval)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(setPollingCmd)

	// Add polling interval flag
	serveCmd.Flags().IntVarP(&pollingInterval, "poll", "p", 5, "Default polling interval in minutes (default: 5)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initStorage initializes the storage service
func initStorage() (storage.Storage, error) {
	// TODO: Make this configurable
	return storage.NewSQLiteStorage("devtrackr.db")
}

// initJira initializes the Jira client
func initJira() (jira.JiraClient, error) {
	// TODO: Make this configurable
	return jira.NewClient("https://issues.redhat.com"), nil
}

// initAPI initializes the API server
func initAPI(trackingService *services.TrackingService) *api.Server {
	return api.NewServer(trackingService)
}
