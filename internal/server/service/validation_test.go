package service_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid simple", "akram", false},
		{"valid with underscore", "code_ninja", false},
		{"valid numbers", "user42", false},
		{"minimum length 3", "abc", false},
		{"maximum length 15", "abcdefghijklmno", false},
		{"too short", "ab", true},
		{"too long 16", "abcdefghijklmnop", true},
		{"leading underscore", "_akram", true},
		{"trailing underscore", "akram_", true},
		{"consecutive underscores", "code__ninja", true},
		{"special chars", "akram!", true},
		{"spaces", "ak ram", true},
		{"uppercase accepted", "Akram", false},
		{"reserved admin", "admin", true},
		{"reserved api", "api", true},
		{"reserved root", "root", true},
		{"empty string", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePostContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"valid short", "Hello!", false},
		{"exactly 140", strings.Repeat("a", 140), false},
		{"141 chars", strings.Repeat("a", 141), true},
		{"empty", "", true},
		{"whitespace only", "   \n\t  ", true},
		{"with newlines", "line1\nline2", false},
		{"trimmed within limit", "  " + strings.Repeat("a", 140) + "  ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePostContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostContent(%q) error = %v, wantErr %v", tt.content, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"valid with subdomain", "user@sub.example.com", false},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Akram", false},
		{"empty", "", true},
		{"max 50", strings.Repeat("a", 50), false},
		{"too long 51", strings.Repeat("a", 51), true},
		{"control char", "hello\x00world", true},
		{"tab allowed", "hello\tworld", false},
		{"newline rejected", "hello\nworld", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateDisplayName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDisplayName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBio(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Building things.", false},
		{"empty allowed", "", false},
		{"max 160", strings.Repeat("a", 160), false},
		{"too long 161", strings.Repeat("a", 161), true},
		{"control char", "bio\x00text", true},
		{"newline allowed", "line1\nline2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBio(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBio(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid 8 chars", "12345678", false},
		{"valid long", "a-very-secure-password", false},
		{"too short 7", "1234567", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
