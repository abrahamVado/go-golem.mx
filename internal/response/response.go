package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// -----------------------------------------------------------------------------
// API RESPONSE CONTRACT
// -----------------------------------------------------------------------------
//
// This package defines the standard response format for the API.
//
// All endpoints must return responses using this structure.
//
// Design goals:
//
//   - consistent response format
//   - predictable error handling
//   - frontend compatibility
//   - observability readiness
//   - easy debugging
//
// Standard response shape:
//
//   Success:
//
//     {
//       "success": true,
//       "data": { ... }
//     }
//
//   Failure:
//
//     {
//       "success": false,
//       "error": {
//         "code": "FORBIDDEN",
//         "message": "Missing required permission"
//       }
//     }
//
// -----------------------------------------------------------------------------

// ErrorBody represents an API error.
//
// Code:
//   machine-readable identifier
//
// Message:
//   human-readable description
//
// Examples:
//
//   INVALID_TOKEN
//   FORBIDDEN
//   VALIDATION_ERROR
//   RATE_LIMITED
//
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Envelope is the standard API response container.
//
// This structure ensures:
//
//   - consistent JSON responses
//   - frontend parsing stability
//   - future extensibility
//
// Fields:
//
//   Success:
//     indicates operation result
//
//   Data:
//     payload for successful responses
//
//   Error:
//     error information for failed responses
//
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// -----------------------------------------------------------------------------
// SUCCESS RESPONSES
// -----------------------------------------------------------------------------

// OK returns HTTP 200.
//
// Use for:
//
//   - successful reads
//   - successful updates
//   - successful operations
//
func OK(c *gin.Context, data any) {
	c.JSON(
		http.StatusOK,
		Envelope{
			Success: true,
			Data:    data,
		},
	)
}

// Created returns HTTP 201.
//
// Use for:
//
//   - resource creation
//   - registration
//   - entity creation
//
func Created(c *gin.Context, data any) {
	c.JSON(
		http.StatusCreated,
		Envelope{
			Success: true,
			Data:    data,
		},
	)
}

// NoContent returns HTTP 204.
//
// Use for:
//
//   - successful deletion
//   - successful idempotent operations
//   - operations without response body
//
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
// ERROR RESPONSES
// -----------------------------------------------------------------------------

// Fail returns a standardized error response.
//
// Behavior:
//
//   - sends JSON error
//   - aborts middleware chain
//
// This function should be used for:
//
//   - authentication failures
//   - authorization failures
//   - validation errors
//   - business logic errors
//   - system errors
//
// Example:
//
//   response.Fail(
//       c,
//       403,
//       "FORBIDDEN",
//       "Missing required permission",
//   )
//
func Fail(
	c *gin.Context,
	status int,
	code string,
	message string,
) {
	c.JSON(
		status,
		Envelope{
			Success: false,
			Error: &ErrorBody{
				Code:    code,
				Message: message,
			},
		},
	)

	c.Abort()
}

// -----------------------------------------------------------------------------
// COMMON ERROR HELPERS
// -----------------------------------------------------------------------------
//
// These helpers reduce duplication and improve readability.
//

func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, "BAD_REQUEST", msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", msg)
}

func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, "FORBIDDEN", msg)
}

func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, "NOT_FOUND", msg)
}

func Conflict(c *gin.Context, msg string) {
	Fail(c, http.StatusConflict, "CONFLICT", msg)
}

func TooManyRequests(c *gin.Context, msg string) {
	Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", msg)
}

func Internal(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}