package namespace

type Namespace struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedBy   string `json:"createdBy"`
}

type Member struct {
	UserID   string `json:"userId"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role"`
}

type CreateNamespaceInput struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type AddMemberInput struct {
	NamespaceSlug string
	Username      string `json:"username"`
	Role          string `json:"role"`
}
