// servit-go/internal/handlers/chat.go
package handlers

import (
	"log"
	"net/http"
	"servit-go/internal/chat"
	"servit-go/internal/middleware"

	"github.com/gorilla/websocket"
)

type Message struct {
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Content    string `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var manager = chat.NewConnectionManager()

// ChatHandler handles WebSocket requests for chat
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	manager.AddConnection(userID, conn)
	defer manager.RemoveConnection(userID)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if msg.ToUserID != "" {
			if recipientConn, ok := manager.GetConnection(msg.ToUserID); ok {
				if err := recipientConn.WriteJSON(msg); err != nil {
					log.Printf("Error sending message to user %s: %v", msg.ToUserID, err)
				}
			} else {
				log.Printf("User %s not connected", msg.ToUserID)
			}
		} else {
			log.Println("No recipient user ID provided")
		}
	}
}
