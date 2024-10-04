package models

type Message struct {
	ID           string `json:"id"`
	FromUserID   string `json:"from_user_id"`
	FromUserName string `json:"from_user_name"`
	ToUserID     string `json:"to_user_id"`
	Content      string `json:"content"`
	IsEdited     bool   `json:"is_edited"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
