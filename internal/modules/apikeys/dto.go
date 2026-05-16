package apikeys

type CreateClientRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateKeyRequest struct {
	Scopes    []string `json:"scopes"`
	ExpiresAt *string  `json:"expires_at"`
}

type UploadPublicKeyRequest struct {
	OpenSSHPublicKey string `json:"openssh_public_key"`
}

type ActivatePublicKeyRequest struct {
	Challenge          string `json:"challenge"`
	ChallengeSignature string `json:"challenge_signature"`
}

type CreateKeyResponse struct {
	KeyID            string   `json:"key_id"`
	Scopes           []string `json:"scopes"`
	ExpiresAt        *string  `json:"expires_at,omitempty"`
	APIKeySecretOnce string   `json:"api_key_secret_once"`
	FullTokenOnce    string   `json:"full_token_once"`
}
