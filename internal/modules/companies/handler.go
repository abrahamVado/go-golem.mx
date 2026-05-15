package companies

import (
	"errors"

	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Current(c *gin.Context) {
	company, err := h.svc.repo.FindByID(tenancy.CompanyID(c))
	if err != nil {
		response.Fail(c, 404, "COMPANY_NOT_FOUND", "Company not found")
		return
	}
	response.OK(c, company)
}
func (h *Handler) UpdateCurrent(c *gin.Context) {
	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid company payload")
		return
	}

	company, err := h.svc.UpdateCurrent(tenancy.CompanyID(c), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrCompanyNameRequired):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to update company")
		}
		return
	}

	response.OK(c, company)
}
