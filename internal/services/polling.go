package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jparrill/devtrackr/internal/jira"
)

// PollingService handles the background polling of issues
type PollingService struct {
	storage         Storage
	jira            jira.JiraClient
	stop            chan struct{}
	pollingInterval time.Duration
}

// NewPollingService creates a new polling service
func NewPollingService(storage Storage, jira jira.JiraClient, pollingInterval time.Duration) *PollingService {
	return &PollingService{
		storage:         storage,
		jira:            jira,
		stop:            make(chan struct{}),
		pollingInterval: pollingInterval,
	}
}

// Start begins the polling service
func (s *PollingService) Start(ctx context.Context) error {
	log.Printf("Starting polling service with default interval of %v", s.pollingInterval)

	// Start periodic polling
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.stop:
			return nil
		case <-ticker.C:
			if err := s.pollIssues(ctx); err != nil {
				log.Printf("Error polling issues: %v", err)
			}
		}
	}
}

// Stop stops the polling service
func (s *PollingService) Stop() {
	close(s.stop)
}

// pollIssues polls all issues
func (s *PollingService) pollIssues(ctx context.Context) error {
	log.Printf("Starting polling cycle...")

	// Get all issues
	issues, err := s.storage.ListIssues()
	if err != nil {
		return fmt.Errorf("failed to list issues: %w", err)
	}

	log.Printf("Found %d issues to check...", len(issues))

	now := time.Now()
	updated := 0
	unchanged := 0
	skipped := 0

	for _, issue := range issues {
		// Check if it's time to poll this issue
		interval := time.Duration(issue.PollingInterval) * time.Second
		if interval == 0 {
			interval = s.pollingInterval
		}

		// Skip if not enough time has passed since last poll
		if !issue.LastPolledAt.IsZero() && now.Sub(issue.LastPolledAt) < interval {
			skipped++
			continue
		}

		// Get latest issue data from Jira
		jiraIssue, err := s.jira.GetIssue(ctx, issue.JiraURL)
		if err != nil {
			log.Printf("Error getting issue %s from Jira: %v", issue.Key, err)
			continue
		}

		// Update issue if status has changed
		if jiraIssue.Status != issue.Status {
			issue.Status = jiraIssue.Status
			issue.Title = jiraIssue.Title
			issue.UpdatedAt = now
			issue.LastPolledAt = now

			if err := s.storage.UpdateIssue(&issue); err != nil {
				log.Printf("Error updating issue %s: %v", issue.Key, err)
				continue
			}

			log.Printf("Updated issue %s: %s -> %s", issue.Key, issue.Status, jiraIssue.Status)
			updated++
		} else {
			// Update LastPolledAt even if status hasn't changed
			issue.LastPolledAt = now
			if err := s.storage.UpdateIssue(&issue); err != nil {
				log.Printf("Error updating LastPolledAt for issue %s: %v", issue.Key, err)
			}
			unchanged++
		}
	}

	log.Printf("Polling cycle complete: %d issues updated, %d issues unchanged, %d issues skipped", updated, unchanged, skipped)
	return nil
}
