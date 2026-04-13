package testenv

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultDatabaseURL = "postgres://skillhub:skillhub_dev@localhost:5432/skillhub?sslmode=disable"

// DatabaseURL returns the database URL used by integration-style tests.
// SKILLHUB_TEST_DATABASE_URL lets tests run against an isolated database
// without coupling verification to whatever is bound to localhost:5432.
func DatabaseURL() string {
	if value := os.Getenv("SKILLHUB_TEST_DATABASE_URL"); value != "" {
		return value
	}
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	return defaultDatabaseURL
}

func OpenDatabase(tb testing.TB) *sql.DB {
	tb.Helper()

	db, err := sql.Open("pgx", DatabaseURL())
	if err != nil {
		tb.Fatalf("open database: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })
	return db
}

func AppendDatabaseEnv(env []string) []string {
	return append(env, "DATABASE_URL="+DatabaseURL())
}
