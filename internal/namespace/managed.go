package namespace

import (
	"time"
)

// ManagedNamespaceView is the JSON shape expected by the web app for GET /api/web/me/namespaces.
type ManagedNamespaceView struct {
	ID              string `json:"id"`
	Slug            string `json:"slug"`
	DisplayName     string `json:"displayName"`
	Type            string `json:"type"`
	Description     string `json:"description,omitempty"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
	CreatedBy       string `json:"createdBy,omitempty"`
	CurrentUserRole string `json:"currentUserRole"`
	Immutable       bool   `json:"immutable"`
	CanFreeze       bool   `json:"canFreeze"`
	CanUnfreeze     bool   `json:"canUnfreeze"`
	CanArchive      bool   `json:"canArchive"`
	CanRestore      bool   `json:"canRestore"`
}

// NamespaceWithMembership is a namespace row plus the current user's membership role.
type NamespaceWithMembership struct {
	ID          string
	Slug        string
	DisplayName string
	Type        string
	Description string
	Status      string
	CreatedBy   string
	CreatedAt   time.Time
	Role        string
}

func managedViewFor(row NamespaceWithMembership) ManagedNamespaceView {
	canManage := row.Role == "OWNER" || row.Role == "ADMIN"
	immutable := row.Type == "GLOBAL"
	status := row.Status

	canFreeze := false
	canUnfreeze := false
	canArchive := false
	canRestore := false
	if !immutable && canManage {
		switch status {
		case "ACTIVE":
			canFreeze = true
			canArchive = true
		case "FROZEN":
			canUnfreeze = true
			canArchive = true
		case "ARCHIVED":
			canRestore = true
		}
	}

	return ManagedNamespaceView{
		ID:              row.ID,
		Slug:            row.Slug,
		DisplayName:     row.DisplayName,
		Type:            row.Type,
		Description:     row.Description,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.UTC().Format(time.RFC3339),
		CreatedBy:       row.CreatedBy,
		CurrentUserRole: row.Role,
		Immutable:       immutable,
		CanFreeze:       canFreeze,
		CanUnfreeze:     canUnfreeze,
		CanArchive:      canArchive,
		CanRestore:      canRestore,
	}
}
