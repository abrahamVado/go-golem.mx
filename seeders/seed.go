package seeders

import (
	"github.com/example/gin-multitenant-backend/internal/config"
	"github.com/example/gin-multitenant-backend/internal/security"
	"gorm.io/gorm"
	"log"
)

var Permissions = []string{"dashboard.read", "users.read", "users.create", "users.update", "users.delete", "users.invite", "users.disable", "roles.read", "roles.create", "roles.update", "roles.delete", "roles.assign", "permissions.read", "permissions.assign", "settings.read", "settings.update", "companies.read", "companies.update", "branches.read", "branches.create", "branches.update", "branches.delete", "audit.read", "security.sessions.read", "security.sessions.revoke"}
var Roles = []string{"Owner", "Admin", "Manager", "Employee", "Auditor"}

func Run(db *gorm.DB, cfg config.Config) error {
	log.Println("seeding default company, branch, roles, permissions, owner")
	hash, err := security.HashPassword(cfg.DefaultOwnerPassword, cfg.BcryptCost)
	if err != nil {
		return err
	}
	_ = hash
	// TODO: wrap in transaction and insert companies, branches, permissions, roles, role_permissions, owner user, user_roles.
	return nil
}
