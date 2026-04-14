package namespace

import (
	"context"
	"errors"
	"strings"

	"skillhub/backend/internal/auth"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateNamespace(ctx context.Context, actor auth.Actor, input CreateNamespaceInput) (Namespace, error) {
	namespace := Namespace{
		ID:          "33333333-3333-3333-3333-" + strings.Repeat("3", 12),
		Slug:        strings.ToLower(strings.TrimSpace(input.Slug)),
		DisplayName: strings.TrimSpace(input.DisplayName),
		Type:        strings.TrimSpace(input.Type),
		Description: strings.TrimSpace(input.Description),
		Status:      "ACTIVE",
		CreatedBy:   actor.UserID,
	}
	if namespace.Slug == "" || namespace.DisplayName == "" || namespace.Type == "" {
		return Namespace{}, errors.New("slug, displayName and type are required")
	}

	if err := s.repo.Create(ctx, namespace); err != nil {
		return Namespace{}, err
	}
	if err := s.repo.AddMember(ctx, namespace.ID, actor.UserID, "OWNER"); err != nil {
		return Namespace{}, err
	}
	return namespace, nil
}

func (s *Service) AddMember(ctx context.Context, actor auth.Actor, input AddMemberInput) error {
	namespace, err := s.repo.GetBySlug(ctx, input.NamespaceSlug)
	if err != nil {
		return err
	}
	role, err := s.repo.GetMemberRole(ctx, namespace.ID, actor.UserID)
	if err != nil {
		return err
	}
	if role != "OWNER" && role != "ADMIN" {
		return errors.New("forbidden")
	}

	userID, err := s.repo.LookupUserIDByUsername(ctx, input.Username)
	if err != nil {
		return err
	}
	return s.repo.AddMember(ctx, namespace.ID, userID, input.Role)
}

func (s *Service) GetNamespace(ctx context.Context, slug string) (Namespace, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *Service) ListMine(ctx context.Context, actor auth.Actor) ([]ManagedNamespaceView, error) {
	rows, err := s.repo.ListNamespacesForUser(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]ManagedNamespaceView, 0, len(rows))
	for _, row := range rows {
		out = append(out, managedViewFor(row))
	}
	return out, nil
}
