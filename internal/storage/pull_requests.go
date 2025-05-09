package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jparrill/devtrackr/internal/models"
)

// CreatePullRequest creates a new pull request
func (s *Storage) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO pull_requests (issue_id, number, title, url, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pr.IssueID,
		pr.Number,
		pr.Title,
		pr.URL,
		pr.Status,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	return err
}

// GetPullRequest retrieves a pull request by issue ID and PR number
func (s *Storage) GetPullRequest(ctx context.Context, issueID int64, number int) (*models.PullRequest, error) {
	var pr models.PullRequest
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, issue_id, number, title, url, status, created_at, updated_at
		FROM pull_requests
		WHERE issue_id = ? AND number = ?`,
		issueID,
		number,
	).Scan(
		&pr.ID,
		&pr.IssueID,
		&pr.Number,
		&pr.Title,
		&pr.URL,
		&pr.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	pr.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	pr.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &pr, nil
}

// ListPullRequests retrieves all pull requests for an issue
func (s *Storage) ListPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, issue_id, number, title, url, status, created_at, updated_at
		FROM pull_requests
		WHERE issue_id = ?
		ORDER BY created_at DESC`,
		issueID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var createdAt, updatedAt string

		err := rows.Scan(
			&pr.ID,
			&pr.IssueID,
			&pr.Number,
			&pr.Title,
			&pr.URL,
			&pr.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		pr.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		pr.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		prs = append(prs, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

// UpdatePullRequest updates a pull request
func (s *Storage) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE pull_requests
		SET title = ?, url = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		pr.Title,
		pr.URL,
		pr.Status,
		time.Now().Format(time.RFC3339),
		pr.ID,
	)
	return err
}

// DeletePullRequest deletes a pull request
func (s *Storage) DeletePullRequest(id int64) error {
	query := `DELETE FROM pull_requests WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// GetUnmergedPullRequests retrieves all unmerged pull requests for an issue
func (s *Storage) GetUnmergedPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, issue_id, number, title, url, status, created_at, updated_at
		FROM pull_requests
		WHERE issue_id = ? AND status != 'merged'
		ORDER BY created_at DESC`,
		issueID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var createdAt, updatedAt string

		err := rows.Scan(
			&pr.ID,
			&pr.IssueID,
			&pr.Number,
			&pr.Title,
			&pr.URL,
			&pr.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		pr.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		pr.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		prs = append(prs, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
