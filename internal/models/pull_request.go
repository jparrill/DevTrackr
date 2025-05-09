package models

import "time"

// PullRequest represents a pull request associated with an issue
type PullRequest struct {
	ID           int64     `json:"id"`
	IssueID      int64     `json:"issue_id"`
	Number       int       `json:"number"`         // PR number in the repository
	Repository   string    `json:"repository"`     // Repository name
	Title        string    `json:"title"`          // PR title
	URL          string    `json:"url"`            // PR URL
	Status       PRStatus  `json:"status"`         // Current status
	TargetBranch string    `json:"target_branch"`  // Branch where the PR is targeting
	IsBackport   bool      `json:"is_backport"`    // Whether this is a backport PR
	OriginalPRID *int64    `json:"original_pr_id"` // Reference to the original PR if this is a backport
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PRStatus represents the possible states of a pull request
type PRStatus string

const (
	PRStatusOpen     PRStatus = "open"
	PRStatusMerged   PRStatus = "merged"
	PRStatusClosed   PRStatus = "closed"
	PRStatusDraft    PRStatus = "draft"
	PRStatusReview   PRStatus = "review"
	PRStatusApproved PRStatus = "approved"
)

// TableName returns the table name for the PullRequest model
func (PullRequest) TableName() string {
	return "pull_requests"
}
