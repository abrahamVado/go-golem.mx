package branches

import (
	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) List(c *gin.Context) {
	response.OK(c, gin.H{"company_id": tenancy.CompanyID(c), "module": "branches", "items": []any{}})
}
func (h *Handler) Get(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "branches"})
}
func (h *Handler) Create(c *gin.Context) {
	response.Created(c, gin.H{"module": "branches", "status": "created"})
}
func (h *Handler) Update(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "branches", "status": "updated"})
}
func (h *Handler) Delete(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "branches", "status": "deleted"})
}
