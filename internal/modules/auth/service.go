package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	auditmod "github.com/abrahamVado/go-paladin.mx/internal/modules/audit"
	companiesmod "github.com/abrahamVado/go-paladin.mx/internal/modules/companies"
	rolesmod "github.com/abrahamVado/go-paladin.mx/internal/modules/roles"
	"github.com/abrahamVado/go-paladin.mx/internal/modules/users"
	"github.com/abrahamVado/go-paladin.mx/internal/platform/config"
	"github.com/abrahamVado/go-paladin.mx/internal/platform/mail"
	"github.com/abrahamVado/go-paladin.mx/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrRegisterNotImplemented   = errors.New("register scaffold: create company, owner, owner role inside transaction")
	ErrProfileNameRequired      = errors.New("name is required")
	ErrPasswordMismatch         = errors.New("current password is incorrect")
	ErrInvalidAvatar            = errors.New("avatar must be an image smaller than 2 MB")
	ErrCompanyNameRequired      = errors.New("company name is required")
	ErrRegisterNameRequired     = errors.New("name is required")
	ErrRegisterEmailExists      = errors.New("a user with this email already exists")
	ErrAccountBlocked           = errors.New("your premium and grace periods have ended; renew your plan to continue")
	ErrResetTokenInvalid        = errors.New("invalid or expired reset token")
	ErrEmailNotVerified         = errors.New("email verification required before login")
	ErrEmailVerificationInvalid = errors.New("invalid or expired verification token")
)

type Service struct {
	repo   *Repository
	cfg    config.Config
	mailer mail.Sender
	audit  *auditmod.Repository
}

// NewService creates the authentication service.
//
// Config is injected because token TTLs, secrets, bcrypt cost, and cookie
// behavior are runtime settings.
func NewService(r *Repository, cfg config.Config, mailer mail.Sender, auditRepo *auditmod.Repository) *Service {
	return &Service{
		repo:   r,
		cfg:    cfg,
		mailer: mailer,
		audit:  auditRepo,
	}
}

