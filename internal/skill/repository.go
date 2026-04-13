package skill

import (
	"context"
	"database/sql"
	"errors"
)

var errSkillNotFound = errors.New("skill not found")
var errNamespaceRoleMissing = errors.New("namespace role missing")

type Repository interface {
	NamespaceRole(ctx context.Context, namespaceSlug string, userID string) (string, error)
	CreateVersion(ctx context.Context, params CreateVersionParams) (Version, error)
	VersionStorage(ctx context.Context, namespaceSlug string, skillSlug string, version string) (string, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}
