package models

import "time"

// Subscription represents a user's subscription to an issue
type Subscription struct {
	ID        int64     `json:"id"`
	IssueID   int64     `json:"issue_id"`
	UserID    int64     `json:"user_id"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}
