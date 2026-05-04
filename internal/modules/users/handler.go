package users

import (
	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/abrahamVado/go-golem.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) List(c *gin.Context) {
	response.OK(c, gin.H{"company_id": tenancy.CompanyID(c), "module": "users", "items": []any{}})
}
func (h *Handler) Get(c *gin.Context) { response.OK(c, gin.H{"id": c.Param("id"), "module": "users"}) }
func (h *Handler) Create(c *gin.Context) {
	response.Created(c, gin.H{"module": "users", "status": "created"})
}
func (h *Handler) Update(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "users", "status": "updated"})
}
func (h *Handler) Delete(c *gin.Context) {
	response.OK(c, gin.H{"id": c.Param("id"), "module": "users", "status": "deleted"})
}
