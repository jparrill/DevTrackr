package models

import (
	"time"
)

// Issue represents a Jira issue being tracked
type Issue struct {
	ID              int64     `json:"id"`
	Key             string    `json:"key"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	JiraURL         string    `json:"jira_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PollingInterval int       `json:"polling_interval"` // Interval in minutes, 0 means no polling
	LastPolledAt    time.Time `json:"last_polled_at"`
}

// TableName returns the table name for the Issue model
func (Issue) TableName() string {
	return "issues"
}
