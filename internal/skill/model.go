package skill

type Skill struct {
	ID          string `json:"id"`
	NamespaceID string `json:"namespaceId"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Summary     string `json:"summary"`
	OwnerID     string `json:"ownerId"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
}

type Version struct {
	ID            string `json:"id"`
	SkillID       string `json:"skillId"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	ManifestJSON  string `json:"manifestJson"`
	MetadataJSON  string `json:"metadataJson"`
	StoragePath   string `json:"storagePath"`
	NamespaceSlug string `json:"namespaceSlug,omitempty"`
	SkillSlug     string `json:"skillSlug,omitempty"`
}

type PublishInput struct {
	NamespaceSlug string
	SkillSlug     string
	Archive       []byte
}

type manifest struct {
	Version     string `json:"version"`
	DisplayName string `json:"displayName"`
	Summary     string `json:"summary"`
}

type CreateVersionParams struct {
	NamespaceSlug string
	SkillSlug     string
	Version       string
	DisplayName   string
	Summary       string
	Status        string
	ManifestJSON  string
	MetadataJSON  string
	StoragePath   string
	SubmittedBy   string
}
