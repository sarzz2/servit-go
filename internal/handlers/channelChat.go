package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"servit-go/internal/middleware"
	"servit-go/internal/models"
	"servit-go/internal/services"
	"strconv"

	"github.com/gorilla/websocket"
)

func ChannelChatHandler(w http.ResponseWriter, r *http.Request, chatService *services.ChannelChatService) {
	userId := r.Context().Value(middleware.UserIDKey).(string)
	username := r.Context().Value(middleware.UserNameKey).(string)
	channelID := r.URL.Query().Get("channel_id")

	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("User %s (%s) connected to channel %s", username, userId, channelID)

	// Join channel and ensure cleanup on exit
	userChannel := chatService.JoinChannel(channelID, userId)
	defer chatService.LeaveChannel(channelID, userId)

	// Start reading messages from WebSocket and broadcasting
	go handleIncomingMessages(conn, chatService, channelID)
	// Write channel messages to WebSocket
	for msg := range userChannel {
		message := models.ChannelMessage{
			UserId:   msg.UserId,
			Username: msg.Username,
			Content:  msg.Content,
		}
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

func handleIncomingMessages(conn *websocket.Conn, chatService *services.ChannelChatService, channelID string) {
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		var message models.ChannelMessage
		if err := json.Unmarshal(msgBytes, &message); err != nil {
			continue
		}

		chatService.SendMessage(channelID, message.UserId, message.Username, message.Content)

		// Save the message to the database
		if err := chatService.SaveChannelMessage(channelID, message.UserId, message.Username, message.Content); err != nil {
			log.Printf("Failed to save message: %v", err)
		}
	}
}

// FetchPaginatedChannelMessages handles fetching paginated messages for a channel
func FetchPaginatedChannelMessages(w http.ResponseWriter, r *http.Request, chatService *services.ChannelChatService) {
	channelID := r.URL.Query().Get("channel_id")
	page := r.URL.Query().Get("page")

	if channelID == "" || page == "" {
		http.Error(w, "Missing required query parameters", http.StatusBadRequest)
		return
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		return
	}
	messages, err := chatService.FetchPaginatedChannelMessages(channelID, pageInt)
	if err != nil {
		log.Printf("Failed to fetch paginated messages: %v", err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		log.Printf("JSON encode error: %v", err)
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
	}
}
