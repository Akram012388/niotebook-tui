package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func HandleGetUser(userSvc *service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "me" {
			id = middleware.UserIDFromContext(r.Context())
		}
		if id == "" {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeUnauthorized,
				Message: "authentication required",
			})
			return
		}

		user, err := userSvc.GetUserByID(r.Context(), id)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
	}
}

func HandleGetUserPosts(postSvc *service.PostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if userID == "me" {
			userID = middleware.UserIDFromContext(r.Context())
		}
		if userID == "" {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "user id is required",
			})
			return
		}

		cursor := time.Now()
		if c := r.URL.Query().Get("cursor"); c != "" {
			parsed, err := time.Parse(time.RFC3339, c)
			if err != nil {
				writeAPIError(w, &models.APIError{
					Code:    models.ErrCodeValidation,
					Message: "invalid cursor format, expected RFC3339",
				})
				return
			}
			cursor = parsed
		}

		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			parsed, err := strconv.Atoi(l)
			if err != nil || parsed < 1 || parsed > 100 {
				writeAPIError(w, &models.APIError{
					Code:    models.ErrCodeValidation,
					Message: "limit must be between 1 and 100",
				})
				return
			}
			limit = parsed
		}

		posts, err := postSvc.GetUserPosts(r.Context(), userID, cursor, limit)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		resp := models.TimelineResponse{
			Posts:   posts,
			HasMore: len(posts) == limit,
		}

		if len(posts) > 0 {
			c := posts[len(posts)-1].CreatedAt.Format(time.RFC3339Nano)
			resp.NextCursor = &c
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func HandleUpdateUser(userSvc *service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		if userID == "" {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeUnauthorized,
				Message: "authentication required",
			})
			return
		}

		var updates models.UserUpdate
		if err := decodeBody(w, r, &updates); err != nil {
			writeAPIError(w, &models.APIError{
				Code:    models.ErrCodeValidation,
				Message: "invalid request body",
			})
			return
		}

		user, err := userSvc.UpdateUser(r.Context(), userID, &updates)
		if err != nil {
			writeAPIError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
	}
}
