package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jparrill/devtrackr/internal/models"
)

// Client represents a Jira API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// JiraIssue represents the Jira API response structure
type JiraIssue struct {
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Description string `json:"description"`
	} `json:"fields"`
}

// NewClient creates a new Jira client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// GetIssue retrieves issue information from Jira
func (c *Client) GetIssue(ctx context.Context, issueURL string) (*models.Issue, error) {
	// Parse the URL to extract the issue key
	parsedURL, err := url.Parse(issueURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Jira URL: %w", err)
	}

	// Extract the issue key from the URL path
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid Jira URL format")
	}
	key := pathParts[len(pathParts)-1]

	// Construct API URL
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s", c.baseURL, key)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for JSON response
	req.Header.Set("Accept", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch issue: status code %d", resp.StatusCode)
	}

	// Parse JSON response
	var jiraIssue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&jiraIssue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Create issue object
	issue := models.Issue{
		Key:     key,
		Title:   jiraIssue.Fields.Summary,
		Status:  jiraIssue.Fields.Status.Name,
		JiraURL: issueURL,
	}

	return &issue, nil
}
