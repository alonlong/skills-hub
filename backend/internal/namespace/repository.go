package namespace

import (
	"context"
	"database/sql"
	"errors"
)

var errNamespaceNotFound = errors.New("namespace not found")
var errUserNotFound = errors.New("user not found")

type Repository interface {
	Create(ctx context.Context, namespace Namespace) error
	AddMember(ctx context.Context, namespaceID string, userID string, role string) error
	GetBySlug(ctx context.Context, slug string) (Namespace, error)
	GetMemberRole(ctx context.Context, namespaceID string, userID string) (string, error)
	LookupUserIDByUsername(ctx context.Context, username string) (string, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, namespace Namespace) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO namespaces (id, slug, display_name, type, description, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, namespace.ID, namespace.Slug, namespace.DisplayName, namespace.Type, namespace.Description, namespace.Status, namespace.CreatedBy)
	return err
}

func (r *SQLRepository) AddMember(ctx context.Context, namespaceID string, userID string, role string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO namespace_members (id, namespace_id, user_id, role)
		VALUES (gen_random_uuid(), $1, $2, $3)
		ON CONFLICT (namespace_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, namespaceID, userID, role)
	return err
}

func (r *SQLRepository) GetBySlug(ctx context.Context, slug string) (Namespace, error) {
	var namespace Namespace
	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, display_name, type, description, status, created_by
		FROM namespaces
		WHERE slug = $1
	`, slug).Scan(
		&namespace.ID,
		&namespace.Slug,
		&namespace.DisplayName,
		&namespace.Type,
		&namespace.Description,
		&namespace.Status,
		&namespace.CreatedBy,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Namespace{}, errNamespaceNotFound
	}
	return namespace, err
}

func (r *SQLRepository) GetMemberRole(ctx context.Context, namespaceID string, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role
		FROM namespace_members
		WHERE namespace_id = $1 AND user_id = $2
	`, namespaceID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errNamespaceNotFound
	}
	return role, err
}

func (r *SQLRepository) LookupUserIDByUsername(ctx context.Context, username string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = $1`, username).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errUserNotFound
	}
	return userID, err
}
