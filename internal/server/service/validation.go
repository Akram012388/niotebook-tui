package service

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9_]*[a-z0-9])?$`)

	reservedUsernames = map[string]bool{
		"admin": true, "root": true, "system": true,
		"niotebook": true, "api": true, "help": true,
		"support": true, "me": true, "about": true,
		"settings": true, "login": true, "register": true,
		"auth": true, "posts": true, "users": true,
		"timeline": true, "search": true, "explore": true,
	}
)

func ValidateUsername(username string) error {
	lower := strings.ToLower(username)
	length := utf8.RuneCountInString(lower)

	if length < 3 || length > 15 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "username must be 3-15 characters",
		}
	}
	if !usernameRegex.MatchString(lower) {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "username must be alphanumeric and underscores only, cannot start or end with underscore",
		}
	}
	if strings.Contains(lower, "__") {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "username cannot contain consecutive underscores",
		}
	}
	if reservedUsernames[lower] {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "username is reserved",
		}
	}
	return nil
}

func ValidatePostContent(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "content",
			Message: "post content cannot be empty",
		}
	}
	if utf8.RuneCountInString(trimmed) > 140 {
		return &models.APIError{
			Code: models.ErrCodeContentLong,
			Message: "post must be 140 characters or fewer",
		}
	}
	if containsControlChars(trimmed, true) {
		return &models.APIError{
			Code:    models.ErrCodeValidation,
			Field:   "content",
			Message: "post content contains invalid characters",
		}
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "email is required",
		}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "invalid email format",
		}
	}
	return nil
}

func ValidateDisplayName(name string) error {
	length := utf8.RuneCountInString(name)
	if name == "" || length > 50 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "display_name",
			Message: "display name must be 1-50 characters",
		}
	}
	if containsControlChars(name, false) {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "display_name",
			Message: "display name contains invalid characters",
		}
	}
	return nil
}

func ValidateBio(bio string) error {
	if utf8.RuneCountInString(bio) > 160 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "bio",
			Message: "bio must be 160 characters or fewer",
		}
	}
	if bio != "" && containsControlChars(bio, true) {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "bio",
			Message: "bio contains invalid characters",
		}
	}
	return nil
}

// containsControlChars checks for control characters.
// If allowNewline is true, \n and \r are permitted (for bio).
func containsControlChars(s string, allowNewline bool) bool {
	for _, r := range s {
		if r < 32 && r != '\t' {
			if allowNewline && (r == '\n' || r == '\r') {
				continue
			}
			return true
		}
	}
	return false
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "password",
			Message: "password must be at least 8 characters",
		}
	}
	return nil
}
