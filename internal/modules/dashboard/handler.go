package dashboard

import (
	"strconv"

	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) Index(c *gin.Context) {
	summary, err := h.svc.Summary(tenancy.CompanyID(c))
	if err != nil {
		response.Internal(c, "Failed to load dashboard summary")
		return
	}
	response.OK(c, summary)
}

func (h *Handler) SystemLogs(c *gin.Context) {
	limit := 12
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			response.BadRequest(c, "Invalid limit")
			return
		}
		limit = parsed
	}

	logs, err := h.svc.SystemLogs(tenancy.CompanyID(c), limit)
	if err != nil {
		response.Internal(c, "Failed to load system logs")
		return
	}
	response.OK(c, logs)
}
