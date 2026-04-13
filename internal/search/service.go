package search

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, query Query) ([]Result, error) {
	query.Text = strings.TrimSpace(query.Text)
	if query.Text == "" {
		return nil, errors.New("search text is required")
	}
	return s.repo.Search(ctx, query)
}
