package services

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type OnlineService struct {
	clients map[string]*websocket.Conn
	mu      sync.Mutex
}

func NewOnlineService() *OnlineService {
	return &OnlineService{
		clients: make(map[string]*websocket.Conn),
	}
}

// Register adds a new user to the online users list
func (s *OnlineService) Register(userId string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[userId] = conn
	s.broadcastStatus(userId, "online")
}

// Unregister removes a user from the online users list
func (s *OnlineService) Unregister(userId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, userId)
	s.broadcastStatus(userId, "offline")
}

// broadcastStatus sends an online/offline status update to all clients
func (s *OnlineService) broadcastStatus(userId, status string) {
	for uid, conn := range s.clients {
		if uid != userId {
			err := conn.WriteJSON(map[string]string{
				"userId": userId,
				"status": status,
			})
			if err != nil {
				log.Printf("Error broadcasting status to %s: %v", uid, err)
			}
		}
	}
}

func (s *OnlineService) GetStatus(userId string) []map[string]string {
	var statuses []map[string]string
	s.mu.Lock()
	defer s.mu.Unlock()
	for uid := range s.clients {
		if uid != userId {
			statuses = append(statuses, map[string]string{
				"userId": uid,
				"status": "online",
			})
		}
	}
	return statuses
}
