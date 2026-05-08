package roles

import (
	"errors"
	"strings"

	permissionsmod "github.com/abrahamVado/go-paladin.mx/internal/modules/permissions"
	"github.com/google/uuid"
)

var (
	ErrRoleNameRequired      = errors.New("role name is required")
	ErrRoleNameExists        = errors.New("a role with this name already exists")
	ErrRolePermissionInvalid = errors.New("one or more selected permissions are invalid")
	ErrRoleNotFound          = errors.New("role not found")
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(companyID uuid.UUID) ([]RoleSummary, error) {
	return s.repo.List(companyID)
}

func (s *Service) Create(companyID uuid.UUID, req CreateRequest) (RoleSummary, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return RoleSummary{}, ErrRoleNameRequired
	}

	exists, err := s.repo.NameExists(companyID, name)
	if err != nil {
		return RoleSummary{}, err
	}
	if exists {
		return RoleSummary{}, ErrRoleNameExists
	}

	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, raw := range req.PermissionIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return RoleSummary{}, ErrRolePermissionInvalid
		}
		permissionIDs = append(permissionIDs, id)
	}

	if len(permissionIDs) > 0 {
		valid, err := s.repo.CountValidPermissions(permissionIDs)
		if err != nil {
			return RoleSummary{}, err
		}
		if valid != len(permissionIDs) {
			return RoleSummary{}, ErrRolePermissionInvalid
		}
	}

	return s.repo.Create(companyID, name, strings.TrimSpace(req.Description), permissionIDs)
}

func (s *Service) Update(companyID, roleID uuid.UUID, req UpdateRequest) (RoleSummary, error) {
	exists, err := s.repo.RoleExists(companyID, roleID)
	if err != nil {
		return RoleSummary{}, err
	}
	if !exists {
		return RoleSummary{}, ErrRoleNotFound
	}

	var nextName *string
	if req.Name != "" {
		trimmed := strings.TrimSpace(req.Name)
		if trimmed == "" {
			return RoleSummary{}, ErrRoleNameRequired
		}

		existing, err := s.repo.NameExists(companyID, trimmed)
		if err != nil {
			return RoleSummary{}, err
		}
		if existing {
			return RoleSummary{}, ErrRoleNameExists
		}
		nextName = &trimmed
	}

	var nextDescription *string
	if req.Description != "" {
		trimmedDescription := strings.TrimSpace(req.Description)
		nextDescription = &trimmedDescription
	}

	replacePermissions := req.PermissionIDs != nil
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, raw := range req.PermissionIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return RoleSummary{}, ErrRolePermissionInvalid
		}
		permissionIDs = append(permissionIDs, id)
	}

	if replacePermissions && len(permissionIDs) > 0 {
		valid, err := s.repo.CountValidPermissions(permissionIDs)
		if err != nil {
			return RoleSummary{}, err
		}
		if valid != len(permissionIDs) {
			return RoleSummary{}, ErrRolePermissionInvalid
		}
	}

	return s.repo.Update(companyID, roleID, nextName, nextDescription, permissionIDs, replacePermissions)
}

func (s *Service) ListPermissions(companyID, roleID uuid.UUID) ([]permissionsmod.Permission, error) {
	exists, err := s.repo.RoleExists(companyID, roleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrRoleNotFound
	}

	return s.repo.ListPermissions(companyID, roleID)
}

func (s *Service) Delete(companyID, roleID uuid.UUID) error {
	exists, err := s.repo.RoleExists(companyID, roleID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRoleNotFound
	}

	return s.repo.Delete(companyID, roleID)
}
