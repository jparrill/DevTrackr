package services

import (
	"context"
	"testing"
	"time"

	"github.com/jparrill/devtrackr/internal/jira"
	"github.com/jparrill/devtrackr/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CreateIssue(issue *models.Issue) error {
	args := m.Called(issue)
	return args.Error(0)
}

func (m *MockStorage) GetIssue(key string) (*models.Issue, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Issue), args.Error(1)
}

func (m *MockStorage) ListIssues() ([]models.Issue, error) {
	args := m.Called()
	return args.Get(0).([]models.Issue), args.Error(1)
}

func (m *MockStorage) UpdateIssue(issue *models.Issue) error {
	args := m.Called(issue)
	return args.Error(0)
}

func (m *MockStorage) DeleteIssue(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockStorage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockStorage) GetSubscription(ctx context.Context, issueID, userID int64) (*models.Subscription, error) {
	args := m.Called(ctx, issueID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockStorage) DeleteSubscription(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStorage) ListPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
	args := m.Called(ctx, issueID)
	return args.Get(0).([]*models.PullRequest), args.Error(1)
}

func (m *MockStorage) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockStorage) GetPullRequest(ctx context.Context, issueID int64, prNumber int) (*models.PullRequest, error) {
	args := m.Called(ctx, issueID, prNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *MockStorage) UpdatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockStorage) GetUnmergedPullRequests(ctx context.Context, issueID int64) ([]*models.PullRequest, error) {
	args := m.Called(ctx, issueID)
	return args.Get(0).([]*models.PullRequest), args.Error(1)
}

func (m *MockStorage) ListSubscriptions(ctx context.Context, userID int64) ([]models.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockStorage) GetSubscriptionByID(ctx context.Context, id int64) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockStorage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func TestTrackIssue(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test tracking new issue
	ctx := context.Background()
	jiraURL := "https://issues.redhat.com/browse/TEST-123"

	// Mock storage to return error (issue doesn't exist)
	mockStorage.On("GetIssue", "TEST-123").Return(nil, assert.AnError)
	mockStorage.On("CreateIssue", mock.Anything).Return(nil)

	issue, err := service.TrackIssue(ctx, jiraURL)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, "TEST-123", issue.Key)
	assert.Equal(t, "Mock Issue", issue.Title)
	assert.Equal(t, "In Progress", issue.Status)

	// Test tracking existing issue
	existingIssue := &models.Issue{
		ID:        1,
		Key:       "TEST-123",
		Title:     "Old Title",
		Status:    "Old Status",
		JiraURL:   jiraURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.On("GetIssue", "TEST-123").Return(existingIssue, nil)
	mockStorage.On("UpdateIssue", mock.Anything).Return(nil)

	issue, err = service.TrackIssue(ctx, jiraURL)
	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, "TEST-123", issue.Key)
	assert.Equal(t, "Mock Issue", issue.Title)
	assert.Equal(t, "In Progress", issue.Status)
}

func TestSubscribeToIssue(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test subscribing to issue
	ctx := context.Background()
	key := "TEST-123"
	userID := int64(123)

	issue := &models.Issue{
		ID:        1,
		Key:       key,
		Title:     "Test Issue",
		Status:    "In Progress",
		JiraURL:   "https://issues.redhat.com/browse/TEST-123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.On("GetIssue", key).Return(issue, nil)
	mockStorage.On("CreateSubscription", ctx, mock.Anything).Return(nil)

	sub, err := service.SubscribeToIssue(ctx, key, userID)
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, issue.ID, sub.IssueID)
	assert.Equal(t, userID, sub.UserID)
	assert.True(t, sub.Active)
}

func TestListPullRequests(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test listing pull requests
	ctx := context.Background()
	key := "TEST-123"

	issue := &models.Issue{
		ID:        1,
		Key:       key,
		Title:     "Test Issue",
		Status:    "In Progress",
		JiraURL:   "https://issues.redhat.com/browse/TEST-123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	prs := []*models.PullRequest{
		{
			ID:         1,
			IssueID:    issue.ID,
			Number:     123,
			Repository: "test-repo",
			Title:      "Test PR",
			URL:        "https://github.com/test-repo/pull/123",
			Status:     models.PRStatusOpen,
		},
	}

	mockStorage.On("GetIssue", key).Return(issue, nil)
	mockStorage.On("ListPullRequests", ctx, issue.ID).Return(prs, nil)

	result, err := service.ListPullRequests(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, prs[0].Number, result[0].Number)
	assert.Equal(t, prs[0].Repository, result[0].Repository)
	assert.Equal(t, prs[0].Status, result[0].Status)
}
