package apikeys

import (
	"github.com/abrahamVado/go-paladin.mx/internal/middleware"
	rbacmod "github.com/abrahamVado/go-paladin.mx/internal/modules/rbac"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(private *gin.RouterGroup, rbac *rbacmod.Service, h *Handler) {
	group := private.Group("/apikeys")
	group.Use(middleware.RequirePermission(rbac, "apikey:manage"))

	group.GET("/clients", h.ListClients)
	group.POST("/clients", h.CreateClient)
	group.POST("/clients/:clientId/keys", h.CreateKey)
	group.POST("/clients/:clientId/keys/:keyId/revoke", h.RevokeKey)
	group.POST("/clients/:clientId/public-keys", h.UploadPublicKey)
	group.POST("/clients/:clientId/public-keys/:publicKeyId/activate", h.ActivatePublicKey)
	group.POST("/clients/:clientId/public-keys/:publicKeyId/revoke", h.RevokePublicKey)
}
