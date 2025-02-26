package services

import (
	"database/sql"
	"fmt"
	"servit-go/internal/db"
	"servit-go/internal/models"
	"time"

	"github.com/gocql/gocql"
)

// ChatServiceInterface defines the methods to be mocked
type ChatServiceInterface interface {
	QueryMessages(fromUserID, toUserID string, pageSize int, pagingState []byte) ([]models.DMMessage, []byte, error)
	QueryChannelMessages(channelID string, pageSize int, pagingState []byte) ([]models.ChannelMessage, []byte, error)
}

type ChatService struct {
	DB *sql.DB
}

// NewChatService creates a new instance of ChatService with the given database connection.
func NewChatService(db *sql.DB) *ChatService {
	return &ChatService{
		DB: db,
	}
}

// createConversationID creates a canonical conversation ID using the two user IDs.
func createConversationID(userA, userB string) string {
	if userA < userB {
		return fmt.Sprintf("%s#%s", userA, userB)
	}
	return fmt.Sprintf("%s#%s", userB, userA)
}

// SaveDMMessage saves a direct message to ScyllaDB using conversation_id.
func SaveDMMessage(msg models.DMMessage) error {
	query := `INSERT INTO direct_messages 
		(conversation_id, timestamp, message_id, sender_id, receiver_id, content) 
		VALUES (?, ?, ?, ?, ?, ?)`

	senderUUID, err := gocql.ParseUUID(msg.SenderID)
	if err != nil {
		return fmt.Errorf("invalid sender UUID: %w", err)
	}

	receiverUUID, err := gocql.ParseUUID(msg.ReceiverID)
	if err != nil {
		return fmt.Errorf("invalid receiver UUID: %w", err)
	}

	messageUUID := gocql.TimeUUID()
	conversationID := createConversationID(msg.SenderID, msg.ReceiverID)

	return db.ScyllaSession.Query(query,
		conversationID,
		msg.Timestamp,
		messageUUID,
		senderUUID,
		receiverUUID,
		msg.Content,
	).Exec()
}

// QueryMessages retrieves messages between two users by using conversation_id.
func (c *ChatService) QueryMessages(userA, userB string, pageSize int, pagingState []byte) ([]models.DMMessage, []byte, error) {
	conversationID := createConversationID(userA, userB)

	query := `SELECT sender_id, receiver_id, timestamp, message_id, content 
		FROM direct_messages 
		WHERE conversation_id = ?
		ORDER BY timestamp DESC`

	// Use PageSize to set the maximum number of rows per page.
	q := db.ScyllaSession.Query(query, conversationID).PageSize(pageSize)
	if pagingState != nil {
		q = q.PageState(pagingState)
	}

	iter := q.Iter()

	var messages []models.DMMessage
	var (
		senderUUID   gocql.UUID
		receiverUUID gocql.UUID
		timestamp    time.Time
		messageID    gocql.UUID
		content      string
	)

	for i := 0; i < pageSize; i++ {
		if !iter.Scan(&senderUUID, &receiverUUID, &timestamp, &messageID, &content) {
			break
		}
		messages = append(messages, models.DMMessage{
			SenderID:   senderUUID.String(),
			ReceiverID: receiverUUID.String(),
			Content:    content,
			Timestamp:  timestamp,
		})
	}

	// Capture the paging state for subsequent queries.
	newPagingState := iter.PageState()
	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("query failed: %w", err)
	}

	return messages, newPagingState, nil
}

func SaveChannelMessage(msg models.ChannelMessage) error {
	// Parse channel and sender IDs as UUIDs.
	channelUUID, err := gocql.ParseUUID(msg.ChannelID)
	if err != nil {
		return fmt.Errorf("invalid channel UUID: %w", err)
	}
	senderUUID, err := gocql.ParseUUID(msg.SenderID)
	if err != nil {
		return fmt.Errorf("invalid sender UUID: %w", err)
	}

	// Use the message timestamp to determine the date bucket.
	// You can change the bucket granularity as needed (day, week, month, etc.).
	dateBucket := msg.Timestamp.Format("2006-01-02")

	// Generate a unique message ID based on time.
	messageUUID := gocql.TimeUUID()

	query := `INSERT INTO channel_messages 
		(channel_id, message_date, timestamp, message_id, sender_id, sender_username, content) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	return db.ScyllaSession.Query(query,
		channelUUID,
		dateBucket,
		msg.Timestamp,
		messageUUID,
		senderUUID,
		msg.SenderUsername,
		msg.Content,
	).Exec()
}

// QueryChannelMessages retrieves paginated messages for a given channel and date bucket.
func (c *ChatService) QueryChannelMessages(channelID string, pageSize int, pagingState []byte) ([]models.ChannelMessage, []byte, error) {
	// Parse channel ID to UUID.
	channelUUID, err := gocql.ParseUUID(channelID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid channel UUID: %w", err)
	}

	query := `SELECT sender_id, sender_username, timestamp, message_id, content 
		FROM channel_messages 
		WHERE channel_id = ? AND message_date = ?
		ORDER BY timestamp DESC`

	// Use PageSize to limit the number of rows per page.
	q := db.ScyllaSession.Query(query, channelUUID, time.Now().Format("2006-01-02")).PageSize(pageSize)
	if pagingState != nil {
		q = q.PageState(pagingState)
	}

	iter := q.Iter()

	var messages []models.ChannelMessage
	var (
		senderUUID     gocql.UUID
		senderUsername string
		timestamp      time.Time
		messageID      gocql.UUID
		content        string
	)

	// Loop up to pageSize times; break if no more rows.
	for i := 0; i < pageSize; i++ {
		if !iter.Scan(&senderUUID, &senderUsername, &timestamp, &messageID, &content) {
			break
		}
		messages = append(messages, models.ChannelMessage{
			SenderID:       senderUUID.String(),
			SenderUsername: senderUsername,
			ChannelID:      channelID,
			Content:        content,
			Timestamp:      timestamp,
		})
	}

	newPagingState := iter.PageState()
	if err := iter.Close(); err != nil {
		return nil, nil, fmt.Errorf("query failed: %w", err)
	}

	return messages, newPagingState, nil
}
