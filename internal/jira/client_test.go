package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	// Create a mock client with initial status
	mockClient := NewMockClient("In Progress")
	assert.NotNil(t, mockClient)
	assert.Equal(t, "In Progress", mockClient.currentStatus)

	// Test GetIssue with valid URL
	ctx := context.Background()
	issue, err := mockClient.GetIssue(ctx, "https://issues.redhat.com/browse/TEST-123")
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, "TEST-123", issue.Key)
	assert.Equal(t, "Mock Issue", issue.Title)
	assert.Equal(t, "In Progress", issue.Status)
	assert.Equal(t, "https://issues.redhat.com/browse/TEST-123", issue.JiraURL)

	// Test GetIssue with invalid URL
	_, err = mockClient.GetIssue(ctx, "invalid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Jira URL")

	// Test status update
	mockClient.UpdateStatus("Done")
	assert.Equal(t, "Done", mockClient.currentStatus)
	issue, err = mockClient.GetIssue(ctx, "https://issues.redhat.com/browse/TEST-123")
	assert.NoError(t, err)
	assert.Equal(t, "Done", issue.Status)
}

func TestClient(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-123", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"fields": {
				"summary": "Test Issue",
				"status": {
					"name": "In Progress"
				},
				"description": "Test Description"
			}
		}`))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient(server.URL)
	assert.NotNil(t, client)

	// Test GetIssue
	ctx := context.Background()
	issue, err := client.GetIssue(ctx, "https://issues.redhat.com/browse/TEST-123")
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, "TEST-123", issue.Key)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "In Progress", issue.Status)
	assert.Equal(t, "https://issues.redhat.com/browse/TEST-123", issue.JiraURL)

	// Test GetIssue with invalid URL
	_, err = client.GetIssue(ctx, "invalid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Jira URL")
}

func TestClientErrorHandling(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient(server.URL)

	// Test GetIssue with server error
	ctx := context.Background()
	_, err := client.GetIssue(ctx, "https://issues.redhat.com/browse/TEST-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status code 404")
}
