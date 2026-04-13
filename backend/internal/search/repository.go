package search

import "context"

type Repository interface {
	Search(ctx context.Context, query Query) ([]Result, error)
}
