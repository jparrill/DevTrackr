package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Storage represents the SQLite database connection
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new SQLite database connection
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{db: db}, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// createTables creates the necessary tables in the database
func createTables(db *sql.DB) error {
	// Read schema from file
	schema, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}
	defer schema.Close()

	// Execute schema
	_, err = db.Exec(`
		-- Create issues table
		CREATE TABLE IF NOT EXISTS issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			jira_url TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		-- Create pull_requests table
		CREATE TABLE IF NOT EXISTS pull_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id INTEGER NOT NULL,
			number INTEGER NOT NULL,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (issue_id) REFERENCES issues(id),
			UNIQUE(issue_id, number)
		);

		-- Create subscriptions table
		CREATE TABLE IF NOT EXISTS subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			active BOOLEAN NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (issue_id) REFERENCES issues(id),
			UNIQUE(issue_id, user_id)
		);

		-- Create triggers for updated_at
		CREATE TRIGGER IF NOT EXISTS update_issues_timestamp
		AFTER UPDATE ON issues
		BEGIN
			UPDATE issues SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		CREATE TRIGGER IF NOT EXISTS update_pull_requests_timestamp
		AFTER UPDATE ON pull_requests
		BEGIN
			UPDATE pull_requests SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		CREATE TRIGGER IF NOT EXISTS update_subscriptions_timestamp
		AFTER UPDATE ON subscriptions
		BEGIN
			UPDATE subscriptions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)

	return err
}
