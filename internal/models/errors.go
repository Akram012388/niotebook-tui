package models

import "fmt"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeValidation  = "validation_error"
	ErrCodeContentLong = "content_too_long"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeTokenExpired = "token_expired"
	ErrCodeForbidden   = "forbidden"
	ErrCodeNotFound    = "not_found"
	ErrCodeConflict    = "conflict"
	ErrCodeRateLimited = "rate_limited"
	ErrCodeInternal    = "internal_error"
)
