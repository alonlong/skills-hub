package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/platform"
	"skillhub/backend/internal/testenv"
)

func TestPasswordLoginReturnsJWTAndMeEndpoint(t *testing.T) {
	app := newTestApp(t)
	seedBootstrapAdmin(t, app.DB)

	loginBody := []byte(`{"username":"admin","password":"ChangeMe!2026"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()

	app.Router.ServeHTTP(loginRes, loginReq)
	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginRes.Code)
	}

	token := extractAccessToken(t, loginRes.Body.Bytes())
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRes := httptest.NewRecorder()

	app.Router.ServeHTTP(meRes, meReq)
	if meRes.Code != http.StatusOK {
		t.Fatalf("expected me 200, got %d", meRes.Code)
	}
}

type testApp struct {
	DB     *sql.DB
	Router http.Handler
}

func newTestApp(t *testing.T) testApp {
	t.Helper()

	t.Setenv("DATABASE_URL", testenv.DatabaseURL())
	t.Setenv("BACKEND_JWT_SECRET", "test-secret")

	db := openAuthDatabase(t)
	if err := runMigrations(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	resetAuthTables(t, db)

	cfg := config.Load()
	repo := auth.NewSQLRepository(db)
	service := auth.NewService(repo, auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	deps := aphttp.Dependencies{AuthService: service}

	return testApp{
		DB:     db,
		Router: aphttp.NewRouter(cfg, deps),
	}
}

func openAuthDatabase(t *testing.T) *sql.DB {
	t.Helper()
	return testenv.OpenDatabase(t)
}

func resetAuthTables(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		TRUNCATE TABLE review_tasks, search_documents, skill_tags, skill_versions, skills, namespace_members, namespaces, users
		RESTART IDENTITY CASCADE
	`); err != nil {
		t.Fatalf("reset tables: %v", err)
	}
}

func seedBootstrapAdmin(t *testing.T, db *sql.DB) {
	t.Helper()

	hashedPassword, err := platform.HashPassword("ChangeMe!2026")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash, display_name, email, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "11111111-1111-1111-1111-111111111111", "admin", hashedPassword, "Admin", "admin@example.com", "ACTIVE"); err != nil {
		t.Fatalf("seed bootstrap admin: %v", err)
	}
}

func extractAccessToken(t *testing.T, body []byte) string {
	t.Helper()

	var payload struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode token payload: %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	return payload.AccessToken
}
