package admin

import (
	"context"
	"errors"

	"skillhub/backend/internal/auth"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Approve(ctx context.Context, actor auth.Actor, taskID string, comment string) error {
	if !actor.HasRole("SUPER_ADMIN") && !actor.HasRole("SKILL_ADMIN") {
		return errors.New("forbidden")
	}
	_, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	return s.repo.Decide(ctx, taskID, actor.UserID, comment, "APPROVE", "PUBLISHED")
}
