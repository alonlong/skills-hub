package search

type Query struct {
	Text          string
	NamespaceSlug string
}

type Result struct {
	NamespaceSlug string  `json:"namespaceSlug"`
	Slug          string  `json:"slug"`
	Title         string  `json:"title"`
	Summary       string  `json:"summary"`
	Rank          float64 `json:"rank"`
}