// Login authenticates a user and creates a new token pair.
//
// Flow:
//
//  1. Normalize email
//  2. Find user by email
//  3. Verify password hash
//  4. Issue short-lived access JWT
//  5. Generate opaque refresh token
//  6. Store only refresh token hash
//  7. Return access token + raw refresh token
//
// Security notes:
//
//   - Never reveal whether email or password was wrong.
//   - Never store raw refresh tokens.
//   - Always audit failed and successful login attempts later.
//   - Consider account lockout after repeated failures.
func (s *Service) Login(
	email string,
	password string,
	companySlug string,
	ip string,
	ua string,
) (AuthResponse, string, error) {
	email = normalizeEmail(email)
	companySlug = strings.TrimSpace(strings.ToLower(companySlug))

	if email == "" || password == "" {
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	user, err := s.repo.FindUserAnyStatusByEmailAndCompanySlug(email, companySlug)
	if err != nil {
		s.auditAuthFailure(nil, "login_failed", ip, ua, map[string]any{"email": email, "company_slug": companySlug})
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		s.auditAuthFailure(&user.CompanyID, "login_unverified", ip, ua, map[string]any{"user_id": user.ID.String(), "email": email})
		return AuthResponse{}, "", ErrEmailNotVerified
	}

	if !security.CheckPassword(user.PasswordHash, password) {
		s.auditAuthFailure(&user.CompanyID, "login_failed", ip, ua, map[string]any{"user_id": user.ID.String(), "email": email})
		return AuthResponse{}, "", ErrInvalidCredentials
	}

	snapshot := users.ResolveAccountSnapshot(user, time.Now().UTC())
	if snapshot.Type != user.AccountType || (snapshot.IsBlocked && user.BlockedAt == nil) || (!snapshot.IsBlocked && user.BlockedAt != nil) {
		if err := s.repo.UpdateAccountState(user.CompanyID, user.ID, snapshot.Type, snapshot.BlockedAt); err != nil {
			return AuthResponse{}, "", err
		}
		user.AccountType = snapshot.Type
		user.BlockedAt = snapshot.BlockedAt
	}
	if snapshot.IsBlocked {
		s.auditAuthFailure(&user.CompanyID, "login_blocked", ip, ua, map[string]any{"user_id": user.ID.String(), "email": email})
		return AuthResponse{}, "", ErrAccountBlocked
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

	s.auditAuthSuccess(user.CompanyID, &user.ID, "login_succeeded", ip, ua, map[string]any{"email": user.Email, "company_slug": companySlug})

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
//  1. Hash incoming refresh token
//  2. Find stored token by hash
//  3. Validate token is active
//  4. Issue new access token
//  5. Generate new refresh token
//  6. Store new refresh token hash
//  7. Revoke old refresh token and link replacement
//
// Security notes:
//
//   - Refresh tokens should be single-use.
//   - Old token must be revoked after successful rotation.
//   - Reuse of a revoked refresh token should be treated as suspicious.
//   - Raw refresh tokens must never be stored.
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

	// RevokedAt is nullable in the model, so nil means the token is still active.
	if refreshRecord.RevokedAt != nil && !refreshRecord.RevokedAt.IsZero() {
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

	s.auditAuthSuccess(refreshRecord.CompanyID, &refreshRecord.UserID, "refresh_succeeded", ip, ua, nil)

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
//  1. Create company
//  2. Create owner user
//  3. Hash owner password
//  4. Create or attach Owner role
//  5. Assign owner role to user
//  6. Seed default settings
//  7. Commit transaction
//
// This must be atomic.
// If one step fails, the whole registration must roll back.
func (s *Service) Register(req RegisterRequest) (uuid.UUID, error) {
	companyName := strings.TrimSpace(req.CompanyName)
	name := strings.TrimSpace(req.Name)
	email := normalizeEmail(req.Email)
	password := strings.TrimSpace(req.Password)

	switch {
	case companyName == "":
		return uuid.Nil, ErrCompanyNameRequired
	case name == "":
		return uuid.Nil, ErrRegisterNameRequired
	case email == "":
		return uuid.Nil, ErrInvalidCredentials
	}

	if err := security.ValidatePassword(password); err != nil {
		return uuid.Nil, err
	}

	emailExists, err := s.repo.EmailExists(email)
	if err != nil {
		return uuid.Nil, err
	}
	if emailExists {
		return uuid.Nil, ErrRegisterEmailExists
	}

	slug, err := s.uniqueCompanySlug(companyName)
	if err != nil {
		return uuid.Nil, err
	}

	hash, err := security.HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return uuid.Nil, err
	}

	companyID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	err = s.repo.CreateCompanyWithClientUser(
		companiesmod.Company{
			ID:     companyID,
			Name:   companyName,
			Slug:   slug,
			Status: "active",
		},
		users.User{
			ID:           userID,
			CompanyID:    companyID,
			Email:        email,
			Name:         name,
			PasswordHash: hash,
			Status:       "active",
		},
		[]string{
			"project:view",
			"task:create",
			"task:view",
		},
		rolesmod.Role{
			ID:          roleID,
			Name:        "Client",
			Description: "Client-facing access limited to project ticket intake",
			IsSystem:    true,
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.sendVerificationEmail(companyID, userID, email, name); err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *Service) VerifyEmail(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrEmailVerificationInvalid
	}

	tokenHash := security.HashToken(token)
	record, err := s.repo.FindActiveEmailVerificationToken(tokenHash)
	if err != nil {
		return ErrEmailVerificationInvalid
	}

	if err := s.repo.VerifyEmailConsumeToken(record.ID, record.CompanyID, record.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmailVerificationInvalid
		}
		return err
	}

	s.auditAuthSuccess(record.CompanyID, &record.UserID, "email_verified", "", "", nil)
	return nil
}

func (s *Service) ResendVerification(email, companySlug string) error {
	email = normalizeEmail(email)
	companySlug = strings.TrimSpace(strings.ToLower(companySlug))
	if email == "" {
		return nil
	}

	user, err := s.repo.FindUserAnyStatusByEmailAndCompanySlug(email, companySlug)
	if err != nil {
		return nil
	}
	if user.EmailVerifiedAt != nil {
		return nil
	}

	return s.sendVerificationEmail(user.CompanyID, user.ID, user.Email, user.Name)
}

// Recover starts password recovery.
//
// Security rule:
//
// Always return nil, even if the email does not exist.
// This prevents account enumeration.
func (s *Service) Recover(email string) error {
	email = normalizeEmail(email)
	if email == "" {
		return nil
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		return nil
	}

	token, err := security.NewOpaqueToken(32)
	if err != nil {
		return err
	}
	tokenHash := security.HashToken(token)
	if err := s.repo.SavePasswordResetToken(&PasswordResetToken{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.PasswordResetTTL),
	}); err != nil {
		return err
	}
	s.auditAuthSuccess(user.CompanyID, &user.ID, "password_reset_requested", "", "", map[string]any{"email": user.Email})

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), token)
	companySlug, err := s.repo.FindCompanySlugByID(user.CompanyID)
	if err != nil {
		return err
	}
	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>We received a password reset request for <strong>%s</strong>.</p><p><a href=\"%s\">Reset your password</a></p><p>If you did not request this, you can ignore this email.</p>",
		user.Name,
		companySlug,
		resetURL,
	)
	textBody := fmt.Sprintf(
		"Hello %s,\n\nWe received a password reset request for %s.\n\nReset your password here: %s\n\nIf you did not request this, you can ignore this email.\n",
		user.Name,
		companySlug,
		resetURL,
	)

	return s.mailer.Send([]string{user.Email}, "Reset your Paladin password", htmlBody, textBody)
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
func (s *Service) Reset(token string, password string) error {
	token = strings.TrimSpace(token)
	password = strings.TrimSpace(password)
	if token == "" {
		return ErrResetTokenInvalid
	}
	if err := security.ValidatePassword(password); err != nil {
		return err
	}

	tokenHash := security.HashToken(token)
	resetToken, err := s.repo.FindActivePasswordResetToken(tokenHash)
	if err != nil {
		return ErrResetTokenInvalid
	}

	if _, err := s.repo.FindUserByID(resetToken.CompanyID, resetToken.UserID); err != nil {
		return ErrResetTokenInvalid
	}

	hash, err := security.HashPassword(password, s.cfg.BcryptCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordConsumeResetTokenAndRevokeSessions(resetToken.ID, resetToken.CompanyID, resetToken.UserID, hash); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResetTokenInvalid
		}
		return err
	}
	s.auditAuthSuccess(resetToken.CompanyID, &resetToken.UserID, "password_reset_completed", "", "", nil)
	return nil
}

