package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"skillhub/backend/internal/label"
)

type LabelRepository struct {
	db *sql.DB
}

func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

func (r *LabelRepository) List(ctx context.Context) ([]label.Definition, error) {
	defRows, err := r.db.QueryContext(ctx, `
		SELECT slug, type, visible_in_filter, sort_order, created_at
		FROM label_definitions
		ORDER BY sort_order ASC, slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer defRows.Close()

	type defRow struct {
		slug            string
		typ             string
		visibleInFilter bool
		sortOrder       int
		createdAt       time.Time
	}
	var defs []defRow
	for defRows.Next() {
		var d defRow
		if err := defRows.Scan(&d.slug, &d.typ, &d.visibleInFilter, &d.sortOrder, &d.createdAt); err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	if err := defRows.Err(); err != nil {
		return nil, err
	}
	if len(defs) == 0 {
		return []label.Definition{}, nil
	}

	transRows, err := r.db.QueryContext(ctx, `
		SELECT label_slug, locale, display_name
		FROM label_translations
		ORDER BY label_slug ASC, locale ASC
	`)
	if err != nil {
		return nil, err
	}
	defer transRows.Close()

	bySlug := make(map[string][]label.Translation)
	for transRows.Next() {
		var slug, loc, name string
		if err := transRows.Scan(&slug, &loc, &name); err != nil {
			return nil, err
		}
		bySlug[slug] = append(bySlug[slug], label.Translation{Locale: loc, DisplayName: name})
	}
	if err := transRows.Err(); err != nil {
		return nil, err
	}

	out := make([]label.Definition, 0, len(defs))
	for _, d := range defs {
		tr := bySlug[d.slug]
		if tr == nil {
			tr = []label.Translation{}
		}
		ca := d.createdAt
		out = append(out, label.Definition{
			Slug:            d.slug,
			Type:            d.typ,
			VisibleInFilter: d.visibleInFilter,
			SortOrder:       d.sortOrder,
			Translations:    tr,
			CreatedAt:       &ca,
		})
	}
	return out, nil
}

func (r *LabelRepository) Get(ctx context.Context, slug string) (label.Definition, error) {
	var d struct {
		slug            string
		typ             string
		visibleInFilter bool
		sortOrder       int
		createdAt       time.Time
	}
	err := r.db.QueryRowContext(ctx, `
		SELECT slug, type, visible_in_filter, sort_order, created_at
		FROM label_definitions WHERE slug = $1
	`, slug).Scan(&d.slug, &d.typ, &d.visibleInFilter, &d.sortOrder, &d.createdAt)
	if err != nil {
		return label.Definition{}, err
	}

	transRows, err := r.db.QueryContext(ctx, `
		SELECT locale, display_name FROM label_translations WHERE label_slug = $1 ORDER BY locale ASC
	`, slug)
	if err != nil {
		return label.Definition{}, err
	}
	defer transRows.Close()

	var tr []label.Translation
	for transRows.Next() {
		var loc, name string
		if err := transRows.Scan(&loc, &name); err != nil {
			return label.Definition{}, err
		}
		tr = append(tr, label.Translation{Locale: loc, DisplayName: name})
	}
	if err := transRows.Err(); err != nil {
		return label.Definition{}, err
	}
	ca := d.createdAt
	return label.Definition{
		Slug:            d.slug,
		Type:            d.typ,
		VisibleInFilter: d.visibleInFilter,
		SortOrder:       d.sortOrder,
		Translations:    tr,
		CreatedAt:       &ca,
	}, nil
}

func (r *LabelRepository) Create(ctx context.Context, in label.CreateInput) error {
	if len(in.Translations) == 0 {
		return errors.New("translations required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO label_definitions (slug, type, visible_in_filter, sort_order)
		VALUES ($1, $2, $3, $4)
	`, in.Slug, in.Type, in.VisibleInFilter, in.SortOrder)
	if err != nil {
		_ = tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("slug exists")
		}
		return err
	}

	for _, tr := range in.Translations {
		if tr.Locale == "" || tr.DisplayName == "" {
			_ = tx.Rollback()
			return errors.New("invalid translation")
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO label_translations (label_slug, locale, display_name)
			VALUES ($1, $2, $3)
		`, in.Slug, tr.Locale, tr.DisplayName)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *LabelRepository) Update(ctx context.Context, slug string, in label.UpdateInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE label_definitions
		SET type = $2, visible_in_filter = $3, sort_order = $4
		WHERE slug = $1
	`, slug, in.Type, in.VisibleInFilter, in.SortOrder)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if n == 0 {
		_ = tx.Rollback()
		return sql.ErrNoRows
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM label_translations WHERE label_slug = $1`, slug); err != nil {
		_ = tx.Rollback()
		return err
	}
	if len(in.Translations) == 0 {
		_ = tx.Rollback()
		return errors.New("translations required")
	}
	for _, tr := range in.Translations {
		if tr.Locale == "" || tr.DisplayName == "" {
			_ = tx.Rollback()
			return errors.New("invalid translation")
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO label_translations (label_slug, locale, display_name)
			VALUES ($1, $2, $3)
		`, slug, tr.Locale, tr.DisplayName)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *LabelRepository) Delete(ctx context.Context, slug string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM label_definitions WHERE slug = $1`, slug)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *LabelRepository) UpdateSortOrder(ctx context.Context, items []label.SortOrderItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, it := range items {
		_, err = tx.ExecContext(ctx, `
			UPDATE label_definitions SET sort_order = $2 WHERE slug = $1
		`, it.Slug, it.SortOrder)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}
