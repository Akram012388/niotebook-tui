package handler

import (
	"net/http"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func HandleCreatePost(postSvc *service.PostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		if userID == "" {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeUnauthorized,
				Message: "authentication required",
			})
			return
		}

		var body struct {
			Content string `json:"content"`
		}
		if err := decodeBody(w, r, &body); err != nil {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "invalid request body",
			})
			return
		}

		post, err := postSvc.CreatePost(r.Context(), userID, body.Content)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{"post": post})
	}
}

func HandleGetPost(postSvc *service.PostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "post id is required",
			})
			return
		}

		post, err := postSvc.GetPostByID(r.Context(), id)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"post": post})
	}
}
