package auth

import (
	"errors"
	"github.com/golem-mx/core-api/internal/config"
	"github.com/golem-mx/core-api/internal/security"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	repo *Repository
	cfg  config.Config
}

func NewService(r *Repository, cfg config.Config) *Service { return &Service{repo: r, cfg: cfg} }

func (s *Service) Login(email, password, ip, ua string) (AuthResponse, string, error) {
	u, err := s.repo.FindUserByEmail(email)
	if err != nil || !security.CheckPassword(u.PasswordHash, password) {
		return AuthResponse{}, "", errors.New("invalid credentials")
	}
	access, err := security.NewAccessToken(s.cfg.JWTAccessSecret, s.cfg.JWTAccessTTL, u.ID, u.CompanyID, u.BranchID)
	if err != nil {
		return AuthResponse{}, "", err
	}
	refresh, err := security.NewOpaqueToken()
	if err != nil {
		return AuthResponse{}, "", err
	}
	hash := security.HashToken(refresh)
	err = s.repo.SaveRefreshToken(&RefreshToken{UserID: u.ID, CompanyID: u.CompanyID, TokenHash: hash, IPAddress: ip, UserAgent: ua, ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL)})
	if err != nil {
		return AuthResponse{}, "", err
	}
	return AuthResponse{AccessToken: access, TokenType: "Bearer", ExpiresIn: int64(s.cfg.JWTAccessTTL.Seconds())}, refresh, nil
}

func (s *Service) Refresh(oldToken, ip, ua string) (AuthResponse, string, error) {
	oldHash := security.HashToken(oldToken)
	rt, err := s.repo.FindRefreshToken(oldHash)
	if err != nil {
		return AuthResponse{}, "", err
	}
	access, err := security.NewAccessToken(s.cfg.JWTAccessSecret, s.cfg.JWTAccessTTL, rt.UserID, rt.CompanyID, nil)
	if err != nil {
		return AuthResponse{}, "", err
	}
	newToken, err := security.NewOpaqueToken()
	if err != nil {
		return AuthResponse{}, "", err
	}
	newHash := security.HashToken(newToken)
	if err := s.repo.SaveRefreshToken(&RefreshToken{UserID: rt.UserID, CompanyID: rt.CompanyID, TokenHash: newHash, IPAddress: ip, UserAgent: ua, ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL)}); err != nil {
		return AuthResponse{}, "", err
	}
	_ = s.repo.RevokeRefreshToken(oldHash, &newHash)
	return AuthResponse{AccessToken: access, TokenType: "Bearer", ExpiresIn: int64(s.cfg.JWTAccessTTL.Seconds())}, newToken, nil
}

func (s *Service) Logout(refresh string) error {
	h := security.HashToken(refresh)
	return s.repo.RevokeRefreshToken(h, nil)
}
func (s *Service) Register(req RegisterRequest) (uuid.UUID, error) {
	return uuid.Nil, errors.New("register scaffold: create company, owner, owner role inside transaction")
}
func (s *Service) Recover(email string) error         { return nil }
func (s *Service) Reset(token, password string) error { return nil }
