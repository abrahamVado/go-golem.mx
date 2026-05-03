package errors

import "net/http"

var (
	ErrUnauthorized = New("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrNotFound     = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest   = New("BAD_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrInternal     = New("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
)
