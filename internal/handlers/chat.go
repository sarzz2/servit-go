package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"servit-go/internal/manager"
	"servit-go/internal/middleware"
	"servit-go/internal/models"
	"servit-go/internal/services"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ChatHandler handles WebSocket requests for chat
func ChatHandler(w http.ResponseWriter, r *http.Request, chatService *services.ChatService) {
	manager := manager.NewConnectionManager()
	userId := r.Context().Value(middleware.UserIDKey).(string)
	userName := r.Context().Value(middleware.UserNameKey).(string)

	// Upgrade the connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Add the user connection to the manager using userID
	manager.AddConnection(userId, conn)

	// Wait for the initial message from the client that contains the recipient ID
	var initPayload struct {
		ToUserID string `json:"toUserID"`
		Type     string `json:"type"`
	}

	if err := conn.ReadJSON(&initPayload); err != nil {
		log.Printf("Failed to read initial payload: %v", err)
		return
	}

	toUserID := initPayload.ToUserID
	if toUserID == "" {
		log.Println("No recipient user ID provided")
		return
	}

	// Send "not_typing" indicator to the recipient upon disconnection
	defer func() {
		manager.RemoveConnection(userId)

		typingIndicator := models.TypingIndicator{
			Type:       "not_typing",
			FromUserID: userId,
			ToUserID:   toUserID,
			Typing:     false,
		}
		if recipientConn, ok := manager.GetConnection(toUserID); ok {
			if err := recipientConn.WriteJSON(typingIndicator); err != nil {
				log.Printf("Error sending not typing indicator on disconnection to user %s: %v", toUserID, err)
			}
		}
	}()

	// Fetch chat history between the current user and the recipient
	messages, err := chatService.FetchMessages(userId, toUserID)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
	} else {
		if err := conn.WriteJSON(messages); err != nil {
			log.Printf("Error sending chat history: %v", err)
		}
	}

	// Handle the rest of the WebSocket communication (receiving and sending messages)
	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		msg.FromUserID = userId
		msg.FromUserName = userName

		// Handle typing indicator
		if msg.Type == "typing" {
			typingIndicator := models.TypingIndicator{
				Type:       "typing",
				FromUserID: userId,
				ToUserID:   toUserID,
				Typing:     true,
			}

			if recipientConn, ok := manager.GetConnection(toUserID); ok {
				if err := recipientConn.WriteJSON(typingIndicator); err != nil {
					log.Printf("Error sending typing indicator to user %s: %v", toUserID, err)
				}
			}
			continue
		}

		// Handle "not_typing" indicator
		if msg.Type == "not_typing" {
			typingIndicator := models.TypingIndicator{
				Type:       "not_typing",
				FromUserID: userId,
				ToUserID:   toUserID,
				Typing:     false,
			}

			if recipientConn, ok := manager.GetConnection(toUserID); ok {
				if err := recipientConn.WriteJSON(typingIndicator); err != nil {
					log.Printf("Error sending not typing indicator to user %s: %v", toUserID, err)
				}
			}
			continue
		}

		// Save the message using ChatService
		if err := chatService.SaveMessage(msg); err != nil {
			log.Printf("Error saving message: %v", err)
		}

		// Send the message to the recipient if connected
		if recipientConn, ok := manager.GetConnection(msg.ToUserID); ok {
			if err := recipientConn.WriteJSON(msg); err != nil {
				log.Printf("Error sending message to user %s: %v", msg.ToUserID, err)
			}
		} else {
			log.Printf("User %s not connected", msg.ToUserID)
		}
	}
}

// FetchPaginatedMessagesHandler handles requests for paginated messages based on page number
func FetchPaginatedMessagesHandler(w http.ResponseWriter, r *http.Request, chatService *services.ChatService) {
	// Retrieve query params: toUserID, page
	toUserID := r.URL.Query().Get("to_user_id")
	pageStr := r.URL.Query().Get("page")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 2
	}

	const limit = 25

	// Get fromUserID from context (authenticated user)
	fromUserID := r.Context().Value(middleware.UserIDKey).(string)

	// Fetch paginated messages
	messages, err := chatService.FetchPaginatedMessages(fromUserID, toUserID, page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
		return
	}
}
