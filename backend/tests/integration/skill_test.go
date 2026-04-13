package integration_test

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"skillhub/backend/internal/auth"
	"skillhub/backend/internal/config"
	aphttp "skillhub/backend/internal/http"
	"skillhub/backend/internal/namespace"
	"skillhub/backend/internal/skill"
	localstorage "skillhub/backend/internal/storage/local"
	"skillhub/backend/internal/store/postgres"
	"skillhub/backend/internal/testenv"
)

func TestPublishSkillVersionAndDownloadArchive(t *testing.T) {
	app := newSkillApp(t, "publisher")
	seedNamespaceOwner(t, app.DB, "team-ai", "publisher")
	archive := sampleSkillArchive(t)

	body, contentType := buildMultipartSkillUpload(t, "skill.zip", archive)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/team-ai/email-helper/versions", body)
	req.Header.Set("Authorization", "Bearer "+app.Token)
	req.Header.Set("Content-Type", contentType)
	res := httptest.NewRecorder()

	app.Router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", res.Code, res.Body.String())
	}

	downloadReq := httptest.NewRequest(http.MethodGet, "/api/v1/skills/team-ai/email-helper/versions/1.0.0/download", nil)
	downloadRes := httptest.NewRecorder()
	app.Router.ServeHTTP(downloadRes, downloadReq)
	if downloadRes.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", downloadRes.Code)
	}
	if !bytes.Equal(downloadRes.Body.Bytes(), archive) {
		t.Fatal("downloaded archive does not match uploaded archive")
	}
}

type skillApp struct {
	DB     *sql.DB
	Router http.Handler
	Token  string
}

func newSkillApp(t *testing.T, username string) skillApp {
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
	seedUser(t, db, "11111111-1111-1111-1111-111111111111", username)

	cfg := config.Load()
	authRepo := auth.NewSQLRepository(db)
	authService := auth.NewService(authRepo, auth.NewJWTIssuer(cfg.JWTSecret), cfg)
	namespaceRepo := namespace.NewSQLRepository(db)
	namespaceService := namespace.NewService(namespaceRepo)
	skillRepo := postgres.NewSkillRepository(db)
	storageService := localstorage.NewService(cfg.StorageBasePath)
	skillService := skill.NewService(skillRepo, storageService)

	token, err := auth.NewJWTIssuer(cfg.JWTSecret).Issue(auth.User{
		UserID:        "11111111-1111-1111-1111-111111111111",
		Username:      username,
		DisplayName:   username,
		PlatformRoles: []string{},
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	return skillApp{
		DB: db,
		Router: aphttp.NewRouter(cfg, aphttp.Dependencies{
			AuthService:      authService,
			NamespaceService: namespaceService,
			SkillService:     skillService,
		}),
		Token: token,
	}
}

func seedNamespaceOwner(t *testing.T, db *sql.DB, slug string, username string) {
	t.Helper()

	var userID string
	if err := db.QueryRow(`SELECT id FROM users WHERE username = $1`, username).Scan(&userID); err != nil {
		t.Fatalf("lookup owner user: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO namespaces (id, slug, display_name, type, description, status, created_by)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE', $6)
	`, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", slug, "Team AI", "TEAM", "AI team", userID); err != nil {
		t.Fatalf("seed namespace: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO namespace_members (id, namespace_id, user_id, role)
		VALUES ($1, $2, $3, 'OWNER')
	`, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", userID); err != nil {
		t.Fatalf("seed namespace member: %v", err)
	}
}

func buildMultipartSkillUpload(t *testing.T, filename string, payload []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write upload payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func sampleSkillArchive(t *testing.T) []byte {
	t.Helper()

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	writeZipFile(t, zw, "manifest.json", `{"version":"1.0.0","displayName":"Email Helper","summary":"Send and summarize email"}`)
	writeZipFile(t, zw, "SKILL.md", "# Email Helper\n\nHelps with email workflows.\n")

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func writeZipFile(t *testing.T, zw *zip.Writer, name string, content string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip file %s: %v", name, err)
	}
	if _, err := io.WriteString(w, content); err != nil {
		t.Fatalf("write zip file %s: %v", name, err)
	}
}
