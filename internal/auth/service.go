package auth

import (
	"context"
	"errors"
	"strings"

	"skillhub/backend/internal/config"
	"skillhub/backend/internal/platform"
)

// ErrChangePasswordInvalidCurrent is returned when the supplied current password does not match.
var ErrChangePasswordInvalidCurrent = errors.New("invalid current password")

// ErrChangePasswordWeak is returned when the new password does not meet minimum requirements.
var ErrChangePasswordWeak = errors.New("new password too short")

// ErrChangePasswordInvalidInput is returned when required password fields are missing or blank.
var ErrChangePasswordInvalidInput = errors.New("password fields required")

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

const minPasswordLength = 8

// ChangePassword updates the password for the authenticated user after verifying the current password.
func (s *Service) ChangePassword(ctx context.Context, actor Actor, currentPassword, newPassword string) error {
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return ErrChangePasswordInvalidInput
	}
	if len(newPassword) < minPasswordLength {
		return ErrChangePasswordWeak
	}

	record, err := s.repo.FindByID(ctx, actor.UserID)
	if err != nil {
		return err
	}
	if err := platform.CheckPassword(record.PasswordHash, currentPassword); err != nil {
		return ErrChangePasswordInvalidCurrent
	}

	hash, err := platform.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePasswordHash(ctx, actor.UserID, hash)
}
