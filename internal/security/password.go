package security

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// -----------------------------------------------------------------------------
// PASSWORD SECURITY UTILITIES
// -----------------------------------------------------------------------------
//
// This package provides password hashing and verification utilities.
//
// Design goals:
//
//   - secure password storage
//   - predictable hashing behavior
//   - safe default configuration
//   - resistance to brute-force attacks
//
// Security principles:
//
//   - passwords are never stored in plaintext
//   - hashes are generated using bcrypt
//   - password comparison uses constant-time verification
//   - hashing cost is controlled centrally
//
// -----------------------------------------------------------------------------

// Recommended bcrypt limits.
//
// bcrypt cost controls CPU work required to compute a hash.
//
// Higher cost:
//   more secure
//   slower login
//
// Lower cost:
//   faster login
//   weaker protection
//
// Typical production values:
//
//   10 -> minimum acceptable
//   12 -> recommended default
//   14 -> high security
//
const (
	MinBcryptCost = 10
	MaxBcryptCost = 16
	DefaultBcryptCost = 12
)

// HashPassword securely hashes a plaintext password using bcrypt.
//
// Behavior:
//
//   - validates password input
//   - enforces safe cost bounds
//   - returns bcrypt hash string
//
// This function should be used for:
//
//   - user registration
//   - password reset
//   - password update
//
// Never:
//
//   - store plaintext passwords
//   - reuse password hashes across systems
//
func HashPassword(password string, cost int) (string, error) {

	// -------------------------------------------------------------------------
	// Validate password presence
	// -------------------------------------------------------------------------

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// -------------------------------------------------------------------------
	// Normalize cost value
	// -------------------------------------------------------------------------

	if cost < MinBcryptCost {
		cost = MinBcryptCost
	}

	if cost > MaxBcryptCost {
		cost = MaxBcryptCost
	}

	// -------------------------------------------------------------------------
	// Generate bcrypt hash
	// -------------------------------------------------------------------------

	hashBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		cost,
	)

	if err != nil {
		return "", err
	}

	return string(hashBytes), nil
}

// CheckPassword verifies whether a plaintext password matches a stored hash.
//
// Behavior:
//
//   - constant-time comparison
//   - resistant to timing attacks
//   - returns boolean result
//
// This function should be used for:
//
//   - login authentication
//   - password verification
//
func CheckPassword(hash, password string) bool {

	if hash == "" || password == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err == nil
}

// -----------------------------------------------------------------------------
// OPTIONAL PASSWORD POLICY VALIDATION
// -----------------------------------------------------------------------------
//
// This helper validates password strength before hashing.
//
// Recommended minimum policy:
//
//   - at least 8 characters
//   - contains uppercase letter
//   - contains lowercase letter
//   - contains number
//
// You can call this before HashPassword().
//
// Example:
//
//   if err := ValidatePassword(password); err != nil {
//       return response.BadRequest(...)
//   }
//
func ValidatePassword(password string) error {

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	var hasUpper bool
	var hasLower bool
	var hasNumber bool

	for _, r := range password {

		if unicode.IsUpper(r) {
			hasUpper = true
		}

		if unicode.IsLower(r) {
			hasLower = true
		}

		if unicode.IsDigit(r) {
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain an uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain a lowercase letter")
	}

	if !hasNumber {
		return errors.New("password must contain a number")
	}

	return nil
}