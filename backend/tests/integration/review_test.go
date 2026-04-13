package integration_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/admin"
	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/skill"
	localstorage "skillhub/backend/internal/storage/local"
	"skillhub/backend/internal/store/postgres"
	"skillhub/backend/internal/testenv"
)

func TestAdminCanApprovePendingReview(t *testing.T) {
	app := newReviewTestApp(t)

	taskID := seedPendingReview(t, app.DB, "team-ai", "email-helper", "1.0.0")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews/"+taskID+"/approve", strings.NewReader(`{"comment":"looks good"}`))
	req.Header.Set("Authorization", "Bearer "+app.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	app.Router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}

	assertReviewStatus(t, app.DB, taskID, "APPROVED")
	assertVersionStatus(t, app.DB, "team-ai", "email-helper", "1.0.0", "PUBLISHED")
}

type reviewTestApp struct {
	DB         *sql.DB
	Router     http.Handler
	AdminToken string
}

func newReviewTestApp(t *testing.T) reviewTestApp {
	t.Helper()

	storageDir := t.TempDir()
	t.Setenv("DATABASE_URL", testenv.DatabaseURL())
	t.Setenv("BACKEND_JWT_SECRET", "test-secret")
	t.Setenv("STORAGE_BASE_PATH", storageDir)

	db := openAuthDatabase(t)
	if err := runMigrations(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	resetAuthTables(t, db)
	seedUser(t, db, "99999999-9999-9999-9999-999999999999", "admin")

	cfg := config.Load()
	authRepo := auth.NewSQLRepository(db)
	authService := auth.NewService(authRepo, auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	namespaceService := namespace.NewService(namespace.NewSQLRepository(db))
	skillService := skill.NewService(postgres.NewSkillRepository(db), localstorage.NewService(cfg.StorageBasePath))
	adminService := admin.NewService(postgres.NewReviewRepository(db))

	token, err := auth.NewJWTIssuer(cfg.JWTSecret).Issue(auth.User{
		UserID:        "99999999-9999-9999-9999-999999999999",
		Username:      "admin",
		DisplayName:   "Admin",
		PlatformRoles: []string{"SUPER_ADMIN"},
	})
	if err != nil {
		t.Fatalf("issue admin token: %v", err)
	}

	return reviewTestApp{
		DB: db,
		Router: aphttp.NewRouter(cfg, aphttp.Dependencies{
			AuthService:      authService,
			NamespaceService: namespaceService,
			SkillService:     skillService,
			AdminService:     adminService,
		}),
		AdminToken: token,
	}
}

func seedPendingReview(t *testing.T, db *sql.DB, namespaceSlug string, skillSlug string, version string) string {
	t.Helper()

	seedNamespaceOwner(t, db, namespaceSlug, "admin")
	archive := sampleSkillArchive(t)
	cfg := config.Load()
	service := skill.NewService(postgres.NewSkillRepository(db), localstorage.NewService(cfg.StorageBasePath))
	issuedBy, err := auth.NewJWTIssuer(cfg.JWTSecret).Issue(auth.User{
		UserID:        "99999999-9999-9999-9999-999999999999",
		Username:      "admin",
		DisplayName:   "Admin",
		PlatformRoles: []string{"SUPER_ADMIN"},
	})
	if err != nil {
		t.Fatalf("issue auth token: %v", err)
	}
	_ = issuedBy

	published, err := service.PublishVersion(t.Context(), auth.Actor{
		UserID:        "99999999-9999-9999-9999-999999999999",
		Username:      "admin",
		PlatformRoles: []string{"SUPER_ADMIN"},
	}, skill.PublishInput{
		NamespaceSlug: namespaceSlug,
		SkillSlug:     skillSlug,
		Archive:       archive,
	})
	if err != nil {
		t.Fatalf("publish skill version: %v", err)
	}

	var taskID string
	if err := db.QueryRow(`
		INSERT INTO review_tasks (id, skill_version_id, namespace_id, status, comment)
		SELECT gen_random_uuid(), sv.id, n.id, 'PENDING_REVIEW', ''
		FROM skill_versions sv
		JOIN skills s ON s.id = sv.skill_id
		JOIN namespaces n ON n.id = s.namespace_id
		WHERE n.slug = $1 AND s.slug = $2 AND sv.version = $3
		RETURNING id
	`, namespaceSlug, skillSlug, version).Scan(&taskID); err != nil {
		t.Fatalf("seed review task: %v", err)
	}

	if published.Version != version {
		t.Fatalf("unexpected published version %s", published.Version)
	}

	return taskID
}

func assertReviewStatus(t *testing.T, db *sql.DB, taskID string, expected string) {
	t.Helper()

	var status string
	if err := db.QueryRow(`SELECT status FROM review_tasks WHERE id = $1`, taskID).Scan(&status); err != nil {
		t.Fatalf("query review status: %v", err)
	}
	if status != expected {
		t.Fatalf("expected review status %s, got %s", expected, status)
	}
}

func assertVersionStatus(t *testing.T, db *sql.DB, namespaceSlug string, skillSlug string, version string, expected string) {
	t.Helper()

	var status string
	if err := db.QueryRow(`
		SELECT sv.status
		FROM skill_versions sv
		JOIN skills s ON s.id = sv.skill_id
		JOIN namespaces n ON n.id = s.namespace_id
		WHERE n.slug = $1 AND s.slug = $2 AND sv.version = $3
	`, namespaceSlug, skillSlug, version).Scan(&status); err != nil {
		t.Fatalf("query version status: %v", err)
	}
	if status != expected {
		t.Fatalf("expected version status %s, got %s", expected, status)
	}
}
