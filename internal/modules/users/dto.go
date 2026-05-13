package users

type CreateRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Name        string   `json:"name" binding:"required"`
	Password    string   `json:"password" binding:"required,min=8"`
	Status      string   `json:"status"`
	AccountType string   `json:"account_type"`
	RoleIDs     []string `json:"role_ids"`
}

type UpdateRequest struct {
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	AccountType string   `json:"account_type"`
	RoleIDs     []string `json:"role_ids"`
}
