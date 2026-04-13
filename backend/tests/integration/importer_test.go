package integration_test

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	internalimporter "skillhub/backend/internal/importer"
)

func TestImporterCopiesKeptUsersNamespacesAndSkills(t *testing.T) {
	legacyDB, targetDB := openLegacyAndNewDatabases(t)
	if err := runMigrations(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	resetImportedTables(t, targetDB)
	createLegacyTables(t, legacyDB)

	legacyPackages := t.TempDir()
	newPackages := t.TempDir()
	seedLegacyCoreData(t, legacyDB, legacyPackages)

	if err := internalimporter.Run(legacyDB, targetDB, legacyPackages, newPackages); err != nil {
		t.Fatalf("run importer: %v", err)
	}

	assertRowCount(t, targetDB, "users", 2)
	assertRowCount(t, targetDB, "namespaces", 1)
	assertRowCount(t, targetDB, "skills", 1)
	assertRowCount(t, targetDB, "review_tasks", 1)
	assertCopiedPackage(t, legacyPackages, newPackages, "team-ai/email-helper/1.0.0.zip")
}

func openLegacyAndNewDatabases(t *testing.T) (*sql.DB, *sql.DB) {
	t.Helper()
	return openTestDatabase(t), openTestDatabase(t)
}

func resetImportedTables(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		TRUNCATE TABLE
			search_documents,
			review_tasks,
			skill_tags,
			skill_versions,
			skills,
			namespace_members,
			namespaces,
			users
		RESTART IDENTITY CASCADE
	`); err != nil {
		t.Fatalf("truncate imported tables: %v", err)
	}
}

func createLegacyTables(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		DROP TABLE IF EXISTS legacy_search_documents;
		DROP TABLE IF EXISTS legacy_review_tasks;
		DROP TABLE IF EXISTS legacy_skill_versions;
		DROP TABLE IF EXISTS legacy_skills;
		DROP TABLE IF EXISTS legacy_namespace_members;
		DROP TABLE IF EXISTS legacy_namespaces;
		DROP TABLE IF EXISTS legacy_users;

		CREATE TABLE legacy_users (
			id UUID PRIMARY KEY,
			username TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			display_name TEXT NOT NULL,
			email TEXT,
			status TEXT NOT NULL
		);

		CREATE TABLE legacy_namespaces (
			id UUID PRIMARY KEY,
			slug TEXT NOT NULL,
			display_name TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			created_by UUID NOT NULL
		);

		CREATE TABLE legacy_namespace_members (
			id UUID PRIMARY KEY,
			namespace_id UUID NOT NULL,
			user_id UUID NOT NULL,
			role TEXT NOT NULL
		);

		CREATE TABLE legacy_skills (
			id UUID PRIMARY KEY,
			namespace_id UUID NOT NULL,
			slug TEXT NOT NULL,
			display_name TEXT NOT NULL,
			summary TEXT NOT NULL,
			owner_id UUID NOT NULL,
			visibility TEXT NOT NULL,
			status TEXT NOT NULL,
			latest_version_id UUID NOT NULL
		);

		CREATE TABLE legacy_skill_versions (
			id UUID PRIMARY KEY,
			skill_id UUID NOT NULL,
			version TEXT NOT NULL,
			status TEXT NOT NULL,
			manifest_json JSONB NOT NULL,
			parsed_metadata_json JSONB NOT NULL,
			storage_path TEXT NOT NULL,
			submitted_by UUID NOT NULL
		);

		CREATE TABLE legacy_review_tasks (
			id UUID PRIMARY KEY,
			skill_version_id UUID NOT NULL,
			namespace_id UUID NOT NULL,
			status TEXT NOT NULL,
			decision TEXT,
			reviewer_id UUID,
			comment TEXT NOT NULL
		);

		CREATE TABLE legacy_search_documents (
			skill_version_id UUID PRIMARY KEY,
			namespace_slug TEXT NOT NULL,
			skill_slug TEXT NOT NULL,
			title TEXT NOT NULL,
			summary TEXT NOT NULL,
			content TEXT NOT NULL,
			visibility TEXT NOT NULL
		);
	`); err != nil {
		t.Fatalf("create legacy tables: %v", err)
	}
}

