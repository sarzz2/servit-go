package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"servit-go/internal/middleware"
	"servit-go/internal/models"
	"servit-go/internal/services"

	"github.com/gin-gonic/gin"
)

// WsHandler upgrades the HTTP connection to a WebSocket and creates a new Client.
func WsHandler(c *gin.Context, r *http.Request, hub *services.Hub) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	username := r.Context().Value(middleware.UserNameKey).(string)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	client := &services.Client{
		ID:         userID,
		Username:   username,
		Conn:       conn,
		Send:       make(chan []byte, 1024),
		Hub:        hub,
		ActiveChat: nil,
		Unread:     make(map[string]int),
	}
	hub.Register(client)
	go client.ReadPump()
	client.WritePump()
}

// FetchHistoricalMessages returns stored messages for a channel or DM.
// In production, this should query your PostgreSQL DB via your FastAPI service.
func FetchHistoricalMessages(c *gin.Context) {
	chatType := c.Query("chat_type") // "channel" or "dm"
	chatID := c.Query("chat_id")
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	var messages interface{}
	if chatType == "channel" {
		messages = []models.ChannelMessage{
			{SenderID: "user1", ChannelID: chatID, Content: "Historical message 1", Timestamp: time.Now().Add(-15 * time.Minute)},
			{SenderID: "user2", ChannelID: chatID, Content: "Historical message 2", Timestamp: time.Now().Add(-10 * time.Minute)},
		}
	} else if chatType == "dm" {
		currentUser := c.GetString("user_id")
		messages = []models.DMMessage{
			{SenderID: chatID, ReceiverID: currentUser, Content: "Historical DM message 1", Timestamp: time.Now().Add(-20 * time.Minute)},
			{SenderID: currentUser, ReceiverID: chatID, Content: "Historical DM message 2", Timestamp: time.Now().Add(-18 * time.Minute)},
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat_type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"messages": messages,
	})
}
