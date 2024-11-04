package models

type ChannelMessage struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

type SaveChannelMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	ChannelId string `json:"channel_id"`
	CreatedAt string `json:"created_at"`
}
