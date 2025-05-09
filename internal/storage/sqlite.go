package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jparrill/devtrackr/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	// Create issues table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			jira_url TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			last_polled_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create subscriptions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (issue_id) REFERENCES issues(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create subscriptions table: %w", err)
	}

	// Create pull_requests table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pull_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id INTEGER NOT NULL,
			number INTEGER NOT NULL,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (issue_id) REFERENCES issues(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create pull_requests table: %w", err)
	}

	return nil
}

// CreateIssue creates a new issue
func (s *SQLiteStorage) CreateIssue(issue *models.Issue) error {
	_, err := s.db.Exec(`
		INSERT INTO issues (key, title, status, jira_url, created_at, updated_at, last_polled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, issue.Key, issue.Title, issue.Status, issue.JiraURL, issue.CreatedAt, issue.UpdatedAt, issue.LastPolledAt)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}
	return nil
}

// GetIssue retrieves an issue by its key
func (s *SQLiteStorage) GetIssue(key string) (*models.Issue, error) {
	var issue models.Issue
	err := s.db.QueryRow(`
		SELECT id, key, title, status, jira_url, created_at, updated_at, last_polled_at
		FROM issues WHERE key = ?
	`, key).Scan(
		&issue.ID,
		&issue.Key,
		&issue.Title,
		&issue.Status,
		&issue.JiraURL,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.LastPolledAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("issue not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return &issue, nil
}

// GetIssueByKey is an alias for GetIssue
func (s *SQLiteStorage) GetIssueByKey(key string) (*models.Issue, error) {
	return s.GetIssue(key)
}

// ListIssues returns all issues
func (s *SQLiteStorage) ListIssues() ([]models.Issue, error) {
	rows, err := s.db.Query(`
		SELECT id, key, title, status, jira_url, created_at, updated_at, last_polled_at
		FROM issues
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	defer rows.Close()

	var issues []models.Issue
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(
			&issue.ID,
			&issue.Key,
			&issue.Title,
			&issue.Status,
			&issue.JiraURL,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.LastPolledAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

// UpdateIssue updates an issue
func (s *SQLiteStorage) UpdateIssue(issue *models.Issue) error {
	_, err := s.db.Exec(`
		UPDATE issues
		SET title = ?, status = ?, updated_at = ?, last_polled_at = ?
		WHERE key = ?
	`, issue.Title, issue.Status, issue.UpdatedAt, issue.LastPolledAt, issue.Key)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}
	return nil
}

// DeleteIssue deletes an issue
func (s *SQLiteStorage) DeleteIssue(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM issues WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}
	return nil
}

// CreateSubscription creates a new subscription
func (s *SQLiteStorage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
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
func (s *SQLiteStorage) GetSubscription(ctx context.Context, issueID, userID int64) (*models.Subscription, error) {
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
func (s *SQLiteStorage) GetSubscriptionByID(ctx context.Context, id int64) (*models.Subscription, error) {
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
func (s *SQLiteStorage) ListSubscriptions(ctx context.Context, userID int64) ([]models.Subscription, error) {
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
func (s *SQLiteStorage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
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
func (s *SQLiteStorage) DeleteSubscription(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM subscriptions WHERE id = ?", id)
	return err
}

// CreatePullRequest creates a new pull request
func (s *SQLiteStorage) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
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
func (s *SQLiteStorage) GetPullRequest(ctx context.Context, issueID int64, number int) (*models.PullRequest, error) {
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
func (s *SQLiteStorage) ListPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
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
func (s *SQLiteStorage) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
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

// GetUnmergedPullRequests retrieves all unmerged pull requests for an issue
func (s *SQLiteStorage) GetUnmergedPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
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

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
