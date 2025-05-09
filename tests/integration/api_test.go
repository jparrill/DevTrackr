package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jparrill/devtrackr/internal/api/handlers"
	"github.com/jparrill/devtrackr/internal/jira"
	"github.com/jparrill/devtrackr/internal/services"
	"github.com/jparrill/devtrackr/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testIssue struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	JiraURL   string    `json:"jira_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createIssueRequest struct {
	JiraURL string `json:"jira_url"`
}

func setupTestServer(t *testing.T, mockStatus string) (*httptest.Server, func()) {
	// Inicializar storage
	db, err := storage.NewStorage("test.db")
	require.NoError(t, err)

	// Crear mock del cliente de Jira
	mockJira := jira.NewMockClient(mockStatus)

	// Crear servicio con el mock
	service := services.NewTrackingService(db, mockJira)

	// Crear handlers
	issueHandler := handlers.NewIssueHandler(service)
	prHandler := handlers.NewPullRequestHandler(service)
	subHandler := handlers.NewSubscriptionHandler(service)

	// Configurar router
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Issue routes
	api.HandleFunc("/issues", issueHandler.ListIssues).Methods("GET")
	api.HandleFunc("/issues", issueHandler.TrackIssue).Methods("POST")
	api.HandleFunc("/issues/{key}", issueHandler.GetIssue).Methods("GET")
	api.HandleFunc("/issues/{key}", issueHandler.DeleteIssue).Methods("DELETE")
	api.HandleFunc("/issues/{key}/status", issueHandler.UpdateIssueStatus).Methods("PUT")
	api.HandleFunc("/issues/{key}/subscribe", issueHandler.SubscribeToIssue).Methods("POST")
	api.HandleFunc("/issues/{key}/unsubscribe", issueHandler.UnsubscribeFromIssue).Methods("DELETE")

	// Pull request routes
	api.HandleFunc("/issues/{key}/pull-requests", prHandler.ListPullRequests).Methods("GET")
	api.HandleFunc("/issues/{key}/pull-requests", prHandler.AddPullRequest).Methods("POST")
	api.HandleFunc("/issues/{key}/pull-requests/{number}", prHandler.UpdatePullRequest).Methods("PUT")

	// Subscription routes
	api.HandleFunc("/subscriptions", subHandler.ListSubscriptions).Methods("GET")
	api.HandleFunc("/subscriptions/{id}", subHandler.GetSubscription).Methods("GET")
	api.HandleFunc("/subscriptions/{id}", subHandler.UpdateSubscription).Methods("PUT")
	api.HandleFunc("/subscriptions/{id}", subHandler.DeleteSubscription).Methods("DELETE")

	// Crear servidor de test
	ts := httptest.NewServer(r)

	// Función de limpieza
	cleanup := func() {
		ts.Close()
		db.Close()
		os.Remove("test.db")
	}

	return ts, cleanup
}

func TestCreateIssue(t *testing.T) {
	// Configurar servidor con mock que devuelve estado "Closed"
	server, cleanup := setupTestServer(t, "Closed")
	defer cleanup()

	// Preparar la petición
	reqBody := createIssueRequest{
		JiraURL: "https://issues.redhat.com/browse/OCPBUGS-48489",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Hacer la petición
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/issues", server.URL),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verificar la respuesta
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var issue testIssue
	err = json.NewDecoder(resp.Body).Decode(&issue)
	require.NoError(t, err)

	// Verificar los campos del issue
	assert.Equal(t, "OCPBUGS-48489", issue.Key)
	assert.Equal(t, "Mock Issue", issue.Title)
	assert.Equal(t, "Closed", issue.Status)
	assert.Equal(t, "https://issues.redhat.com/browse/OCPBUGS-48489", issue.JiraURL)
}

func TestListIssues(t *testing.T) {
	// Configurar servidor con mock que devuelve estado "Closed"
	server, cleanup := setupTestServer(t, "Closed")
	defer cleanup()

	// Primero crear un issue
	reqBody := createIssueRequest{
		JiraURL: "https://issues.redhat.com/browse/OCPBUGS-48489",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	_, err = http.Post(
		fmt.Sprintf("%s/api/v1/issues", server.URL),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(t, err)

	// Obtener la lista de issues
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/issues", server.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verificar la respuesta
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var issues []testIssue
	err = json.NewDecoder(resp.Body).Decode(&issues)
	require.NoError(t, err)

	// Verificar que hay al menos un issue
	assert.GreaterOrEqual(t, len(issues), 1)

	// Verificar el primer issue
	issue := issues[0]
	assert.Equal(t, "OCPBUGS-48489", issue.Key)
	assert.Equal(t, "Mock Issue", issue.Title)
	assert.Equal(t, "Closed", issue.Status)
	assert.Equal(t, "https://issues.redhat.com/browse/OCPBUGS-48489", issue.JiraURL)
}

func TestUpdateIssueStatus(t *testing.T) {
	// Configurar servidor con mock que devuelve estado inicial "Closed"
	server, cleanup := setupTestServer(t, "Closed")
	defer cleanup()

	// Primero crear un issue
	reqBody := createIssueRequest{
		JiraURL: "https://issues.redhat.com/browse/OCPBUGS-48489",
	}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/issues", server.URL),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var issue testIssue
	err = json.NewDecoder(resp.Body).Decode(&issue)
	require.NoError(t, err)

	// Verificar estado inicial
	assert.Equal(t, "Closed", issue.Status)

	// Actualizar el estado a "In Progress"
	updateReq := struct {
		Status string `json:"status"`
	}{
		Status: "In Progress",
	}
	jsonBody, err = json.Marshal(updateReq)
	require.NoError(t, err)

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/v1/issues/%s/status", server.URL, issue.Key),
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verificar que el estado se actualizó
	resp, err = http.Get(fmt.Sprintf("%s/api/v1/issues/%s", server.URL, issue.Key))
	require.NoError(t, err)
	defer resp.Body.Close()

	var updatedIssue testIssue
	err = json.NewDecoder(resp.Body).Decode(&updatedIssue)
	require.NoError(t, err)

	assert.Equal(t, "In Progress", updatedIssue.Status)
}
