package auth

import (
	"context"
	"errors"

	"skillhub/backend/internal/config"
	"skillhub/backend/internal/platform"
)

type Service struct {
	repo   Repository
	issuer JWTIssuer
	cfg    config.Config
}

func NewService(repo Repository, issuer JWTIssuer, cfg config.Config) *Service {
	return &Service{
		repo:   repo,
		issuer: issuer,
		cfg:    cfg,
	}
}

func (s *Service) EnsureBootstrapAdmin(ctx context.Context) error {
	return s.repo.EnsureBootstrapAdmin(ctx, s.cfg)
}

func (s *Service) Login(ctx context.Context, username string, password string) (string, User, error) {
	record, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", User{}, err
	}
	if record.Status != "ACTIVE" {
		return "", User{}, errors.New("account disabled")
	}
	if err := platform.CheckPassword(record.PasswordHash, password); err != nil {
		return "", User{}, err
	}

	user := toUser(record)
	token, err := s.issuer.Issue(user)
	if err != nil {
		return "", User{}, err
	}
	return token, user, nil
}

func (s *Service) CurrentUser(ctx context.Context, actor Actor) (User, error) {
	record, err := s.repo.FindByID(ctx, actor.UserID)
	if err != nil {
		return User{}, err
	}
	return toUser(record), nil
}

func (s *Service) AuthenticateToken(token string) (Actor, error) {
	return s.issuer.Parse(token)
}
