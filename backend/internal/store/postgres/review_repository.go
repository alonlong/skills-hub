package postgres

import (
	"context"
	"database/sql"
	"errors"

	"skillhub/backend/internal/admin"
)

type ReviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) GetTask(ctx context.Context, taskID string) (admin.ReviewTask, error) {
	var task admin.ReviewTask
	err := r.db.QueryRowContext(ctx, `
		SELECT id, skill_version_id, namespace_id, status
		FROM review_tasks
		WHERE id = $1
	`, taskID).Scan(&task.ID, &task.SkillVersionID, &task.NamespaceID, &task.Status)
	return task, err
}

func (r *ReviewRepository) Decide(ctx context.Context, taskID string, reviewerID string, comment string, decision string, nextVersionStatus string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var skillVersionID string
	err = tx.QueryRowContext(ctx, `SELECT skill_version_id FROM review_tasks WHERE id = $1`, taskID).Scan(&skillVersionID)
	if err != nil {
		return err
	}

	status := "APPROVED"
	if decision != "APPROVE" {
		status = "REJECTED"
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE review_tasks
		SET status = $2, decision = $3, reviewer_id = $4, comment = $5, reviewed_at = NOW()
		WHERE id = $1
	`, taskID, status, decision, reviewerID, comment); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE skill_versions
		SET status = $2
		WHERE id = $1
	`, skillVersionID, nextVersionStatus)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = errors.New("skill version not found")
		return err
	}

	err = tx.Commit()
	return err
}
