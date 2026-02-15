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
			Message: "Username must be 3-15 characters",
		}
	}
	if !usernameRegex.MatchString(lower) {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username must be alphanumeric and underscores only, cannot start or end with underscore",
		}
	}
	if strings.Contains(lower, "__") {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username cannot contain consecutive underscores",
		}
	}
	if reservedUsernames[lower] {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username is reserved",
		}
	}
	return nil
}

func ValidatePostContent(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "content",
			Message: "Post content cannot be empty",
		}
	}
	if utf8.RuneCountInString(trimmed) > 140 {
		return &models.APIError{
			Code: models.ErrCodeContentLong,
			Message: "Post must be 140 characters or fewer",
		}
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "Email is required",
		}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "Invalid email format",
		}
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "password",
			Message: "Password must be at least 8 characters",
		}
	}
	return nil
}
