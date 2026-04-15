package label

import "time"

// Translation matches the web AdminLabelInput JSON shape.
type Translation struct {
	Locale      string `json:"locale"`
	DisplayName string `json:"displayName"`
}

// Definition is returned by GET /api/v1/admin/labels and mutations.
type Definition struct {
	Slug            string        `json:"slug"`
	Type            string        `json:"type"`
	VisibleInFilter bool          `json:"visibleInFilter"`
	SortOrder       int           `json:"sortOrder"`
	Translations    []Translation `json:"translations"`
	CreatedAt       *time.Time    `json:"createdAt,omitempty"`
}

// CreateInput is the POST /api/v1/admin/labels body.
type CreateInput struct {
	Slug              string        `json:"slug"`
	Type              string        `json:"type"`
	VisibleInFilter   bool          `json:"visibleInFilter"`
	SortOrder         int           `json:"sortOrder"`
	Translations      []Translation `json:"translations"`
}

// UpdateInput is the PUT /api/v1/admin/labels/{slug} body (no slug field).
type UpdateInput struct {
	Type            string        `json:"type"`
	VisibleInFilter bool          `json:"visibleInFilter"`
	SortOrder       int           `json:"sortOrder"`
	Translations    []Translation `json:"translations"`
}

// SortOrderItem is one entry in PUT /api/v1/admin/labels/sort-order.
type SortOrderItem struct {
	Slug      string `json:"slug"`
	SortOrder int    `json:"sortOrder"`
}
