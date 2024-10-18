package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"servit-go/internal/middleware"
	"servit-go/internal/services"
)

func OnlineHandler(c http.ResponseWriter, r *http.Request, onlineService *services.OnlineService) {
	userId := r.Context().Value(middleware.UserIDKey).(string)

	conn, err := upgrader.Upgrade(c, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()
	// Register the user connection
	onlineService.Register(userId, conn)
	defer onlineService.Unregister(userId)

	// Listen for messages from the client (optional, can listen for pings, etc.)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
	}
}

// Go Gin Handler (for fetching online statuses)
func GetFriendsOnlineStatus(c http.ResponseWriter, r *http.Request, onlineService *services.OnlineService) {
	userId := r.Context().Value(middleware.UserIDKey).(string)
	onlineStatuses := onlineService.GetStatus(userId)

	c.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(c).Encode(onlineStatuses); err != nil {
		http.Error(c, err.Error(), http.StatusInternalServerError)
	}
}
