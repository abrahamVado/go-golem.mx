package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/abrahamVado/go-golem.mx/internal/platform/config"
	"github.com/abrahamVado/go-golem.mx/internal/security"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// AUTH SERVICE
// -----------------------------------------------------------------------------
//
// The Auth Service owns authentication business rules.
//
// Responsibilities:
//
//   - login validation
//   - password verification
//   - access token creation
//   - refresh token creation
//   - refresh token rotation
//   - logout / token revocation
//   - password recovery orchestration
//
// This layer should NOT:
//
//   - parse HTTP requests
//   - write HTTP responses
//   - directly expose database details
//
// Handler  -> HTTP input/output
// Service  -> business rules
// Repository -> database persistence
//
// -----------------------------------------------------------------------------

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRegisterNotImplemented = errors.New("register scaffold: create company, owner, owner role inside transaction")
)

type Service struct {
	repo *Repository
	cfg  config.Config
}

// NewService creates the authentication service.
//
// Config is injected because token TTLs, secrets, bcrypt cost, and cookie
// behavior are runtime settings.
func NewService(r *Repository, cfg config.Config) *Service {
	return &Service{
		repo: r,
		cfg:  cfg,
	}
}

// Login authenticates a user and creates a new token pair.
//
// Flow:
//
//   1. Normalize email
//   2. Find user by email
//   3. Verify password hash
//   4. Issue short-lived access JWT
//   5. Generate opaque refresh token
//   6. Store only refresh token hash
//   7. Return access token + raw refresh token
//
// Security notes:
//
//   - Never reveal whether email or password was wrong.
//   - Never store raw refresh tokens.
//   - Always audit failed and successful login attempts later.
//   - Consider account lockout after repeated failures.
//
func (s *Service) Login(
	email string,
	password string,
	ip string,
	ua string,
) (AuthResponse, string, error) {
	email = normalizeEmail(email)

	if email == "" || password == "" {
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	if !security.CheckPassword(user.PasswordHash, password) {
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	accessToken, err := security.NewAccessToken(
		s.cfg.JWTAccessSecret,
		s.cfg.JWTAccessTTL,
		user.ID,
		user.CompanyID,
		user.BranchID,
	)
	if err != nil {
		return AuthResponse{}, "", err
	}

	refreshToken, err := security.NewOpaqueToken(32)
	if err != nil {
		return AuthResponse{}, "", err
	}

	refreshHash := security.HashToken(refreshToken)

	now := time.Now().UTC()

	err = s.repo.SaveRefreshToken(&RefreshToken{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		TokenHash: refreshHash,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
	})
	if err != nil {
		return AuthResponse{}, "", err
	}

	return AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.JWTAccessTTL.Seconds()),
	}, refreshToken, nil
}

// Refresh rotates a refresh token and returns a new token pair.
//
// Flow:
//
//   1. Hash incoming refresh token
//   2. Find stored token by hash
//   3. Validate token is active
//   4. Issue new access token
//   5. Generate new refresh token
//   6. Store new refresh token hash
//   7. Revoke old refresh token and link replacement
//
// Security notes:
//
//   - Refresh tokens should be single-use.
//   - Old token must be revoked after successful rotation.
//   - Reuse of a revoked refresh token should be treated as suspicious.
//   - Raw refresh tokens must never be stored.
//
func (s *Service) Refresh(
	oldToken string,
	ip string,
	ua string,
) (AuthResponse, string, error) {
	if oldToken == "" {
		return AuthResponse{}, "", ErrInvalidRefreshToken
	}

	oldHash := security.HashToken(oldToken)

	refreshRecord, err := s.repo.FindRefreshToken(oldHash)
	if err != nil {
		return AuthResponse{}, "", ErrInvalidRefreshToken
	}

	now := time.Now().UTC()

	// These fields depend on your RefreshToken model.
	// Keep this validation if your model has RevokedAt.
	if !refreshRecord.RevokedAt.IsZero() {
		return AuthResponse{}, "", ErrInvalidRefreshToken
	}

	if refreshRecord.ExpiresAt.Before(now) {
		return AuthResponse{}, "", ErrInvalidRefreshToken
	}

	accessToken, err := security.NewAccessToken(
		s.cfg.JWTAccessSecret,
		s.cfg.JWTAccessTTL,
		refreshRecord.UserID,
		refreshRecord.CompanyID,
		nil,
	)
	if err != nil {
		return AuthResponse{}, "", err
	}

	newRefreshToken, err := security.NewOpaqueToken(32)
	if err != nil {
		return AuthResponse{}, "", err
	}

	newRefreshHash := security.HashToken(newRefreshToken)

	if err := s.repo.SaveRefreshToken(&RefreshToken{
		UserID:    refreshRecord.UserID,
		CompanyID: refreshRecord.CompanyID,
		TokenHash: newRefreshHash,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
	}); err != nil {
		return AuthResponse{}, "", err
	}

	if err := s.repo.RevokeRefreshToken(oldHash, &newRefreshHash); err != nil {
		return AuthResponse{}, "", err
	}

	return AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.JWTAccessTTL.Seconds()),
	}, newRefreshToken, nil
}

// Logout revokes the current refresh token.
//
// Access tokens are short-lived and normally not revoked directly unless
// you implement an access-token denylist.
//
// Recommended behavior:
//
//   - revoke refresh token
//   - clear refresh cookie in handler
//   - audit logout event
//
func (s *Service) Logout(refresh string) error {
	if refresh == "" {
		return nil
	}

	hash := security.HashToken(refresh)

	return s.repo.RevokeRefreshToken(hash, nil)
}

// Register will create the initial SaaS tenant flow.
//
// Expected future transaction:
//
//   1. Create company
//   2. Create owner user
//   3. Hash owner password
//   4. Create or attach Owner role
//   5. Assign owner role to user
//   6. Seed default settings
//   7. Commit transaction
//
// This must be atomic.
// If one step fails, the whole registration must roll back.
func (s *Service) Register(req RegisterRequest) (uuid.UUID, error) {
	return uuid.Nil, ErrRegisterNotImplemented
}

// Recover starts password recovery.
//
// Security rule:
//
// Always return nil, even if the email does not exist.
// This prevents account enumeration.
func (s *Service) Recover(email string) error {
	_ = normalizeEmail(email)

	return nil
}

// Reset completes password reset.
//
// Expected future behavior:
//
//   - hash incoming reset token
//   - find valid unexpired reset token
//   - validate password strength
//   - hash new password
//   - update user password
//   - revoke reset token
//   - revoke existing refresh sessions
//
func (s *Service) Reset(token string, password string) error {
	return nil
}

// normalizeEmail prepares email values for lookup.
//
// Important:
//
// Database uniqueness should also enforce normalized email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}