package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jparrill/devtrackr/internal/jira"
	"github.com/jparrill/devtrackr/internal/models"
)

// Storage defines the interface for storage operations
type Storage interface {
	CreateIssue(issue *models.Issue) error
	GetIssue(key string) (*models.Issue, error)
	ListIssues() ([]models.Issue, error)
	UpdateIssue(issue *models.Issue) error
	DeleteIssue(ctx context.Context, key string) error
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	GetSubscription(ctx context.Context, issueID, userID int64) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, id int64) error
	ListPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error)
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) error
	GetPullRequest(ctx context.Context, issueID int64, prNumber int) (*models.PullRequest, error)
	UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error
	GetUnmergedPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error)
	ListSubscriptions(ctx context.Context, userID int64) ([]models.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	GetIssueByKey(key string) (*models.Issue, error)
}

// TrackingService handles the business logic for tracking issues and pull requests
type TrackingService struct {
	storage Storage
	jira    jira.JiraClient
}

// NewTrackingService creates a new tracking service
func NewTrackingService(storage Storage, jira jira.JiraClient) *TrackingService {
	return &TrackingService{
		storage: storage,
		jira:    jira,
	}
}

// TrackIssue tracks a Jira issue by its URL
func (s *TrackingService) TrackIssue(ctx context.Context, jiraURL string) (*models.Issue, error) {
	// Get issue from Jira
	jiraIssue, err := s.jira.GetIssue(ctx, jiraURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue from Jira: %w", err)
	}

	// Check if issue already exists
	existingIssue, err := s.storage.GetIssueByKey(jiraIssue.Key)
	if err == nil {
		// Issue already exists, update it
		existingIssue.Title = jiraIssue.Title
		existingIssue.Status = jiraIssue.Status
		existingIssue.UpdatedAt = time.Now()

		if err := s.storage.UpdateIssue(existingIssue); err != nil {
			return nil, fmt.Errorf("failed to update issue: %w", err)
		}

		return existingIssue, nil
	}

	// Create new issue
	issue := &models.Issue{
		Key:             jiraIssue.Key,
		Title:           jiraIssue.Title,
		Status:          jiraIssue.Status,
		JiraURL:         jiraURL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		PollingInterval: 300, // Default to 5 minutes (300 seconds)
		LastPolledAt:    time.Now(),
	}

	if err := s.storage.CreateIssue(issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

// ListIssues returns all tracked issues
func (s *TrackingService) ListIssues(ctx context.Context) ([]*models.Issue, error) {
	issues, err := s.storage.ListIssues()
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	// Convert []models.Issue to []*models.Issue
	result := make([]*models.Issue, len(issues))
	for i := range issues {
		result[i] = &issues[i]
	}
	return result, nil
}

// GetIssue retrieves a tracked issue by its key
func (s *TrackingService) GetIssue(ctx context.Context, key string) (*models.Issue, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// SubscribeToIssue subscribes a user to an issue
func (s *TrackingService) SubscribeToIssue(ctx context.Context, key string, userID int64) (*models.Subscription, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	sub := &models.Subscription{
		IssueID: issue.ID,
		UserID:  userID,
		Active:  true,
	}

	if err := s.storage.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// UnsubscribeFromIssue unsubscribes a user from an issue
func (s *TrackingService) UnsubscribeFromIssue(ctx context.Context, key string, userID int64) error {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	sub, err := s.storage.GetSubscription(ctx, issue.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	return s.storage.DeleteSubscription(ctx, sub.ID)
}

// ListPullRequests returns all pull requests for an issue
func (s *TrackingService) ListPullRequests(ctx context.Context, key string) ([]*models.PullRequest, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	return s.storage.ListPullRequests(ctx, issue.ID)
}

// AddPullRequest adds a pull request to an issue
func (s *TrackingService) AddPullRequest(ctx context.Context, key string, pr *models.PullRequest) (*models.PullRequest, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	pr.IssueID = issue.ID
	if err := s.storage.CreatePullRequest(ctx, pr); err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}

// UpdatePullRequest updates a pull request
func (s *TrackingService) UpdatePullRequest(ctx context.Context, key string, prNumber int, pr *models.PullRequest) error {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	existingPR, err := s.storage.GetPullRequest(ctx, issue.ID, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	pr.ID = existingPR.ID
	pr.IssueID = issue.ID
	return s.storage.UpdatePullRequest(ctx, pr)
}

// ListSubscriptions returns all subscriptions for a user
func (s *TrackingService) ListSubscriptions(ctx context.Context, userID int64) ([]models.Subscription, error) {
	return s.storage.ListSubscriptions(ctx, userID)
}

// GetSubscription returns a subscription by ID
func (s *TrackingService) GetSubscription(ctx context.Context, id int64) (*models.Subscription, error) {
	return s.storage.GetSubscriptionByID(ctx, id)
}

// UpdateSubscription updates a subscription
func (s *TrackingService) UpdateSubscription(ctx context.Context, id int64, sub *models.Subscription) error {
	existingSub, err := s.storage.GetSubscriptionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	sub.ID = existingSub.ID
	sub.IssueID = existingSub.IssueID
	sub.UserID = existingSub.UserID
	return s.storage.UpdateSubscription(ctx, sub)
}

// DeleteSubscription deletes a subscription
func (s *TrackingService) DeleteSubscription(ctx context.Context, id int64) error {
	return s.storage.DeleteSubscription(ctx, id)
}

// HasUnmergedPullRequests checks if an issue has any unmerged pull requests
func (s *TrackingService) HasUnmergedPullRequests(ctx context.Context, key string) (bool, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return false, fmt.Errorf("failed to get issue: %w", err)
	}

	prs, err := s.storage.GetUnmergedPullRequests(ctx, issue.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get unmerged pull requests: %w", err)
	}

	return len(prs) > 0, nil
}

// DeleteIssue deletes a tracked issue
func (s *TrackingService) DeleteIssue(ctx context.Context, key string) error {
	return s.storage.DeleteIssue(ctx, key)
}

// UpdateIssue updates an existing issue
func (s *TrackingService) UpdateIssue(ctx context.Context, issue *models.Issue) error {
	return s.storage.UpdateIssue(issue)
}

// UpdateIssueStatus updates the status of an issue
func (s *TrackingService) UpdateIssueStatus(ctx context.Context, issue *models.Issue, status string) error {
	// Update the issue status
	issue.Status = status
	if err := s.storage.UpdateIssue(issue); err != nil {
		return fmt.Errorf("failed to update issue status: %w", err)
	}

	return nil
}

// GetIssueByKey retrieves an issue by its key
func (s *TrackingService) GetIssueByKey(ctx context.Context, key string) (*models.Issue, error) {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// UpdateIssuePollingInterval updates the polling interval for an issue
func (s *TrackingService) UpdateIssuePollingInterval(ctx context.Context, key string, interval int) error {
	issue, err := s.storage.GetIssue(key)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	issue.PollingInterval = interval
	issue.LastPolledAt = time.Now()

	if err := s.storage.UpdateIssue(issue); err != nil {
		return fmt.Errorf("failed to update issue polling interval: %w", err)
	}

	return nil
}
