package postgres

import (
	"context"
	"database/sql"

	"skillhub/backend/internal/search"
)

type SearchRepository struct {
	db *sql.DB
}

func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) Search(ctx context.Context, query search.Query) ([]search.Result, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT namespace_slug, skill_slug, title, summary,
		       ts_rank(document, websearch_to_tsquery('simple', $1)) AS rank
		FROM search_documents
		WHERE document @@ websearch_to_tsquery('simple', $1)
		  AND ($2 = '' OR namespace_slug = $2)
		ORDER BY rank DESC, title ASC
		LIMIT 20
	`, query.Text, query.NamespaceSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []search.Result{}
	for rows.Next() {
		var item search.Result
		if err := rows.Scan(&item.NamespaceSlug, &item.Slug, &item.Title, &item.Summary, &item.Rank); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
