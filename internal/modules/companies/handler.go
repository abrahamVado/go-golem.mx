package companies

import (
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/abrahamVado/go-golem.mx/internal/tenancy"
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
func (h *Handler) UpdateCurrent(c *gin.Context) { response.OK(c, gin.H{"status": "updated"}) }
