package services

import (
	"database/sql"
	"log"
	"servit-go/internal/manager"
	"servit-go/internal/models"
)

// ChannelChatServiceInterface defines methods for channel-based chat.
type ChannelChatServiceInterface interface {
	JoinChannel(channelID, userID string) chan models.ChannelMessage
	LeaveChannel(channelID, userID string)
	SendMessage(channelID, userID, username, message string)
}

type ChannelChatService struct {
	chatManager *manager.ChannelChatManager
	DB          *sql.DB
}

var _ ChannelChatServiceInterface = &ChannelChatService{
	chatManager: &manager.ChannelChatManager{},
	DB:          &sql.DB{},
}

func NewChannelChatService(db *sql.DB, chatManager *manager.ChannelChatManager) *ChannelChatService {
	return &ChannelChatService{
		chatManager: chatManager,
		DB:          db,
	}
}

func (s *ChannelChatService) JoinChannel(channelID, userID string) chan models.ChannelMessage {
	messageChan := s.chatManager.JoinChannel(channelID, userID)
	messageStructChan := make(chan models.ChannelMessage)

	go func() {
		for msg := range messageChan {
			messageStructChan <- models.ChannelMessage{
				Content:  msg.Content,
				UserId:   msg.UserId,
				Username: msg.Username,
			}
		}
		close(messageStructChan)
	}()

	return messageStructChan
}

func (s *ChannelChatService) LeaveChannel(channelID, userID string) {
	s.chatManager.LeaveChannel(channelID, userID)
}

func (s *ChannelChatService) SendMessage(channelID, userID, username, message string) {
	s.chatManager.SendMessage(channelID, userID, username, message)
}

func (s *ChannelChatService) SaveChannelMessage(channelId string, userId string, username string, content string) error {
	query := `
		INSERT INTO channel_messages (from_user_id, from_username, to_channel_id, content)
		VALUES ($1, $2, $3, $4)
		`

	// Execute the query to save the message
	_, err := s.DB.Exec(query, userId, username, channelId, content)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return err
	}
	return nil
}

func (s *ChannelChatService) FetchPaginatedChannelMessages(channelID string, page int) ([]models.SaveChannelMessage, error) {
	const pageSize = 30
	offset := (page - 1) * pageSize

	query := `
		SELECT from_user_id, from_username, to_channel_id, content, created_at
		FROM channel_messages
		WHERE to_channel_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.DB.Query(query, channelID, pageSize, offset)
	if err != nil {
		log.Printf("Failed to fetch messages: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []models.SaveChannelMessage
	for rows.Next() {
		var msg models.SaveChannelMessage
		if err := rows.Scan(&msg.UserID, &msg.Username, &msg.ChannelId, &msg.Content, &msg.CreatedAt); err != nil {
			log.Printf("Failed to scan message: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return nil, err
	}

	return messages, nil
}
