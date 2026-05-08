package users

import (
	"errors"

	"github.com/abrahamVado/go-paladin.mx/internal/platform/config"
	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(tenancy.CompanyID(c))
	if err != nil {
		response.Internal(c, "Failed to load users")
		return
	}
	response.OK(c, gin.H{"company_id": tenancy.CompanyID(c), "module": "users", "items": items})
}
func (h *Handler) Get(c *gin.Context) { response.OK(c, gin.H{"id": c.Param("id"), "module": "users"}) }
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid user payload")
		return
	}

	cfg := config.Load()
	user, err := h.svc.Create(tenancy.CompanyID(c), req, cfg.BcryptCost)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserEmailRequired), errors.Is(err, ErrUserNameRequired), errors.Is(err, ErrUserPasswordRequired), errors.Is(err, ErrUserRoleInvalid):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrUserEmailExists):
			response.Conflict(c, err.Error())
		default:
			response.Internal(c, "Failed to create user")
		}
		return
	}

	response.Created(c, user)
}
func (h *Handler) Update(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user id")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid user payload")
		return
	}

	user, err := h.svc.Update(tenancy.CompanyID(c), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrUserEmailRequired), errors.Is(err, ErrUserNameRequired), errors.Is(err, ErrUserRoleInvalid):
			response.BadRequest(c, err.Error())
		case errors.Is(err, ErrUserEmailExists):
			response.Conflict(c, err.Error())
		default:
			response.Internal(c, "Failed to update user")
		}
		return
	}

	response.OK(c, user)
}
func (h *Handler) Delete(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user id")
		return
	}

	if err := h.svc.Delete(tenancy.CompanyID(c), userID); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(c, err.Error())
		default:
			response.Internal(c, "Failed to delete user")
		}
		return
	}

	response.OK(c, gin.H{"id": c.Param("id"), "module": "users", "status": "deleted"})
}
