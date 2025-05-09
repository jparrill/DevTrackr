package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jparrill/devtrackr/internal/models"
)

// CreateSubscription creates a new subscription
func (s *Storage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO subscriptions (issue_id, user_id, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		sub.IssueID,
		sub.UserID,
		sub.Active,
		time.Now().Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)
	return err
}

// GetSubscription retrieves a subscription by issue ID and user ID
func (s *Storage) GetSubscription(ctx context.Context, issueID, userID int64) (*models.Subscription, error) {
	var sub models.Subscription
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, issue_id, user_id, active, created_at, updated_at
		FROM subscriptions
		WHERE issue_id = ? AND user_id = ?`,
		issueID,
		userID,
	).Scan(
		&sub.ID,
		&sub.IssueID,
		&sub.UserID,
		&sub.Active,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	sub.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	sub.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &sub, nil
}

// GetSubscriptionByID retrieves a subscription by its ID
func (s *Storage) GetSubscriptionByID(ctx context.Context, id int64) (*models.Subscription, error) {
	var sub models.Subscription
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, issue_id, user_id, active, created_at, updated_at
		FROM subscriptions
		WHERE id = ?`,
		id,
	).Scan(
		&sub.ID,
		&sub.IssueID,
		&sub.UserID,
		&sub.Active,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	sub.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	sub.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &sub, nil
}

// ListSubscriptions retrieves all subscriptions for a user
func (s *Storage) ListSubscriptions(ctx context.Context, userID int64) ([]models.Subscription, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, issue_id, user_id, active, created_at, updated_at
		FROM subscriptions
		WHERE user_id = ?
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		var createdAt, updatedAt string

		err := rows.Scan(
			&sub.ID,
			&sub.IssueID,
			&sub.UserID,
			&sub.Active,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		sub.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		sub.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// UpdateSubscription updates a subscription
func (s *Storage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE subscriptions
		SET active = ?, updated_at = ?
		WHERE id = ?`,
		sub.Active,
		time.Now().Format(time.RFC3339),
		sub.ID,
	)
	return err
}

// DeleteSubscription deletes a subscription
func (s *Storage) DeleteSubscription(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM subscriptions WHERE id = ?", id)
	return err
}

// HasUnmergedPullRequests checks if an issue has any unmerged pull requests
func (s *Storage) HasUnmergedPullRequests(issueID int64) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM pull_requests
		WHERE issue_id = ? AND status != 'merged'
	`
	var hasUnmerged bool
	err := s.db.QueryRow(query, issueID).Scan(&hasUnmerged)
	return hasUnmerged, err
}
