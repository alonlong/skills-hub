package importer

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type userRow struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	Email        sql.NullString
	Status       string
}

type namespaceRow struct {
	ID          string
	Slug        string
	DisplayName string
	Type        string
	Description string
	Status      string
	CreatedBy   string
}

type namespaceMemberRow struct {
	ID          string
	NamespaceID string
	UserID      string
	Role        string
}

type skillRow struct {
	ID              string
	NamespaceID     string
	Slug            string
	DisplayName     string
	Summary         string
	OwnerID         string
	Visibility      string
	Status          string
	LatestVersionID string
}

type skillVersionRow struct {
	ID                 string
	SkillID            string
	Version            string
	Status             string
	ManifestJSON       []byte
	ParsedMetadataJSON []byte
	StoragePath        string
	SubmittedBy        string
}

type reviewTaskRow struct {
	ID             string
	SkillVersionID string
	NamespaceID    string
	Status         string
	Decision       sql.NullString
	ReviewerID     sql.NullString
	Comment        string
}

type searchDocumentRow struct {
	SkillVersionID string
	NamespaceSlug  string
	SkillSlug      string
	Title          string
	Summary        string
	Content        string
	Visibility     string
}

func Run(legacyDB *sql.DB, targetDB *sql.DB, legacyPackageRoot string, packageRoot string) error {
	users, err := loadUsers(legacyDB)
	if err != nil {
		return fmt.Errorf("load users: %w", err)
	}
	namespaces, err := loadNamespaces(legacyDB)
	if err != nil {
		return fmt.Errorf("load namespaces: %w", err)
	}
	namespaceMembers, err := loadNamespaceMembers(legacyDB)
	if err != nil {
		return fmt.Errorf("load namespace members: %w", err)
	}
	skills, err := loadSkills(legacyDB)
	if err != nil {
		return fmt.Errorf("load skills: %w", err)
	}
	skillVersions, err := loadSkillVersions(legacyDB)
	if err != nil {
		return fmt.Errorf("load skill versions: %w", err)
	}
	reviewTasks, err := loadReviewTasks(legacyDB)
	if err != nil {
		return fmt.Errorf("load review tasks: %w", err)
	}
	searchDocuments, err := loadSearchDocuments(legacyDB)
	if err != nil {
		return fmt.Errorf("load search documents: %w", err)
	}

	tx, err := targetDB.Begin()
	if err != nil {
		return fmt.Errorf("begin target transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := insertUsers(tx, users); err != nil {
		return fmt.Errorf("insert users: %w", err)
	}
	if err := insertNamespaces(tx, namespaces); err != nil {
		return fmt.Errorf("insert namespaces: %w", err)
	}
	if err := insertNamespaceMembers(tx, namespaceMembers); err != nil {
		return fmt.Errorf("insert namespace members: %w", err)
	}
	if err := insertSkills(tx, skills); err != nil {
		return fmt.Errorf("insert skills: %w", err)
	}
	if err := insertSkillVersions(tx, skillVersions); err != nil {
		return fmt.Errorf("insert skill versions: %w", err)
	}
	if err := updateLatestVersions(tx, skills); err != nil {
		return fmt.Errorf("update latest versions: %w", err)
	}
	if err := insertReviewTasks(tx, reviewTasks); err != nil {
		return fmt.Errorf("insert review tasks: %w", err)
	}
	if err := insertSearchDocuments(tx, searchDocuments); err != nil {
		return fmt.Errorf("insert search documents: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit target transaction: %w", err)
	}

	for _, version := range skillVersions {
		if err := copyPackageFile(legacyPackageRoot, packageRoot, version.StoragePath); err != nil {
			return fmt.Errorf("copy package %s: %w", version.StoragePath, err)
		}
	}

	return nil
}

func loadUsers(db *sql.DB) ([]userRow, error) {
	rows, err := db.Query(`SELECT id, username, password_hash, display_name, email, status FROM legacy_users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userRow
	for rows.Next() {
		var row userRow
		if err := rows.Scan(&row.ID, &row.Username, &row.PasswordHash, &row.DisplayName, &row.Email, &row.Status); err != nil {
			return nil, err
		}
		users = append(users, row)
	}
	return users, rows.Err()
}

func loadNamespaces(db *sql.DB) ([]namespaceRow, error) {
	rows, err := db.Query(`SELECT id, slug, display_name, type, description, status, created_by FROM legacy_namespaces ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var namespaces []namespaceRow
	for rows.Next() {
		var row namespaceRow
		if err := rows.Scan(&row.ID, &row.Slug, &row.DisplayName, &row.Type, &row.Description, &row.Status, &row.CreatedBy); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, row)
	}
	return namespaces, rows.Err()
}

func loadNamespaceMembers(db *sql.DB) ([]namespaceMemberRow, error) {
	rows, err := db.Query(`SELECT id, namespace_id, user_id, role FROM legacy_namespace_members ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []namespaceMemberRow
	for rows.Next() {
		var row namespaceMemberRow
		if err := rows.Scan(&row.ID, &row.NamespaceID, &row.UserID, &row.Role); err != nil {
			return nil, err
		}
		members = append(members, row)
	}
	return members, rows.Err()
}

func loadSkills(db *sql.DB) ([]skillRow, error) {
	rows, err := db.Query(`SELECT id, namespace_id, slug, display_name, summary, owner_id, visibility, status, latest_version_id FROM legacy_skills ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []skillRow
	for rows.Next() {
		var row skillRow
		if err := rows.Scan(&row.ID, &row.NamespaceID, &row.Slug, &row.DisplayName, &row.Summary, &row.OwnerID, &row.Visibility, &row.Status, &row.LatestVersionID); err != nil {
			return nil, err
		}
		skills = append(skills, row)
	}
	return skills, rows.Err()
}

func loadSkillVersions(db *sql.DB) ([]skillVersionRow, error) {
	rows, err := db.Query(`SELECT id, skill_id, version, status, manifest_json, parsed_metadata_json, storage_path, submitted_by FROM legacy_skill_versions ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []skillVersionRow
	for rows.Next() {
		var row skillVersionRow
		if err := rows.Scan(&row.ID, &row.SkillID, &row.Version, &row.Status, &row.ManifestJSON, &row.ParsedMetadataJSON, &row.StoragePath, &row.SubmittedBy); err != nil {
			return nil, err
		}
		versions = append(versions, row)
	}
	return versions, rows.Err()
}

func loadReviewTasks(db *sql.DB) ([]reviewTaskRow, error) {
	rows, err := db.Query(`SELECT id, skill_version_id, namespace_id, status, decision, reviewer_id, comment FROM legacy_review_tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []reviewTaskRow
	for rows.Next() {
		var row reviewTaskRow
		if err := rows.Scan(&row.ID, &row.SkillVersionID, &row.NamespaceID, &row.Status, &row.Decision, &row.ReviewerID, &row.Comment); err != nil {
			return nil, err
		}
		tasks = append(tasks, row)
	}
	return tasks, rows.Err()
}

func loadSearchDocuments(db *sql.DB) ([]searchDocumentRow, error) {
	rows, err := db.Query(`SELECT skill_version_id, namespace_slug, skill_slug, title, summary, content, visibility FROM legacy_search_documents ORDER BY skill_slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []searchDocumentRow
	for rows.Next() {
		var row searchDocumentRow
		if err := rows.Scan(&row.SkillVersionID, &row.NamespaceSlug, &row.SkillSlug, &row.Title, &row.Summary, &row.Content, &row.Visibility); err != nil {
			return nil, err
		}
		docs = append(docs, row)
	}
	return docs, rows.Err()
}

func insertUsers(tx *sql.Tx, users []userRow) error {
	for _, row := range users {
		if _, err := tx.Exec(`
			INSERT INTO users (id, username, password_hash, display_name, email, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, row.ID, row.Username, row.PasswordHash, row.DisplayName, nullableString(row.Email), row.Status); err != nil {
			return err
		}
	}
	return nil
}

func insertNamespaces(tx *sql.Tx, namespaces []namespaceRow) error {
	for _, row := range namespaces {
		if _, err := tx.Exec(`
			INSERT INTO namespaces (id, slug, display_name, type, description, status, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, row.ID, row.Slug, row.DisplayName, row.Type, row.Description, row.Status, row.CreatedBy); err != nil {
			return err
		}
	}
	return nil
}

func insertNamespaceMembers(tx *sql.Tx, members []namespaceMemberRow) error {
	for _, row := range members {
		if _, err := tx.Exec(`
			INSERT INTO namespace_members (id, namespace_id, user_id, role)
			VALUES ($1, $2, $3, $4)
		`, row.ID, row.NamespaceID, row.UserID, row.Role); err != nil {
			return err
		}
	}
	return nil
}

func insertSkills(tx *sql.Tx, skills []skillRow) error {
	for _, row := range skills {
		if _, err := tx.Exec(`
			INSERT INTO skills (id, namespace_id, slug, display_name, summary, owner_id, visibility, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, row.ID, row.NamespaceID, row.Slug, row.DisplayName, row.Summary, row.OwnerID, row.Visibility, row.Status); err != nil {
			return err
		}
	}
	return nil
}

func insertSkillVersions(tx *sql.Tx, versions []skillVersionRow) error {
	for _, row := range versions {
		if _, err := tx.Exec(`
			INSERT INTO skill_versions (id, skill_id, version, status, manifest_json, parsed_metadata_json, storage_path, submitted_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, row.ID, row.SkillID, row.Version, row.Status, row.ManifestJSON, row.ParsedMetadataJSON, row.StoragePath, row.SubmittedBy); err != nil {
			return err
		}
	}
	return nil
}

func updateLatestVersions(tx *sql.Tx, skills []skillRow) error {
	for _, row := range skills {
		if _, err := tx.Exec(`UPDATE skills SET latest_version_id = $2 WHERE id = $1`, row.ID, row.LatestVersionID); err != nil {
			return err
		}
	}
	return nil
}

func insertReviewTasks(tx *sql.Tx, tasks []reviewTaskRow) error {
	for _, row := range tasks {
		if _, err := tx.Exec(`
			INSERT INTO review_tasks (id, skill_version_id, namespace_id, status, decision, reviewer_id, comment)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, row.ID, row.SkillVersionID, row.NamespaceID, row.Status, nullableString(row.Decision), nullableString(row.ReviewerID), row.Comment); err != nil {
			return err
		}
	}
	return nil
}

func insertSearchDocuments(tx *sql.Tx, docs []searchDocumentRow) error {
	for _, row := range docs {
		if _, err := tx.Exec(`
			INSERT INTO search_documents (skill_version_id, namespace_slug, skill_slug, title, summary, content, visibility)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, row.SkillVersionID, row.NamespaceSlug, row.SkillSlug, row.Title, row.Summary, row.Content, row.Visibility); err != nil {
			return err
		}
	}
	return nil
}

func copyPackageFile(legacyRoot string, targetRoot string, storagePath string) error {
	sourcePath := filepath.Join(legacyRoot, filepath.Clean(storagePath))
	targetPath := filepath.Join(targetRoot, filepath.Clean(storagePath))

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}
	return nil
}

func nullableString(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}
