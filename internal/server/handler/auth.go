package handler

import (
	"net/http"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func HandleRegister(authSvc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := decodeBody(w, r, &req); err != nil {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "invalid request body",
			})
			return
		}

		resp, err := authSvc.Register(r.Context(), &req)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, resp)
	}
}

func HandleLogin(authSvc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := decodeBody(w, r, &req); err != nil {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "invalid request body",
			})
			return
		}

		resp, err := authSvc.Login(r.Context(), &req)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func HandleRefresh(authSvc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RefreshRequest
		if err := decodeBody(w, r, &req); err != nil {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "invalid request body",
			})
			return
		}

		tokens, err := authSvc.Refresh(r.Context(), req.RefreshToken)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"tokens": tokens})
	}
}
