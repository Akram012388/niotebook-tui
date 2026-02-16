package models_test

import (
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func TestAPIErrorString(t *testing.T) {
	err := &models.APIError{
		Code:    models.ErrCodeValidation,
		Message: "username must be 3-15 characters",
	}

	got := err.Error()
	want := "validation_error: username must be 3-15 characters"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIErrorImplementsError(t *testing.T) {
	var err error = &models.APIError{Code: "test", Message: "msg"}
	if err.Error() != "test: msg" {
		t.Errorf("unexpected error string: %q", err.Error())
	}
}
