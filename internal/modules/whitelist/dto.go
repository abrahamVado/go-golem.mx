package whitelist

type CreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Company string `json:"company"`
	Subject string `json:"subject"`
	Message string `json:"message" binding:"required"`
}
