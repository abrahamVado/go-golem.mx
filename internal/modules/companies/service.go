package companies

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrCompanyNameRequired = errors.New("company name is required")

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{repo: r} }

func (s *Service) UpdateCurrent(companyID uuid.UUID, req UpdateCompanyRequest) (Company, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return Company{}, ErrCompanyNameRequired
	}

	return s.repo.UpdateCurrent(companyID, name)
}
