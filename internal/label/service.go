package label

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"skillhub/backend/internal/auth"
)

type Repository interface {
	List(ctx context.Context) ([]Definition, error)
	Get(ctx context.Context, slug string) (Definition, error)
	Create(ctx context.Context, in CreateInput) error
	Update(ctx context.Context, slug string, in UpdateInput) error
	Delete(ctx context.Context, slug string) error
	UpdateSortOrder(ctx context.Context, items []SortOrderItem) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) requireSuperAdmin(actor auth.Actor) bool {
	return actor.HasRole("SUPER_ADMIN")
}

func (s *Service) List(ctx context.Context, actor auth.Actor) ([]Definition, error) {
	if !s.requireSuperAdmin(actor) {
		return nil, ErrForbidden
	}
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, actor auth.Actor, in CreateInput) (Definition, error) {
	if !s.requireSuperAdmin(actor) {
		return Definition{}, ErrForbidden
	}
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	if in.Slug == "" {
		return Definition{}, ErrInvalid
	}
	if err := s.repo.Create(ctx, in); err != nil {
		if strings.Contains(err.Error(), "slug exists") {
			return Definition{}, ErrConflict
		}
		return Definition{}, err
	}
	return s.repo.Get(ctx, in.Slug)
}

func (s *Service) Update(ctx context.Context, actor auth.Actor, slug string, in UpdateInput) (Definition, error) {
	if !s.requireSuperAdmin(actor) {
		return Definition{}, ErrForbidden
	}
	if err := s.repo.Update(ctx, slug, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Definition{}, ErrNotFound
		}
		return Definition{}, err
	}
	return s.repo.Get(ctx, slug)
}

func (s *Service) Delete(ctx context.Context, actor auth.Actor, slug string) error {
	if !s.requireSuperAdmin(actor) {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) UpdateSortOrder(ctx context.Context, actor auth.Actor, items []SortOrderItem) ([]Definition, error) {
	if !s.requireSuperAdmin(actor) {
		return nil, ErrForbidden
	}
	if err := s.repo.UpdateSortOrder(ctx, items); err != nil {
		return nil, err
	}
	return s.repo.List(ctx)
}

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrInvalid   = errors.New("invalid")
)
