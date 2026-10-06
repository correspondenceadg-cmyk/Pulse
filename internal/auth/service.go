package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/config"
)

var (
	ErrBadCredentials = errors.New("invalid credentials")
	ErrTokenReuse     = errors.New("refresh token reuse detected")
)

type Service struct {
	repo   *Repository
	issuer *TokenIssuer
	cfg    *config.Config
}

type TokenPair struct {
	Access     string
	Refresh    string
	RefreshExp time.Time
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{
		repo:   repo,
		issuer: NewTokenIssuer(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		cfg:    cfg,
	}
}

func (s *Service) Register(ctx context.Context, email, password, display string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("invalid email")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateUser(ctx, email, hash, display)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, *TokenPair, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil, ErrBadCredentials
		}
		return nil, nil, err
	}
	if !VerifyPassword(u.PasswordHash, password) {
		return nil, nil, ErrBadCredentials
	}

	pair, err := s.issuePair(ctx, u, nil)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (*User, *TokenPair, error) {
	hash := HashRefreshToken(rawRefresh)
	rt, err := s.repo.FindRefresh(ctx, hash)
	if err != nil {
		if errors.Is(err, ErrRefreshNotFound) {
			return nil, nil, ErrBadCredentials
		}
		return nil, nil, err
	}

	if rt.RevokedAt != nil {
		// Reuse detected: nuke the whole chain for this user
		_ = s.repo.RevokeUserChain(ctx, rt.UserID)
		return nil, nil, ErrTokenReuse
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, nil, ErrBadCredentials
	}

	u, err := s.repo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, nil, err
	}

	// Rotate: revoke old, issue new pair linked by rotated_from
	if err := s.repo.RevokeRefresh(ctx, hash); err != nil {
		return nil, nil, err
	}

	prevID := rt.ID
	pair, err := s.issuePair(ctx, u, &prevID)
	if err != nil {
		return nil, nil, err
	}
	return u, pair, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	return s.repo.RevokeRefresh(ctx, HashRefreshToken(rawRefresh))
}

func (s *Service) ParseAccess(raw string) (*Claims, error) {
	return s.issuer.ParseAccess(raw)
}

func (s *Service) issuePair(ctx context.Context, u *User, rotatedFrom *string) (*TokenPair, error) {
	access, err := s.issuer.IssueAccess(u.ID, u.Role)
	if err != nil {
		return nil, err
	}
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(s.issuer.RefreshTTL())
	if err := s.repo.StoreRefresh(ctx, u.ID, hash, exp, rotatedFrom); err != nil {
		return nil, err
	}
	return &TokenPair{Access: access, Refresh: raw, RefreshExp: exp}, nil
}