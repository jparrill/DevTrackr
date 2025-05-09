package storage

import (
	"context"

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
	Close() error
}
