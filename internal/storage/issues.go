package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jparrill/devtrackr/internal/models"
)

// GetIssueByKey retrieves an issue by its key
func (s *Storage) GetIssueByKey(ctx context.Context, key string) (*models.Issue, error) {
	var issue models.Issue
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, key, title, status, jira_url, created_at, updated_at
		FROM issues WHERE key = ?`,
		key,
	).Scan(
		&issue.ID,
		&issue.Key,
		&issue.Title,
		&issue.Status,
		&issue.JiraURL,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	issue.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	issue.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &issue, nil
}

// CreateIssue creates a new issue in the database
func (s *Storage) CreateIssue(issue *models.Issue) error {
	query := `
		INSERT INTO issues (key, title, status, jira_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now().Format(time.RFC3339)
	result, err := s.db.Exec(query,
		issue.Key,
		issue.Title,
		issue.Status,
		issue.JiraURL,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	issue.ID = id
	issue.CreatedAt, _ = time.Parse(time.RFC3339, now)
	issue.UpdatedAt = issue.CreatedAt
	return nil
}

// UpdateIssue updates an existing issue
func (s *Storage) UpdateIssue(issue *models.Issue) error {
	query := `
		UPDATE issues
		SET title = ?, status = ?, jira_url = ?, updated_at = ?
		WHERE key = ?
	`
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(query,
		issue.Title,
		issue.Status,
		issue.JiraURL,
		now,
		issue.Key,
	)
	if err != nil {
		return err
	}

	issue.UpdatedAt, _ = time.Parse(time.RFC3339, now)
	return nil
}

// ListIssues retrieves all issues from the database
func (s *Storage) ListIssues() ([]models.Issue, error) {
	query := `
		SELECT id, key, title, status, jira_url, created_at, updated_at
		FROM issues
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.Issue
	for rows.Next() {
		var issue models.Issue
		var createdAt, updatedAt string
		err := rows.Scan(
			&issue.ID,
			&issue.Key,
			&issue.Title,
			&issue.Status,
			&issue.JiraURL,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		issue.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		issue.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		issues = append(issues, issue)
	}

	return issues, nil
}

// DeleteIssue deletes an issue by its key
func (s *Storage) DeleteIssue(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM issues WHERE key = ?", key)
	return err
}
