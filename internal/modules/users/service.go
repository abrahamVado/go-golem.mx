package users

import (
	"errors"
	"strings"
	"time"

	"github.com/abrahamVado/go-paladin.mx/internal/security"
	"github.com/google/uuid"
)

var (
	ErrUserEmailRequired    = errors.New("email is required")
	ErrUserNameRequired     = errors.New("name is required")
	ErrUserPasswordRequired = errors.New("password must be at least 8 characters")
	ErrUserEmailExists      = errors.New("a user with this email already exists")
	ErrUserRoleInvalid      = errors.New("one or more selected roles are invalid for this company")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAccountInvalid   = errors.New("invalid account type")
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(companyID uuid.UUID) ([]UserSummary, error) {
	return s.repo.List(companyID)
}

func (s *Service) Create(companyID uuid.UUID, req CreateRequest, bcryptCost int) (UserSummary, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	password := strings.TrimSpace(req.Password)
	status := strings.TrimSpace(req.Status)

	switch {
	case email == "":
		return UserSummary{}, ErrUserEmailRequired
	case name == "":
		return UserSummary{}, ErrUserNameRequired
	case len(password) < 8:
		return UserSummary{}, ErrUserPasswordRequired
	}

	if status == "" {
		status = "active"
	}

	exists, err := s.repo.EmailExists(email)
	if err != nil {
		return UserSummary{}, err
	}
	if exists {
		return UserSummary{}, ErrUserEmailExists
	}

	accountType, premiumExpiresAt, freeExpiresAt, blockedAt := accountLifecycle(req.AccountType, time.Now().UTC())

	roleIDs, err := parseUUIDStrings(req.RoleIDs)
	if err != nil {
		return UserSummary{}, ErrUserRoleInvalid
	}
	if len(roleIDs) > 0 {
		valid, err := s.repo.CountValidRoles(companyID, roleIDs)
		if err != nil {
			return UserSummary{}, err
		}
		if valid != len(roleIDs) {
			return UserSummary{}, ErrUserRoleInvalid
		}
	}

	passwordHash, err := security.HashPassword(password, bcryptCost)
	if err != nil {
		return UserSummary{}, err
	}

	return s.repo.Create(companyID, email, name, passwordHash, status, accountType, premiumExpiresAt, freeExpiresAt, blockedAt, roleIDs)
}

func (s *Service) Update(companyID, userID uuid.UUID, req UpdateRequest) (UserSummary, error) {
	exists, err := s.repo.UserExists(companyID, userID)
	if err != nil {
		return UserSummary{}, err
	}
	if !exists {
		return UserSummary{}, ErrUserNotFound
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	status := strings.TrimSpace(req.Status)

	switch {
	case email == "":
		return UserSummary{}, ErrUserEmailRequired
	case name == "":
		return UserSummary{}, ErrUserNameRequired
	}

	if status == "" {
		status = "active"
	}

	emailTaken, err := s.repo.EmailExistsForOtherUser(companyID, userID, email)
	if err != nil {
		return UserSummary{}, err
	}
	if emailTaken {
		return UserSummary{}, ErrUserEmailExists
	}

	accountType, premiumExpiresAt, freeExpiresAt, blockedAt := accountLifecycle(req.AccountType, time.Now().UTC())

	roleIDs, err := parseUUIDStrings(req.RoleIDs)
	if err != nil {
		return UserSummary{}, ErrUserRoleInvalid
	}
	if len(roleIDs) > 0 {
		valid, err := s.repo.CountValidRoles(companyID, roleIDs)
		if err != nil {
			return UserSummary{}, err
		}
		if valid != len(roleIDs) {
			return UserSummary{}, ErrUserRoleInvalid
		}
	}

	return s.repo.Update(companyID, userID, email, name, status, accountType, premiumExpiresAt, freeExpiresAt, blockedAt, roleIDs)
}

func (s *Service) Delete(companyID, userID uuid.UUID) error {
	exists, err := s.repo.UserExists(companyID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	return s.repo.Delete(companyID, userID)
}

func parseUUIDStrings(values []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func accountLifecycle(raw string, now time.Time) (string, *time.Time, *time.Time, *time.Time) {
	accountType := NormalizeAccountType(raw)
	now = now.UTC()

	switch accountType {
	case AccountTypeOwner, AccountTypeFounder:
		return accountType, nil, nil, nil
	case AccountTypePremiumClient:
		premiumUntil := now.AddDate(0, 0, 30)
		freeUntil := premiumUntil.AddDate(0, 0, 30)
		return accountType, &premiumUntil, &freeUntil, nil
	case AccountTypeInvalidClient:
		blockedAt := now
		return accountType, nil, nil, &blockedAt
	default:
		freeUntil := now.AddDate(0, 0, 30)
		return AccountTypeFreeClient, nil, &freeUntil, nil
	}
}
