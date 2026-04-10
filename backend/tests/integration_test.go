// Integration tests for the TaskFlow API.
//
// These tests spin up the full HTTP stack against a real PostgreSQL database.
// Set TEST_DATABASE_URL to point at a test DB before running, e.g.:
//
//	TEST_DATABASE_URL="postgres://taskflow:taskflow_secret@localhost:5432/taskflow_test?sslmode=disable" \
//	  go test ./tests/ -v -count=1
//
// Tests are automatically skipped when TEST_DATABASE_URL is not set so they
// never break a plain `go test ./...` run in CI without a database.
package tests_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/config"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/database"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/handler"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/repository"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/router"
	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/service"
)

// testServer wraps httptest.Server and adds a JWT token field for authenticated requests.
type testServer struct {
	*httptest.Server
	pool *pgxpool.Pool
}

// newTestServer creates a full HTTP server backed by a real database.
// Skips the test if TEST_DATABASE_URL is not set.
func newTestServer(t *testing.T) *testServer {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dbURL)
	require.NoError(t, err, "connect to test database")

	// Run migrations
	m, err := migrate.New("file://../migrations", dbURL)
	require.NoError(t, err, "create migrator")
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		require.NoError(t, err, "run migrations")
	}
	m.Close()

	// Wire up the application
	cfg := &config.Config{
		JWTSecret:  "test-secret-key-for-integration-tests",
		BcryptCost: 4, // low cost for test speed
	}

	userRepo := repository.NewUserRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.BcryptCost)
	projectSvc := service.NewProjectService(projectRepo, taskRepo, nil)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, nil)

	authH := handler.NewAuthHandler(authSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	taskH := handler.NewTaskHandler(taskSvc)
	sseH := handler.NewSSEHandler(nil)

	r := router.New(cfg.JWTSecret, authH, projectH, taskH, sseH)
	srv := httptest.NewServer(r)

	ts := &testServer{Server: srv, pool: pool}

	// Clean up after each test
	t.Cleanup(func() {
		truncateAll(t, ctx, pool)
		srv.Close()
		pool.Close()
	})

	return ts
}

// truncateAll wipes test data so tests are independent.
func truncateAll(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, `TRUNCATE tasks, projects, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err, "truncate tables")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func post(t *testing.T, srv *testServer, path string, body any, token string) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func get(t *testing.T, srv *testServer, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
}

// ─── Test 1: Full auth flow ────────────────────────────────────────────────────

func TestAuth_RegisterAndLogin(t *testing.T) {
	srv := newTestServer(t)

	// Register a new user
	resp := post(t, srv, "/auth/register", map[string]string{
		"name": "Alice", "email": "alice@test.com", "password": "password123",
	}, "")
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var regBody map[string]any
	decode(t, resp, &regBody)
	assert.NotEmpty(t, regBody["token"], "token should be in register response")
	assert.Equal(t, "alice@test.com", regBody["user"].(map[string]any)["email"])

	// Login with the same credentials
	resp = post(t, srv, "/auth/login", map[string]string{
		"email": "alice@test.com", "password": "password123",
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginBody map[string]any
	decode(t, resp, &loginBody)
	token, ok := loginBody["token"].(string)
	assert.True(t, ok && token != "", "login should return a JWT token")

	// Wrong password must return 401
	resp = post(t, srv, "/auth/login", map[string]string{
		"email": "alice@test.com", "password": "wrongpassword",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Missing fields must return 400
	resp = post(t, srv, "/auth/login", map[string]string{"email": ""}, "")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Test 2: Project create and list ──────────────────────────────────────────

func TestProjects_CreateAndList(t *testing.T) {
	srv := newTestServer(t)

	// Register → get token
	resp := post(t, srv, "/auth/register", map[string]string{
		"name": "Bob", "email": "bob@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]any
	decode(t, resp, &reg)
	token := reg["token"].(string)

	// Unauthenticated request must return 401
	resp = get(t, srv, "/projects", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Create a project
	resp = post(t, srv, "/projects", map[string]string{
		"name": "Alpha Project", "description": "test project",
	}, token)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var proj map[string]any
	decode(t, resp, &proj)
	assert.Equal(t, "Alpha Project", proj["name"])
	projectID := proj["id"].(string)
	assert.NotEmpty(t, projectID)

	// List projects — should include the one we just created
	resp = get(t, srv, "/projects", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var listBody map[string]any
	decode(t, resp, &listBody)
	data := listBody["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Alpha Project", data[0].(map[string]any)["name"])

	// Missing name should return 400
	resp = post(t, srv, "/projects", map[string]string{"name": ""}, token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Test 3: Task create, list, and status filter ─────────────────────────────

func TestTasks_CreateAndFilter(t *testing.T) {
	srv := newTestServer(t)

	// Setup: register user, create project
	resp := post(t, srv, "/auth/register", map[string]string{
		"name": "Carol", "email": "carol@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]any
	decode(t, resp, &reg)
	token := reg["token"].(string)

	resp = post(t, srv, "/projects", map[string]string{"name": "Beta Project"}, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var proj map[string]any
	decode(t, resp, &proj)
	projectID := proj["id"].(string)

	taskURL := fmt.Sprintf("/projects/%s/tasks", projectID)

	// Create 3 tasks with different statuses
	tasks := []map[string]string{
		{"title": "Task 1", "status": "todo"},
		{"title": "Task 2", "status": "in_progress"},
		{"title": "Task 3", "status": "done"},
	}
	for _, td := range tasks {
		resp = post(t, srv, taskURL, td, token)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// List all tasks — should return 3
	resp = get(t, srv, taskURL, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var allTasks map[string]any
	decode(t, resp, &allTasks)
	assert.Len(t, allTasks["data"].([]any), 3)

	// Filter by status=todo — should return only 1
	resp = get(t, srv, taskURL+"?status=todo", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var todoTasks map[string]any
	decode(t, resp, &todoTasks)
	assert.Len(t, todoTasks["data"].([]any), 1)
	assert.Equal(t, "Task 1", todoTasks["data"].([]any)[0].(map[string]any)["title"])

	// Filter by status=done — should return only 1
	resp = get(t, srv, taskURL+"?status=done", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var doneTasks map[string]any
	decode(t, resp, &doneTasks)
	assert.Len(t, doneTasks["data"].([]any), 1)

	// Invalid status filter should return 400
	resp = get(t, srv, taskURL+"?status=invalid_status", token)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ─── Test 4: Authorization — only project owner can delete ────────────────────

func TestProjects_OwnershipEnforced(t *testing.T) {
	srv := newTestServer(t)

	// Alice creates a project
	resp := post(t, srv, "/auth/register", map[string]string{
		"name": "Alice", "email": "alice2@test.com", "password": "pass123",
	}, "")
	var alice map[string]any
	decode(t, resp, &alice)
	aliceToken := alice["token"].(string)

	resp = post(t, srv, "/projects", map[string]string{"name": "Alice's Project"}, aliceToken)
	var proj map[string]any
	decode(t, resp, &proj)
	projectID := proj["id"].(string)

	// Bob registers separately
	resp = post(t, srv, "/auth/register", map[string]string{
		"name": "Bob", "email": "bob2@test.com", "password": "pass123",
	}, "")
	var bob map[string]any
	decode(t, resp, &bob)
	bobToken := bob["token"].(string)

	// Bob tries to delete Alice's project — must get 403
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/projects/"+projectID, nil)
	req.Header.Set("Authorization", "Bearer "+bobToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// Alice can delete her own project
	req, _ = http.NewRequest(http.MethodDelete, srv.URL+"/projects/"+projectID, nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
