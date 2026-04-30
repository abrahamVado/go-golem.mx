package roles

import (
	"github.com/example/gin-multitenant-backend/internal/response"
	"github.com/example/gin-multitenant-backend/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) List(c *gin.Context) {
	response.OK(c, gin.H{"company_id": tenancy.CompanyID(c), "module": "roles", "items": []any{}})
}
func (h *Handler) Get(c *gin.Context) { response.OK(c, gin.H{"id": c.Param("id"), "module": "roles"}) }
func (h *Handler) Create(c *gin.Context) {
	response.Created(c, gin.H{"module": "roles", "status": "created"})
}
func (h *Handler) Update(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "roles", "status": "updated"})
}
func (h *Handler) Delete(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "roles", "status": "deleted"})
}
