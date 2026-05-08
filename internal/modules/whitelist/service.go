package whitelist

import (
	"errors"
	"strings"
)

var (
	ErrNameRequired    = errors.New("name is required")
	ErrEmailRequired   = errors.New("email is required")
	ErrMessageRequired = errors.New("message is required")
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(req CreateRequest) (*Request, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	message := strings.TrimSpace(req.Message)

	switch {
	case name == "":
		return nil, ErrNameRequired
	case email == "":
		return nil, ErrEmailRequired
	case message == "":
		return nil, ErrMessageRequired
	}

	item := &Request{
		Name:    name,
		Email:   email,
		Company: strings.TrimSpace(req.Company),
		Subject: strings.TrimSpace(req.Subject),
		Message: message,
		Source:  "landing_page",
		Status:  "pending",
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}
