package services

import (
	"database/sql"
	"log"
	"servit-go/internal/models"
	"time"
)

// ChatServiceInterface defines the methods to be mocked
type ChatServiceInterface interface {
	SaveMessage(msg models.Message) error
	FetchMessages(fromUserID, toUserID string) ([]models.Message, error)
	FetchPaginatedMessages(fromUserID, toUserID string, page, limit int) ([]models.Message, error)
	FetchUserChatHistory(userID string) ([]models.ChatHistory, error)
}

type ChatService struct {
	DB *sql.DB
}

var _ ChatServiceInterface = &ChatService{
	DB: &sql.DB{},
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
		ORDER BY created_at DESC LIMIT 25
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

// FetchPaginatedMessages retrieves a paginated set of messages based on the page number.
func (s *ChatService) FetchPaginatedMessages(fromUserID, toUserID string, page, limit int) ([]models.Message, error) {
	// Calculate the offset based on the page number
	offset := (page - 1) * limit

	query := `
		SELECT id, from_user_id, to_user_id, content, is_edited, created_at, updated_at
		FROM direct_messages
		WHERE (from_user_id = $1 AND to_user_id = $2)
		OR (from_user_id = $2 AND to_user_id = $1)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := s.DB.Query(query, fromUserID, toUserID, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch paginated messages: %v", err)
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

	// Reverse the order to maintain chronological order (since we ordered by DESC)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (s *ChatService) FetchUserChatHistory(userID string) ([]models.ChatHistory, error) {
	query := `
		SELECT DISTINCT ON (dm.to_user_id) 
			u.username, 
			    u.id,
				u.profile_picture_url,
			dm.updated_at
		FROM direct_messages dm
		JOIN users u ON u.id = dm.to_user_id
		WHERE dm.from_user_id = $1
		ORDER BY dm.to_user_id, dm.updated_at DESC;
	`

	rows, err := s.DB.Query(query, userID)
	if err != nil {
		log.Printf("Failed to fetch chat history: %v", err)
		return nil, err
	}
	defer rows.Close()

	var chatHistory []models.ChatHistory
	for rows.Next() {
		var msg models.ChatHistory
		if err := rows.Scan(&msg.Username, &msg.FriendId, &msg.ProfilePictureURL, &msg.UpdatedAt); err != nil {
			log.Printf("Failed to scan message: %v", err)
			return nil, err
		}

		chatHistory = append(chatHistory, models.ChatHistory{
			FriendId:          msg.FriendId,
			ProfilePictureURL: msg.ProfilePictureURL,
			Username:          msg.Username,
			UpdatedAt:         msg.UpdatedAt,
		})
	}
	return chatHistory, nil

}
