package app

import "github.com/Akram012388/niotebook-tui/internal/models"

// Auth messages
type MsgAuthSuccess struct {
	User   *models.User
	Tokens *models.TokenPair
}
type MsgAuthExpired struct{}
type MsgAuthError struct {
	Message string
	Field   string
}

// Timeline messages
type MsgTimelineLoaded struct {
	Posts      []models.Post
	NextCursor string
	HasMore    bool
}
type MsgTimelineRefreshed struct {
	Posts      []models.Post
	NextCursor string
	HasMore    bool
}

// Post messages
type MsgPostPublished struct{ Post models.Post }

// Profile messages
type MsgProfileLoaded struct {
	User  *models.User
	Posts []models.Post
}
type MsgProfileUpdated struct{ User *models.User }

// Navigation messages
type MsgSwitchToRegister struct{}
type MsgSwitchToLogin struct{}

// Generic messages
type MsgAPIError struct{ Message string }
type MsgStatusClear struct{}

// Server connection messages (splash screen)
type MsgServerConnected struct{}
type MsgServerFailed struct{ Err string }

// Splash animation messages
type MsgRevealTick struct{}
