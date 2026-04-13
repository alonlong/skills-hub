package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"skillhub/backend/internal/config"
	"skillhub/backend/internal/testenv"
)

func TestNewHandlerWiresCoreRoutes(t *testing.T) {
	t.Setenv("DATABASE_URL", testenv.DatabaseURL())
	t.Setenv("BACKEND_JWT_SECRET", "test-secret")
	t.Setenv("STORAGE_BASE_PATH", t.TempDir())
	t.Setenv("BOOTSTRAP_ADMIN_ENABLED", "true")

	if err := runMigrationsForServerTest(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	db := openServerTestDatabase(t)
	handler, err := newHandler(config.Load(), db)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	searchReq := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	searchRes := httptest.NewRecorder()
	handler.ServeHTTP(searchRes, searchReq)
	if searchRes.Code != http.StatusBadRequest {
		t.Fatalf("search route status: got %d want %d", searchRes.Code, http.StatusBadRequest)
	}

	namespaceReq := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", nil)
	namespaceRes := httptest.NewRecorder()
	handler.ServeHTTP(namespaceRes, namespaceReq)
	if namespaceRes.Code != http.StatusUnauthorized {
		t.Fatalf("namespace route status: got %d want %d", namespaceRes.Code, http.StatusUnauthorized)
	}

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/skills/team-ai/email-helper/versions", nil)
	publishRes := httptest.NewRecorder()
	handler.ServeHTTP(publishRes, publishReq)
	if publishRes.Code != http.StatusUnauthorized {
		t.Fatalf("skill publish route status: got %d want %d", publishRes.Code, http.StatusUnauthorized)
	}
}

func openServerTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	return testenv.OpenDatabase(t)
}

func runMigrationsForServerTest(t *testing.T) error {
	t.Helper()

	cmd := exec.Command("go", "run", "./cmd/migrate")
	cmd.Dir = "../.."
	cmd.Env = testenv.AppendDatabaseEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
