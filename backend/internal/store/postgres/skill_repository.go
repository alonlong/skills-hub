package postgres

import (
	"context"
	"database/sql"
	"errors"

	"skillhub/backend/internal/skill"
)

type SkillRepository struct {
	db *sql.DB
}

func NewSkillRepository(db *sql.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) NamespaceRole(ctx context.Context, namespaceSlug string, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT nm.role
		FROM namespace_members nm
		JOIN namespaces n ON n.id = nm.namespace_id
		WHERE n.slug = $1 AND nm.user_id = $2
	`, namespaceSlug, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", skill.ErrNamespaceRoleMissing()
	}
	return role, err
}

func (r *SkillRepository) CreateVersion(ctx context.Context, params skill.CreateVersionParams) (skill.Version, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return skill.Version{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var namespaceID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM namespaces WHERE slug = $1`, params.NamespaceSlug).Scan(&namespaceID)
	if err != nil {
		return skill.Version{}, err
	}

	var skillID string
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM skills
		WHERE namespace_id = $1 AND slug = $2
	`, namespaceID, params.SkillSlug).Scan(&skillID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO skills (id, namespace_id, slug, display_name, summary, owner_id, visibility, status)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'NAMESPACE_ONLY', 'ACTIVE')
			RETURNING id
		`, namespaceID, params.SkillSlug, params.DisplayName, params.Summary, params.SubmittedBy).Scan(&skillID)
	}
	if err != nil {
		return skill.Version{}, err
	}

	var version skill.Version
	err = tx.QueryRowContext(ctx, `
		INSERT INTO skill_versions (id, skill_id, version, status, manifest_json, parsed_metadata_json, storage_path, submitted_by)
		VALUES (gen_random_uuid(), $1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)
		RETURNING id, skill_id, version, status, manifest_json::text, parsed_metadata_json::text, storage_path
	`, skillID, params.Version, params.Status, params.ManifestJSON, params.MetadataJSON, params.StoragePath, params.SubmittedBy).Scan(
		&version.ID,
		&version.SkillID,
		&version.Version,
		&version.Status,
		&version.ManifestJSON,
		&version.MetadataJSON,
		&version.StoragePath,
	)
	if err != nil {
		return skill.Version{}, err
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE skills
		SET latest_version_id = $1, updated_at = NOW()
		WHERE id = $2
	`, version.ID, skillID); err != nil {
		return skill.Version{}, err
	}

	if err = tx.Commit(); err != nil {
		return skill.Version{}, err
	}
	return version, nil
}

func (r *SkillRepository) VersionStorage(ctx context.Context, namespaceSlug string, skillSlug string, version string) (string, error) {
	var storagePath string
	err := r.db.QueryRowContext(ctx, `
		SELECT sv.storage_path
		FROM skill_versions sv
		JOIN skills s ON s.id = sv.skill_id
		JOIN namespaces n ON n.id = s.namespace_id
		WHERE n.slug = $1 AND s.slug = $2 AND sv.version = $3
	`, namespaceSlug, skillSlug, version).Scan(&storagePath)
	if errors.Is(err, sql.ErrNoRows) {
		return "", skill.ErrSkillNotFound()
	}
	return storagePath, err
}
