package tenancy

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CompanyIDKey = "company_id"
const UserIDKey = "user_id"
const BranchIDKey = "branch_id"

func Set(c *gin.Context, userID, companyID uuid.UUID, branchID *uuid.UUID) {
	c.Set(UserIDKey, userID)
	c.Set(CompanyIDKey, companyID)
	if branchID != nil {
		c.Set(BranchIDKey, *branchID)
	}
}
func CompanyID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(CompanyIDKey)
	id, _ := v.(uuid.UUID)
	return id
}
func UserID(c *gin.Context) uuid.UUID { v, _ := c.Get(UserIDKey); id, _ := v.(uuid.UUID); return id }
func BranchID(c *gin.Context) *uuid.UUID {
	v, ok := c.Get(BranchIDKey)
	if !ok {
		return nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
