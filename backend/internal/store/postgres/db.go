package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// TableExists reports whether the named table exists in the public schema.
func TableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`,
		table,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query table existence for %s: %w", table, err)
	}
	return exists, nil
}
