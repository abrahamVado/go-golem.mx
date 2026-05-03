package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// TOKEN SECURITY UTILITIES
// -----------------------------------------------------------------------------
//
// This file handles:
//
//   - JWT access token creation
//   - JWT access token parsing
//   - opaque refresh/reset token generation
//   - secure token hashing
//
// Security model:
//
//   - Access tokens are short-lived JWTs
//   - Refresh/reset tokens should be opaque random strings
//   - Opaque tokens should be stored hashed, never plaintext
//
// -----------------------------------------------------------------------------

const (
	// AccessTokenType identifies this token as an access token.
	//
	// This prevents accidentally accepting refresh/reset tokens where
	// access tokens are required.
	AccessTokenType = "access"

	// MinJWTSecretLength prevents weak HMAC secrets.
	//
	// HS256 requires a strong shared secret.
	// Recommended minimum: 32 bytes.
	MinJWTSecretLength = 32

	// OpaqueTokenBytes controls refresh/reset token entropy.
	//
	// 48 random bytes gives strong security while keeping tokens URL-safe.
	OpaqueTokenBytes = 48
)

// AccessClaims defines the JWT payload for authenticated API access.
//
// Required tenant identity:
//
//   - user_id
//   - company_id
//
// Optional scope:
//
//   - branch_id
//
// Registered claims include:
//
//   - sub: user ID
//   - exp: expiration time
//   - iat: issued-at time
//   - nbf: not-before time
//   - jti: unique token ID
//
type AccessClaims struct {
	UserID    uuid.UUID  `json:"user_id"`
	CompanyID uuid.UUID  `json:"company_id"`
	BranchID  *uuid.UUID `json:"branch_id,omitempty"`

	TokenType string `json:"typ"`

	jwt.RegisteredClaims
}

// NewAccessToken creates a signed JWT access token.
//
// Access tokens should be:
//
//   - short-lived
//   - signed
//   - validated on every private request
//   - never stored in the database as the source of truth
//
// Recommended TTL:
//
//   5 to 15 minutes
//
func NewAccessToken(
	secret string,
	ttl time.Duration,
	userID uuid.UUID,
	companyID uuid.UUID,
	branchID *uuid.UUID,
) (string, error) {
	if err := validateJWTSecret(secret); err != nil {
		return "", err
	}

	if ttl <= 0 {
		return "", errors.New("access token TTL must be greater than zero")
	}

	if userID == uuid.Nil {
		return "", errors.New("user ID is required")
	}

	if companyID == uuid.Nil {
		return "", errors.New("company ID is required")
	}

	now := time.Now().UTC()

	claims := AccessClaims{
		UserID:    userID,
		CompanyID: companyID,
		BranchID:  branchID,
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}

// ParseAccessToken parses and validates an access JWT.
//
// Validation includes:
//
//   - HMAC signing method
//   - signature verification
//   - expiration
//   - not-before
//   - token validity
//   - required custom claims
//   - token type
//
// This function should be called only from authentication middleware.
func ParseAccessToken(
	secret string,
	tokenString string,
) (*AccessClaims, error) {
	if err := validateJWTSecret(secret); err != nil {
		return nil, err
	}

	if tokenString == "" {
		return nil, errors.New("access token is required")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessClaims{},
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected JWT signing method")
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.TokenType != AccessTokenType {
		return nil, errors.New("invalid token type")
	}

	if claims.UserID == uuid.Nil {
		return nil, errors.New("missing user ID claim")
	}

	if claims.CompanyID == uuid.Nil {
		return nil, errors.New("missing company ID claim")
	}

	if claims.Subject != claims.UserID.String() {
		return nil, errors.New("subject does not match user ID")
	}

	return claims, nil
}

// NewOpaqueToken creates a cryptographically secure random token.
//
// Use opaque tokens for:
//
//   - refresh tokens
//   - password reset tokens
//   - email verification tokens
//   - invite tokens
//
// Important:
//
// Opaque tokens should be shown to the user/client only once.
// Store only HashToken(token) in the database.
func NewOpaqueToken() (string, error) {
	randomBytes := make([]byte, OpaqueTokenBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

// HashToken hashes an opaque token before database storage.
//
// Why hash tokens:
//
//   - database leaks do not expose usable tokens
//   - refresh/reset tokens remain safer at rest
//
// Recommended use:
//
//   rawToken := NewOpaqueToken()
//   tokenHash := HashToken(rawToken)
//   store tokenHash in database
//
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

// validateJWTSecret enforces minimum HMAC secret strength.
//
// HS256 uses a shared secret.
// Weak secrets can be brute-forced if a token leaks.
func validateJWTSecret(secret string) error {
	if len(secret) < MinJWTSecretLength {
		return errors.New("JWT secret must be at least 32 characters")
	}

	return nil
}