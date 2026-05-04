package dashboard

import (
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/abrahamVado/go-golem.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Index(c *gin.Context) {
	response.OK(c, gin.H{"company_id": tenancy.CompanyID(c), "summary": "dashboard scaffold"})
}
