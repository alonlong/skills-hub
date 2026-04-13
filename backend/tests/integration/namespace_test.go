package integration_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/platform"
	"skillhub/backend/internal/testenv"
)

func TestOwnerCanCreateNamespaceAndAddMember(t *testing.T) {
	app := newAuthenticatedApp(t, "owner-user")

	createRes := postJSON(t, app.Router, "/api/v1/namespaces", `{
	  "slug":"platform",
	  "displayName":"Platform",
	  "type":"TEAM",
	  "description":"Platform team skills"
	}`, app.Token)
	assertStatus(t, createRes, http.StatusCreated)

	memberRes := postJSON(t, app.Router, "/api/v1/namespaces/platform/members", `{
	  "username":"reviewer",
	  "role":"ADMIN"
	}`, app.Token)
	assertStatus(t, memberRes, http.StatusCreated)
}

type authenticatedApp struct {
	DB     *sql.DB
	Router http.Handler
	Token  string
}

func newAuthenticatedApp(t *testing.T, username string) authenticatedApp {
	t.Helper()

	t.Setenv("DATABASE_URL", testenv.DatabaseURL())
	t.Setenv("BACKEND_JWT_SECRET", "test-secret")

	db := openAuthDatabase(t)
	if err := runMigrations(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	resetAuthTables(t, db)
	seedUser(t, db, "11111111-1111-1111-1111-111111111111", username)
	seedUser(t, db, "22222222-2222-2222-2222-222222222222", "reviewer")

	cfg := config.Load()
	authRepo := auth.NewSQLRepository(db)
	authService := auth.NewService(authRepo, auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	namespaceRepo := namespace.NewSQLRepository(db)
	namespaceService := namespace.NewService(namespaceRepo)

	token, err := auth.NewJWTIssuer(cfg.JWTSecret).Issue(auth.User{
		UserID:        "11111111-1111-1111-1111-111111111111",
		Username:      username,
		DisplayName:   username,
		PlatformRoles: []string{},
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	return authenticatedApp{
		DB: db,
		Router: aphttp.NewRouter(cfg, aphttp.Dependencies{
			AuthService:      authService,
			NamespaceService: namespaceService,
		}),
		Token: token,
	}
}

func seedUser(t *testing.T, db *sql.DB, userID string, username string) {
	t.Helper()

	hashedPassword, err := authHashPassword("ChangeMe!2026")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO users (id, username, password_hash, display_name, email, status)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE')
	`, userID, username, hashedPassword, username, username+"@example.com"); err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
}

func authHashPassword(password string) (string, error) {
	return platform.HashPassword(password)
}

func postJSON(t *testing.T, handler http.Handler, path string, body string, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func assertStatus(t *testing.T, res *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if res.Code != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, res.Code, res.Body.String())
	}
}
