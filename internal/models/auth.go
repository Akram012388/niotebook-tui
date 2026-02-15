package models

import "time"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AuthResponse struct {
	User   *User     `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

type TimelineResponse struct {
	Posts      []Post  `json:"posts"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}
