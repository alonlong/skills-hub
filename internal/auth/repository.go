package auth

import (
	"context"
	"database/sql"
	"errors"

	"skillhub/backend/internal/config"
	"skillhub/backend/internal/platform"
)

var errUserNotFound = errors.New("user not found")

type Repository interface {
	FindByUsername(ctx context.Context, username string) (userRecord, error)
	FindByID(ctx context.Context, userID string) (userRecord, error)
	UpdatePasswordHash(ctx context.Context, userID string, passwordHash string) error
	EnsureBootstrapAdmin(ctx context.Context, cfg config.Config) error
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindByUsername(ctx context.Context, username string) (userRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, display_name, COALESCE(email, ''), status
		FROM users
		WHERE username = $1
	`, username)
	return scanUser(row)
}

func (r *SQLRepository) FindByID(ctx context.Context, userID string) (userRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, display_name, COALESCE(email, ''), status
		FROM users
		WHERE id = $1
	`, userID)
	return scanUser(row)
}

func (r *SQLRepository) UpdatePasswordHash(ctx context.Context, userID string, passwordHash string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, passwordHash, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errUserNotFound
	}
	return nil
}

func (r *SQLRepository) EnsureBootstrapAdmin(ctx context.Context, cfg config.Config) error {
	if !cfg.BootstrapAdminEnabled {
		return nil
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, cfg.BootstrapAdminUserID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	hashedPassword, err := platform.HashPassword(cfg.BootstrapAdminPassword)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, password_hash, display_name, email, status)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE')
	`, cfg.BootstrapAdminUserID, cfg.BootstrapAdminUsername, hashedPassword, cfg.BootstrapAdminName, cfg.BootstrapAdminEmail)
	return err
}

func scanUser(row *sql.Row) (userRecord, error) {
	var record userRecord
	err := row.Scan(&record.ID, &record.Username, &record.PasswordHash, &record.DisplayName, &record.Email, &record.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return userRecord{}, errUserNotFound
	}
	if err != nil {
		return userRecord{}, err
	}
	return record, nil
}
