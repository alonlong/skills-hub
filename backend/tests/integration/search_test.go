package integration_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/search"
	"skillhub/backend/internal/skill"
	localstorage "skillhub/backend/internal/storage/local"
	"skillhub/backend/internal/store/postgres"
	"skillhub/backend/internal/testenv"
)

func TestSearchReturnsVisiblePublishedSkills(t *testing.T) {
	app := newSearchTestApp(t)
	seedPublishedSkill(t, app.DB, "team-ai", "email-helper", "Email Helper", "Send and summarize email")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=email", nil)
	res := httptest.NewRecorder()
	app.Router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"slug":"email-helper"`) {
		t.Fatalf("expected search result to include email-helper, got %s", res.Body.String())
	}
}

type searchTestApp struct {
	DB     *sql.DB
	Router http.Handler
}

func newSearchTestApp(t *testing.T) searchTestApp {
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
	seedUser(t, db, "44444444-4444-4444-4444-444444444444", "searcher")

	cfg := config.Load()
	authService := auth.NewService(auth.NewSQLRepository(db), auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	namespaceService := namespace.NewService(namespace.NewSQLRepository(db))
	skillService := skill.NewService(postgres.NewSkillRepository(db), localstorage.NewService(cfg.StorageBasePath))
	searchService := search.NewService(postgres.NewSearchRepository(db))

	return searchTestApp{
		DB: db,
		Router: aphttp.NewRouter(cfg, aphttp.Dependencies{
			AuthService:      authService,
			NamespaceService: namespaceService,
			SkillService:     skillService,
			SearchService:    searchService,
		}),
	}
}

func seedPublishedSkill(t *testing.T, db *sql.DB, namespaceSlug string, skillSlug string, displayName string, summary string) {
	t.Helper()

	seedNamespaceOwner(t, db, namespaceSlug, "searcher")
	if _, err := db.Exec(`
		INSERT INTO skills (id, namespace_id, slug, display_name, summary, owner_id, visibility, status)
		SELECT gen_random_uuid(), n.id, $2, $3, $4, u.id, 'PUBLIC', 'ACTIVE'
		FROM namespaces n, users u
		WHERE n.slug = $1 AND u.username = 'searcher'
	`, namespaceSlug, skillSlug, displayName, summary); err != nil {
		t.Fatalf("insert skill: %v", err)
	}

	if _, err := db.Exec(`
		WITH inserted_version AS (
			INSERT INTO skill_versions (id, skill_id, version, status, manifest_json, parsed_metadata_json, storage_path, submitted_by)
			SELECT gen_random_uuid(), s.id, '1.0.0', 'PUBLISHED', '{"version":"1.0.0"}'::jsonb, '{"summary":"Send and summarize email"}'::jsonb, '/tmp/fake.zip', u.id
			FROM skills s
			JOIN namespaces n ON n.id = s.namespace_id
			JOIN users u ON u.username = 'searcher'
			WHERE n.slug = $1 AND s.slug = $2 AND u.username = 'searcher'
			RETURNING id, skill_id
		)
		UPDATE skills
		SET latest_version_id = inserted_version.id
		FROM inserted_version
		WHERE skills.id = inserted_version.skill_id
	`, namespaceSlug, skillSlug); err != nil {
		t.Fatalf("insert skill version: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO search_documents (skill_version_id, namespace_slug, skill_slug, title, summary, content, visibility)
		SELECT sv.id, n.slug, s.slug, s.display_name, s.summary, 'Email helper content', s.visibility
		FROM skill_versions sv
		JOIN skills s ON s.id = sv.skill_id
		JOIN namespaces n ON n.id = s.namespace_id
		WHERE n.slug = $1 AND s.slug = $2 AND sv.version = '1.0.0'
	`, namespaceSlug, skillSlug); err != nil {
		t.Fatalf("insert search document: %v", err)
	}
}
