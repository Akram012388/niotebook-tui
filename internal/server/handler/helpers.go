package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeAPIError(w http.ResponseWriter, err error) {
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		status := errorCodeToHTTPStatus(apiErr.Code)
		writeJSON(w, status, map[string]interface{}{"error": apiErr})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"error": models.APIError{
			Code:    models.ErrCodeInternal,
			Message: "Something went wrong. Please try again.",
		},
	})
}

func errorCodeToHTTPStatus(code string) int {
	switch code {
	case models.ErrCodeValidation, models.ErrCodeContentLong:
		return http.StatusBadRequest
	case models.ErrCodeUnauthorized, models.ErrCodeTokenExpired:
		return http.StatusUnauthorized
	case models.ErrCodeForbidden:
		return http.StatusForbidden
	case models.ErrCodeNotFound:
		return http.StatusNotFound
	case models.ErrCodeConflict:
		return http.StatusConflict
	case models.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func decodeBody(r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