type MeResponse struct {
	UserID               uuid.UUID  `json:"user_id"`
	CompanyID            uuid.UUID  `json:"company_id"`
	BranchID             *uuid.UUID `json:"branch_id,omitempty"`
	Name                 string     `json:"name"`
	Email                string     `json:"email"`
	EmailVerifiedAt      *time.Time `json:"email_verified_at,omitempty"`
	IsEmailVerified      bool       `json:"is_email_verified"`
	Role                 string     `json:"role,omitempty"`
	PermissionNames      []string   `json:"permission_names,omitempty"`
	AvatarURL            string     `json:"avatar_url,omitempty"`
	AccountType          string     `json:"account_type"`
	IsPremium            bool       `json:"is_premium"`
	IsBlocked            bool       `json:"is_blocked"`
	PremiumDaysRemaining int        `json:"premium_days_remaining"`
	FreeDaysRemaining    int        `json:"free_days_remaining"`
	PremiumExpiresAt     *time.Time `json:"premium_expires_at,omitempty"`
	FreeExpiresAt        *time.Time `json:"free_expires_at,omitempty"`
	BlockedAt            *time.Time `json:"blocked_at,omitempty"`
}

func (s *Service) Me(companyID, userID uuid.UUID) (MeResponse, error) {
	user, err := s.repo.FindUserByID(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	role, err := s.repo.FindPrimaryRoleName(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	permissionNames, err := s.repo.FindPermissionNames(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	return toMeResponse(user, role, permissionNames), nil
}

func (s *Service) UpdateMe(companyID, userID uuid.UUID, req MeUpdateRequest) (MeResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return MeResponse{}, ErrProfileNameRequired
	}

	user, err := s.repo.UpdateMyProfile(companyID, userID, name)
	if err != nil {
		return MeResponse{}, err
	}

	role, err := s.repo.FindPrimaryRoleName(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	permissionNames, err := s.repo.FindPermissionNames(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	return toMeResponse(user, role, permissionNames), nil
}

func (s *Service) ChangeMyPassword(companyID, userID uuid.UUID, req ChangeMyPasswordRequest) error {
	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)

	user, err := s.repo.FindUserByID(companyID, userID)
	if err != nil {
		return err
	}
	if !security.CheckPassword(user.PasswordHash, oldPassword) {
		return ErrPasswordMismatch
	}
	if err := security.ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := security.HashPassword(newPassword, s.cfg.BcryptCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePasswordAndRevokeSessions(companyID, userID, hash); err != nil {
		return err
	}
	s.auditAuthSuccess(companyID, &userID, "password_changed", "", "", nil)
	return nil
}

func (s *Service) auditAuthSuccess(companyID uuid.UUID, userID *uuid.UUID, action string, ip string, ua string, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Create(auditmod.AuditLog{
		CompanyID: companyID,
		UserID:    userID,
		Action:    action,
		Resource:  "auth",
		IPAddress: ip,
		UserAgent: ua,
		Metadata:  auditmod.JSONMetadata(metadata),
	})
}

func (s *Service) auditAuthFailure(companyID *uuid.UUID, action string, ip string, ua string, metadata map[string]any) {
	if s.audit == nil || companyID == nil {
		return
	}
	_ = s.audit.Create(auditmod.AuditLog{
		CompanyID: *companyID,
		Action:    action,
		Resource:  "auth",
		IPAddress: ip,
		UserAgent: ua,
		Metadata:  auditmod.JSONMetadata(metadata),
	})
}

func (s *Service) UpdateMyAvatar(companyID, userID uuid.UUID, header *multipart.FileHeader) (MeResponse, error) {
	dataURL, err := avatarDataURL(header)
	if err != nil {
		return MeResponse{}, err
	}

	user, err := s.repo.UpdateMyAvatar(companyID, userID, dataURL)
	if err != nil {
		return MeResponse{}, err
	}

	role, err := s.repo.FindPrimaryRoleName(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	permissionNames, err := s.repo.FindPermissionNames(companyID, userID)
	if err != nil {
		return MeResponse{}, err
	}

	return toMeResponse(user, role, permissionNames), nil
}

func toMeResponse(user users.User, role string, permissionNames []string) MeResponse {
	snapshot := users.ResolveAccountSnapshot(user, time.Now().UTC())
	return MeResponse{
		UserID:               user.ID,
		CompanyID:            user.CompanyID,
		BranchID:             user.BranchID,
		Name:                 user.Name,
		Email:                user.Email,
		EmailVerifiedAt:      user.EmailVerifiedAt,
		IsEmailVerified:      user.EmailVerifiedAt != nil,
		Role:                 role,
		PermissionNames:      permissionNames,
		AvatarURL:            user.AvatarURL,
		AccountType:          snapshot.Type,
		IsPremium:            snapshot.IsPremium,
		IsBlocked:            snapshot.IsBlocked,
		PremiumDaysRemaining: snapshot.PremiumDaysRemaining,
		FreeDaysRemaining:    snapshot.FreeDaysRemaining,
		PremiumExpiresAt:     user.PremiumExpiresAt,
		FreeExpiresAt:        user.FreeExpiresAt,
		BlockedAt:            snapshot.BlockedAt,
	}
}

func avatarDataURL(header *multipart.FileHeader) (string, error) {
	if header == nil || header.Size <= 0 || header.Size > 2<<20 {
		return "", ErrInvalidAvatar
	}

	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		return "", ErrInvalidAvatar
	}

	return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data)), nil
}

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func (s *Service) uniqueCompanySlug(companyName string) (string, error) {
	base := slugifyCompanyName(companyName)
	candidate := base
	for i := 1; i <= 1000; i++ {
		exists, err := s.repo.CompanySlugExists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+1)
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix()), nil
}

func slugifyCompanyName(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = slugSanitizer.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "company"
	}
	return value
}

