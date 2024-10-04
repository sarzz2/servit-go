package services

import (
	"database/sql"
	"fmt"
	"log"
	"servit-go/internal/models"
	"time"
)

type ChatService struct {
	DB *sql.DB
}

// NewChatService creates a new instance of ChatService with the given database connection.
func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{
		DB: db,
	}
}

// SaveMessage saves a chat message to the database.
func (s *ChatService) SaveMessage(msg models.Message) error {
	query := `
		INSERT INTO direct_messages (from_user_id, to_user_id, content, is_edited, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now()
	msg.CreatedAt = now.Format(time.RFC3339)
	msg.UpdatedAt = now.Format(time.RFC3339)
	fmt.Print(msg)

	// Execute the query to save the message
	_, err := s.DB.Exec(query, msg.FromUserID, msg.ToUserID, msg.Content, msg.IsEdited, msg.CreatedAt, msg.UpdatedAt)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return err
	}
	return nil
}

// FetchMessages retrieves all messages between two users, ordered by the time of creation.
func (s *ChatService) FetchMessages(fromUserID, toUserID string) ([]models.Message, error) {
	query := `
		SELECT id, from_user_id, to_user_id, content, is_edited, created_at, updated_at
		FROM direct_messages
		WHERE (from_user_id = $1 AND to_user_id = $2)
		OR (from_user_id = $2 AND to_user_id = $1)
		ORDER BY created_at ASC
	`

	rows, err := s.DB.Query(query, fromUserID, toUserID)
	if err != nil {
		log.Printf("Failed to fetch messages: %v", err)
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.FromUserID, &msg.ToUserID, &msg.Content, &msg.IsEdited, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
			log.Printf("Failed to scan message: %v", err)
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
