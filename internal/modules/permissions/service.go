package permissions

import (
	"errors"
	"strings"
)

var (
	ErrPermissionNameRequired = errors.New("permission name is required")
	ErrPermissionNameExists   = errors.New("a permission with this name already exists")
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) List() ([]Permission, error) {
	return s.repo.List()
}

func (s *Service) Create(req CreateRequest) (Permission, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Permission{}, ErrPermissionNameRequired
	}

	exists, err := s.repo.NameExists(name)
	if err != nil {
		return Permission{}, err
	}
	if exists {
		return Permission{}, ErrPermissionNameExists
	}

	return s.repo.Create(name, strings.TrimSpace(req.Description))
}
