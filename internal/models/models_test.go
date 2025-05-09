package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIssueModel(t *testing.T) {
	// Create a test issue
	now := time.Now()
	issue := Issue{
		ID:        1,
		Key:       "TEST-123",
		Title:     "Test Issue",
		Status:    "Open",
		JiraURL:   "https://issues.redhat.com/browse/TEST-123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test fields
	assert.Equal(t, int64(1), issue.ID)
	assert.Equal(t, "TEST-123", issue.Key)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "Open", issue.Status)
	assert.Equal(t, "https://issues.redhat.com/browse/TEST-123", issue.JiraURL)
	assert.Equal(t, now, issue.CreatedAt)
	assert.Equal(t, now, issue.UpdatedAt)
}

func TestPullRequestModel(t *testing.T) {
	// Create a test pull request
	now := time.Now()
	pr := PullRequest{
		ID:           1,
		IssueID:      1,
		Number:       123,
		Repository:   "test-repo",
		Title:        "Test PR",
		URL:          "https://github.com/test-repo/pull/123",
		Status:       PRStatusOpen,
		TargetBranch: "main",
		IsBackport:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Test fields
	assert.Equal(t, int64(1), pr.ID)
	assert.Equal(t, int64(1), pr.IssueID)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "test-repo", pr.Repository)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "https://github.com/test-repo/pull/123", pr.URL)
	assert.Equal(t, PRStatusOpen, pr.Status)
	assert.Equal(t, "main", pr.TargetBranch)
	assert.False(t, pr.IsBackport)
	assert.Equal(t, now, pr.CreatedAt)
	assert.Equal(t, now, pr.UpdatedAt)
}

func TestPRStatusConstants(t *testing.T) {
	// Test PR status constants
	assert.Equal(t, PRStatus("open"), PRStatusOpen)
	assert.Equal(t, PRStatus("merged"), PRStatusMerged)
	assert.Equal(t, PRStatus("closed"), PRStatusClosed)
	assert.Equal(t, PRStatus("draft"), PRStatusDraft)
	assert.Equal(t, PRStatus("review"), PRStatusReview)
	assert.Equal(t, PRStatus("approved"), PRStatusApproved)
}

func TestSubscriptionModel(t *testing.T) {
	// Create a test subscription
	now := time.Now()
	sub := Subscription{
		ID:        1,
		IssueID:   1,
		UserID:    123,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test fields
	assert.Equal(t, int64(1), sub.ID)
	assert.Equal(t, int64(1), sub.IssueID)
	assert.Equal(t, int64(123), sub.UserID)
	assert.True(t, sub.Active)
	assert.Equal(t, now, sub.CreatedAt)
	assert.Equal(t, now, sub.UpdatedAt)
}