// normalizeEmail prepares email values for lookup.
//
// Important:
//
// Database uniqueness should also enforce normalized email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) sendVerificationEmail(companyID, userID uuid.UUID, email, name string) error {
	if s.mailer == nil || !s.mailer.Enabled() {
		return nil
	}

	token, err := security.NewOpaqueToken(32)
	if err != nil {
		return err
	}
	tokenHash := security.HashToken(token)
	if err := s.repo.SaveEmailVerificationToken(&EmailVerificationToken{
		UserID:    userID,
		CompanyID: companyID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.EmailVerificationTTL),
	}); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), token)
	htmlBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>Please verify your email to activate your Paladin account.</p><p><a href=\"%s\">Verify email</a></p><p>If you did not create this account, you can ignore this email.</p>",
		name,
		verifyURL,
	)
	textBody := fmt.Sprintf(
		"Hello %s,\n\nPlease verify your email to activate your Paladin account.\n\nVerify here: %s\n\nIf you did not create this account, you can ignore this email.\n",
		name,
		verifyURL,
	)

	if err := s.mailer.Send([]string{email}, "Verify your Paladin email", htmlBody, textBody); err != nil {
		return err
	}

	s.auditAuthSuccess(companyID, &userID, "email_verification_sent", "", "", map[string]any{"email": email})
	return nil
}
