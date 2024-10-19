package models

type Message struct {
	ID           string `json:"id"`
	FromUserID   string `json:"from_user_id"`
	FromUserName string `json:"from_user_name"`
	ToUserID     string `json:"to_user_id"`
	Content      string `json:"content"`
	IsEdited     bool   `json:"is_edited"`
	Type         string `json:"type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type TypingIndicator struct {
	Type       string `json:"type"` // e.g., "typing" or "not_typing"
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Typing     bool   `json:"typing"` // if typing, false if not typing
}

type ChatHistory struct {
	Username          string `json:"username"`
	FriendId          string `json:"friend_id"`
	ProfilePictureURL string `json:"profile_picture_url"`
	UpdatedAt         string `json:"updated_at"`
}