func seedLegacyCoreData(t *testing.T, db *sql.DB, packageRoot string) {
	t.Helper()

	if _, err := db.Exec(`
		INSERT INTO legacy_users (id, username, password_hash, display_name, email, status) VALUES
			('11111111-1111-1111-1111-111111111111', 'admin', '$2a$10$bootstrap', 'Admin', 'admin@example.com', 'ACTIVE'),
			('22222222-2222-2222-2222-222222222222', 'publisher', '$2a$10$publisher', 'Publisher', 'publisher@example.com', 'ACTIVE');

		INSERT INTO legacy_namespaces (id, slug, display_name, type, description, status, created_by) VALUES
			('33333333-3333-3333-3333-333333333333', 'team-ai', 'Team AI', 'TEAM', 'AI team namespace', 'ACTIVE', '11111111-1111-1111-1111-111111111111');

		INSERT INTO legacy_namespace_members (id, namespace_id, user_id, role) VALUES
			('44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'OWNER'),
			('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'MEMBER');

		INSERT INTO legacy_skills (id, namespace_id, slug, display_name, summary, owner_id, visibility, status, latest_version_id) VALUES
			('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', 'email-helper', 'Email Helper', 'Send and summarize email', '22222222-2222-2222-2222-222222222222', 'PUBLIC', 'PUBLISHED', '77777777-7777-7777-7777-777777777777');

		INSERT INTO legacy_skill_versions (id, skill_id, version, status, manifest_json, parsed_metadata_json, storage_path, submitted_by) VALUES
			('77777777-7777-7777-7777-777777777777', '66666666-6666-6666-6666-666666666666', '1.0.0', 'PUBLISHED', '{"version":"1.0.0"}', '{"summary":"Send and summarize email"}', 'team-ai/email-helper/1.0.0.zip', '22222222-2222-2222-2222-222222222222');

		INSERT INTO legacy_review_tasks (id, skill_version_id, namespace_id, status, decision, reviewer_id, comment) VALUES
			('88888888-8888-8888-8888-888888888888', '77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', 'APPROVED', 'APPROVE', '11111111-1111-1111-1111-111111111111', 'Looks good');

		INSERT INTO legacy_search_documents (skill_version_id, namespace_slug, skill_slug, title, summary, content, visibility) VALUES
			('77777777-7777-7777-7777-777777777777', 'team-ai', 'email-helper', 'Email Helper', 'Send and summarize email', 'email helper content', 'PUBLIC');
	`); err != nil {
		t.Fatalf("seed legacy rows: %v", err)
	}

	packagePath := filepath.Join(packageRoot, "team-ai", "email-helper", "1.0.0.zip")
	if err := os.MkdirAll(filepath.Dir(packagePath), 0o755); err != nil {
		t.Fatalf("mkdir legacy package path: %v", err)
	}
	if err := os.WriteFile(packagePath, []byte("legacy-package"), 0o644); err != nil {
		t.Fatalf("write legacy package: %v", err)
	}
}

func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("expected %d rows in %s, got %d", want, table, got)
	}
}

func assertCopiedPackage(t *testing.T, legacyRoot string, newRoot string, relativePath string) {
	t.Helper()

	legacyBytes, err := os.ReadFile(filepath.Join(legacyRoot, relativePath))
	if err != nil {
		t.Fatalf("read legacy package: %v", err)
	}
	newBytes, err := os.ReadFile(filepath.Join(newRoot, relativePath))
	if err != nil {
		t.Fatalf("read imported package: %v", err)
	}
	if !bytes.Equal(legacyBytes, newBytes) {
		t.Fatal("imported package does not match legacy package")
	}
}
