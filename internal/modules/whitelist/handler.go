package whitelist

import (
	"errors"

	"github.com/abrahamVado/go-golem.mx/internal/response"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid whitelist payload")
		return
	}

	item, err := h.svc.Create(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNameRequired), errors.Is(err, ErrEmailRequired), errors.Is(err, ErrMessageRequired):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to save whitelist request")
		}
		return
	}

	response.Created(c, item)
}
