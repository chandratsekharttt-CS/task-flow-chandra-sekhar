package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/internal/handlers"
	"taskflow/internal/middleware"
	"taskflow/internal/repository"
)

// testDB connects to a test database.
// Set TEST_DATABASE_URL env var to run integration tests.
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Clean up test data before each test
	pool.Exec(context.Background(), "DELETE FROM tasks")
	pool.Exec(context.Background(), "DELETE FROM projects")
	pool.Exec(context.Background(), "DELETE FROM users")

	return pool
}

func setupAuth(pool *pgxpool.Pool) (*handlers.AuthHandler, *middleware.AuthMiddleware) {
	userRepo := repository.NewUserRepository(pool)
	authMiddle := middleware.NewAuthMiddleware("test-jwt-secret")
	authHandler := handlers.NewAuthHandler(userRepo, authMiddle, 4) // low cost for speed
	return authHandler, authMiddle
}

// Test 1: Register a new user
func TestRegister(t *testing.T) {
	pool := testDB(t)
	authHandler, _ := setupAuth(pool)

	body := `{"name":"Test User","email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected token in response")
	}
	if user, ok := resp["user"].(map[string]interface{}); ok {
		if user["email"] != "test@example.com" {
			t.Errorf("expected email test@example.com, got %v", user["email"])
		}
	} else {
		t.Error("expected user object in response")
	}
}

// Test 2: Login with valid credentials
func TestLogin(t *testing.T) {
	pool := testDB(t)
	authHandler, _ := setupAuth(pool)

	// First register
	regBody := `{"name":"Login Test","email":"login@example.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	authHandler.Register(regRec, regReq)

	if regRec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", regRec.Code, regRec.Body.String())
	}

	// Then login
	loginBody := `{"email":"login@example.com","password":"password123"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	authHandler.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", loginRec.Code, loginRec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(loginRec.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected token in login response")
	}
}

// Test 3: Create task in a project and verify it appears
func TestCreateAndListTasks(t *testing.T) {
	pool := testDB(t)
	authHandler, authMiddle := setupAuth(pool)

	// Register user
	regBody := `{"name":"Task Tester","email":"tasks@example.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	authHandler.Register(regRec, regReq)

	var authResp map[string]interface{}
	json.Unmarshal(regRec.Body.Bytes(), &authResp)
	token := authResp["token"].(string)
	userMap := authResp["user"].(map[string]interface{})
	userID := userMap["id"].(string)

	// Create project directly in DB
	projectRepo := repository.NewProjectRepository(pool)
	desc := "Test project"
	project, err := projectRepo.Create(context.Background(), "Test Project", &desc, userID)
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create task via handler
	taskRepo := repository.NewTaskRepository(pool)
	taskHandler := handlers.NewTaskHandler(taskRepo, projectRepo)

	taskBody := `{"title":"Test Task","status":"todo","priority":"high"}`
	taskReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%s/tasks", project.ID), bytes.NewBufferString(taskBody))
	taskReq.Header.Set("Content-Type", "application/json")
	taskReq.Header.Set("Authorization", "Bearer "+token)

	// Add chi URL params and auth context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", project.ID)
	ctx := context.WithValue(taskReq.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.UserEmailKey, "tasks@example.com")
	taskReq = taskReq.WithContext(ctx)

	taskRec := httptest.NewRecorder()
	taskHandler.Create(taskRec, taskReq)

	if taskRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", taskRec.Code, taskRec.Body.String())
	}

	// List tasks
	listReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%s/tasks", project.ID), nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listCtx := context.WithValue(listReq.Context(), chi.RouteCtxKey, rctx)
	listCtx = context.WithValue(listCtx, middleware.UserIDKey, userID)
	listReq = listReq.WithContext(listCtx)

	listRec := httptest.NewRecorder()
	taskHandler.List(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listRec.Code, listRec.Body.String())
	}

	var listResp map[string]interface{}
	json.Unmarshal(listRec.Body.Bytes(), &listResp)
	data := listResp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("expected 1 task, got %d", len(data))
	}

	// Suppress unused import warning
	_ = authMiddle
}
