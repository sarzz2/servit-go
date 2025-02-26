package models

import (
	"encoding/json"
	"time"
)

// WSMessage is the common structure for all WebSocket messages.
type WSMessage struct {
	Type     string          `json:"type"`                // e.g. "switch_chat", "channel_message", "direct_message"
	ChatType string          `json:"chat_type,omitempty"` // "channel" or "dm"
	ChatID   string          `json:"chat_id,omitempty"`   // channel id or DM partner id
	Data     json.RawMessage `json:"data"`
}

// ChannelMessage represents a message sent in a channel.
type ChannelMessage struct {
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"username"`
	ChannelID      string    `json:"channel_id"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
}

// DMMessage represents a direct message.
type DMMessage struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

// ActiveChat represents the currently open chat window.
type ActiveChat struct {
	ChatType string `json:"chat_type"` // "channel" or "dm"
	ChatID   string `json:"chat_id"`   // active channel id or DM partner id
}

// TypingEvent represents a typing indicator payload.
type TypingEvent struct {
	FromUserID   string `json:"from_user_id"`
	ToUserID     string `json:"to_user_id"`
	FromUserName string `json:"from_user_name,omitempty"`
	ChatType     string `json:"chat_type,omitempty"` // "dm" or "channel"
	ChatID       string `json:"chat_id,omitempty"`   // For channel type, the channel id
}
