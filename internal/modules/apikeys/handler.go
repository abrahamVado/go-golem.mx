package apikeys

import (
	"errors"

	"github.com/abrahamVado/go-paladin.mx/internal/response"
	"github.com/abrahamVado/go-paladin.mx/internal/tenancy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

func (h *Handler) ListClients(c *gin.Context) {
	items, err := h.svc.ListClients(tenancy.CompanyID(c))
	if err != nil {
		response.Internal(c, "Failed to load API clients")
		return
	}
	response.OK(c, items)
}

func (h *Handler) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid API client payload")
		return
	}

	client, err := h.svc.CreateClient(tenancy.CompanyID(c), tenancy.UserID(c), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrClientNameRequired):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to create API client")
		}
		return
	}

	response.Created(c, client)
}

func (h *Handler) CreateKey(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.BadRequest(c, "Invalid API client id")
		return
	}

	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid API key payload")
		return
	}

	key, err := h.svc.CreateKey(tenancy.CompanyID(c), clientID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrClientNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrScopesRequired):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to create API key")
		}
		return
	}

	response.Created(c, key)
}

func (h *Handler) RevokeKey(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.BadRequest(c, "Invalid API client id")
		return
	}

	if err := h.svc.RevokeKey(tenancy.CompanyID(c), clientID, c.Param("keyId")); err != nil {
		switch {
		case errors.Is(err, ErrClientNotFound):
			response.NotFound(c, err.Error())
		default:
			response.Internal(c, "Failed to revoke API key")
		}
		return
	}

	response.OK(c, gin.H{"status": "revoked"})
}

func (h *Handler) UploadPublicKey(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.BadRequest(c, "Invalid API client id")
		return
	}

	var req UploadPublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid public key payload")
		return
	}

	publicKey, err := h.svc.UploadPublicKey(tenancy.CompanyID(c), clientID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrClientNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrPublicKeyRequired), errors.Is(err, ErrPublicKeyInvalid):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to upload public key")
		}
		return
	}

	response.Created(c, publicKey)
}

func (h *Handler) ActivatePublicKey(c *gin.Context) {
	clientID, publicKeyID, ok := parseIDs(c)
	if !ok {
		return
	}

	var req ActivatePublicKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid activation payload")
		return
	}

	if err := h.svc.ActivatePublicKey(tenancy.CompanyID(c), clientID, publicKeyID, req); err != nil {
		switch {
		case errors.Is(err, ErrPublicKeyNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrChallengeRequired), errors.Is(err, ErrChallengeSignatureInvalid):
			response.BadRequest(c, err.Error())
		default:
			response.Internal(c, "Failed to activate public key")
		}
		return
	}

	response.OK(c, gin.H{"status": "active"})
}

func (h *Handler) RevokePublicKey(c *gin.Context) {
	clientID, publicKeyID, ok := parseIDs(c)
	if !ok {
		return
	}

	if err := h.svc.RevokePublicKey(tenancy.CompanyID(c), clientID, publicKeyID); err != nil {
		switch {
		case errors.Is(err, ErrPublicKeyNotFound):
			response.NotFound(c, err.Error())
		default:
			response.Internal(c, "Failed to revoke public key")
		}
		return
	}

	response.OK(c, gin.H{"status": "revoked"})
}

func parseIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	clientID, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		response.BadRequest(c, "Invalid API client id")
		return uuid.Nil, uuid.Nil, false
	}
	publicKeyID, err := uuid.Parse(c.Param("publicKeyId"))
	if err != nil {
		response.BadRequest(c, "Invalid public key id")
		return uuid.Nil, uuid.Nil, false
	}
	return clientID, publicKeyID, true
}
