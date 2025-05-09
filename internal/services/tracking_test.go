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

func (m *MockStorage) GetIssueByKey(key string) (*models.Issue, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Issue), args.Error(1)
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
	mockStorage.On("GetIssueByKey", "TEST-123").Return(nil, assert.AnError)
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

	mockStorage.On("GetIssueByKey", "TEST-123").Return(existingIssue, nil)
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

func TestUpdateIssueStatus(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test updating issue status
	ctx := context.Background()
	issue := &models.Issue{
		ID:        1,
		Key:       "TEST-123",
		Title:     "Test Issue",
		Status:    "In Progress",
		JiraURL:   "https://issues.redhat.com/browse/TEST-123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newStatus := "Done"

	mockStorage.On("UpdateIssue", mock.MatchedBy(func(i *models.Issue) bool {
		return i.Status == newStatus
	})).Return(nil)

	err := service.UpdateIssueStatus(ctx, issue, newStatus)
	assert.NoError(t, err)
	assert.Equal(t, newStatus, issue.Status)
}

func TestHasUnmergedPullRequests(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test when there are unmerged PRs
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

	unmergedPRs := []*models.PullRequest{
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

	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("GetUnmergedPullRequests", ctx, issue.ID).Return(unmergedPRs, nil).Once()

	hasUnmerged, err := service.HasUnmergedPullRequests(ctx, key)
	assert.NoError(t, err)
	assert.True(t, hasUnmerged)

	// Test when there are no unmerged PRs
	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("GetUnmergedPullRequests", ctx, issue.ID).Return([]*models.PullRequest{}, nil).Once()

	hasUnmerged, err = service.HasUnmergedPullRequests(ctx, key)
	assert.NoError(t, err)
	assert.False(t, hasUnmerged)
}

func TestUnsubscribeFromIssue(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test unsubscribing from issue
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

	subscription := &models.Subscription{
		ID:      1,
		IssueID: issue.ID,
		UserID:  userID,
		Active:  true,
	}

	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("GetSubscription", ctx, issue.ID, userID).Return(subscription, nil).Once()
	mockStorage.On("DeleteSubscription", ctx, subscription.ID).Return(nil).Once()

	err := service.UnsubscribeFromIssue(ctx, key, userID)
	assert.NoError(t, err)

	// Test when subscription doesn't exist
	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("GetSubscription", ctx, issue.ID, userID).Return(nil, assert.AnError).Once()

	err = service.UnsubscribeFromIssue(ctx, key, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), assert.AnError.Error())
}

func TestAddPullRequest(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test adding a pull request
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

	newPR := &models.PullRequest{
		Number:     123,
		Repository: "test-repo",
		Title:      "Test PR",
		URL:        "https://github.com/test-repo/pull/123",
		Status:     models.PRStatusOpen,
	}

	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("CreatePullRequest", ctx, mock.MatchedBy(func(pr *models.PullRequest) bool {
		return pr.IssueID == issue.ID && pr.Number == newPR.Number
	})).Return(nil).Once()

	pr, err := service.AddPullRequest(ctx, key, newPR)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, issue.ID, pr.IssueID)
	assert.Equal(t, newPR.Number, pr.Number)
	assert.Equal(t, newPR.Repository, pr.Repository)
}

func TestUpdatePullRequest(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test updating a pull request
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

	existingPR := &models.PullRequest{
		ID:         1,
		IssueID:    issue.ID,
		Number:     123,
		Repository: "test-repo",
		Title:      "Old Title",
		URL:        "https://github.com/test-repo/pull/123",
		Status:     models.PRStatusOpen,
	}

	updatedPR := &models.PullRequest{
		Number:     123,
		Repository: "test-repo",
		Title:      "Updated Title",
		URL:        "https://github.com/test-repo/pull/123",
		Status:     models.PRStatusMerged,
	}

	mockStorage.On("GetIssue", key).Return(issue, nil).Once()
	mockStorage.On("GetPullRequest", ctx, issue.ID, updatedPR.Number).Return(existingPR, nil).Once()
	mockStorage.On("UpdatePullRequest", ctx, mock.MatchedBy(func(pr *models.PullRequest) bool {
		return pr.ID == existingPR.ID && pr.Status == models.PRStatusMerged
	})).Return(nil).Once()

	err := service.UpdatePullRequest(ctx, key, updatedPR.Number, updatedPR)
	assert.NoError(t, err)
}

func TestListSubscriptions(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test listing subscriptions
	ctx := context.Background()
	userID := int64(123)

	subscriptions := []models.Subscription{
		{
			ID:      1,
			IssueID: 1,
			UserID:  userID,
			Active:  true,
		},
		{
			ID:      2,
			IssueID: 2,
			UserID:  userID,
			Active:  false,
		},
	}

	mockStorage.On("ListSubscriptions", ctx, userID).Return(subscriptions, nil).Once()

	result, err := service.ListSubscriptions(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, subscriptions[0].ID, result[0].ID)
	assert.Equal(t, subscriptions[1].ID, result[1].ID)
}

func TestUpdateSubscription(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test updating a subscription
	ctx := context.Background()
	subID := int64(1)

	existingSub := &models.Subscription{
		ID:      subID,
		IssueID: 1,
		UserID:  123,
		Active:  true,
	}

	updatedSub := &models.Subscription{
		Active: false,
	}

	mockStorage.On("GetSubscriptionByID", ctx, subID).Return(existingSub, nil).Once()
	mockStorage.On("UpdateSubscription", ctx, mock.MatchedBy(func(sub *models.Subscription) bool {
		return sub.ID == existingSub.ID && !sub.Active
	})).Return(nil).Once()

	err := service.UpdateSubscription(ctx, subID, updatedSub)
	assert.NoError(t, err)

	// Test updating non-existent subscription
	mockStorage.On("GetSubscriptionByID", ctx, subID).Return(nil, assert.AnError).Once()

	err = service.UpdateSubscription(ctx, subID, updatedSub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get subscription")
}

func TestDeleteSubscription(t *testing.T) {
	// Create mocks
	mockStorage := &MockStorage{}
	mockJira := jira.NewMockClient("In Progress")

	// Create service
	service := NewTrackingService(mockStorage, mockJira)

	// Test deleting a subscription
	ctx := context.Background()
	subID := int64(1)

	// Test successful deletion
	mockStorage.On("DeleteSubscription", ctx, subID).Return(nil).Once()

	err := service.DeleteSubscription(ctx, subID)
	assert.NoError(t, err)

	// Test deletion error
	mockStorage.On("DeleteSubscription", ctx, subID).Return(assert.AnError).Once()

	err = service.DeleteSubscription(ctx, subID)
	assert.Error(t, err)
}
