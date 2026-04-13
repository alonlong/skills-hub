package integration_test

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"testing"

	"skillhub/backend/internal/store/postgres"
	"skillhub/backend/internal/testenv"
)

func TestMigrationsCreateCoreTables(t *testing.T) {
	db := openTestDatabase(t)
	if err := runMigrations(t); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	for _, table := range []string{"users", "namespaces", "skills", "skill_versions", "review_tasks"} {
		exists, err := postgres.TableExists(context.Background(), db, table)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("expected table %s to exist", table)
		}
	}
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	return testenv.OpenDatabase(t)
}

func runMigrations(t *testing.T) error {
	t.Helper()

	cmd := exec.Command("go", "run", "./cmd/migrate")
	cmd.Dir = "../.."
	cmd.Env = testenv.AppendDatabaseEnv(os.Environ())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
