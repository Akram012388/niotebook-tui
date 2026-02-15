package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func HandleTimeline(postSvc *service.PostService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		posts, err := postSvc.GetTimeline(r.Context(), cursor, limit)
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
