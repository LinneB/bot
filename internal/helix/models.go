package helix

import (
	"fmt"
	"time"
)

// User returned from the /users endpoint
type User struct {
	Id          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	// Type of user can be "admin", "global_mod", "staff", ""
	// See: https://dev.twitch.tv/docs/api/reference/#get-users
	Type string `json:"type"`
	// Type of broadcaster can be "affiliate" "partner" ""
	// See: https://dev.twitch.tv/docs/api/reference/#get-users
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	// Note that this only works with user:read:email AND the looked up user is the token holder
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Stream returned from the /streams endpoint
type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Tags         []string  `json:"tags"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	IsMature     bool      `json:"is_mature"`
}

type ErrorStatus struct {
	StatusCode int
}

func (e *ErrorStatus) Error() string {
	return fmt.Sprintf("Requested resource returned unhandled status code: %d", e.StatusCode)
}
