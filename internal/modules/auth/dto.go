package auth

type RegisterRequest struct {
	CompanyName string `json:"company_name" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=10"`
}
type LoginRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	CompanySlug string `json:"company_slug"`
}
type CLIRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type RecoverRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}
type ResendVerificationRequest struct {
	Email       string `json:"email" binding:"required,email"`
	CompanySlug string `json:"company_slug"`
}
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=10"`
}
type MeUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

type ChangeMyPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type BrowserSessionResponse struct {
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
}

type CLISessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
