package manager

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type OnlineManager struct {
	clients   map[string]*websocket.Conn // map of userId to websocket connection
	broadcast chan UserStatus            // channel for broadcasting online status
	mu        sync.Mutex                 // to ensure safe concurrent access
}

type UserStatus struct {
	UserID string `json:"userId"`
	Status string `json:"status"` // "online" or "offline"
}

func NewOnlineManager() *OnlineManager {
	return &OnlineManager{
		clients:   make(map[string]*websocket.Conn),
		broadcast: make(chan UserStatus),
	}
}

func (m *OnlineManager) AddClient(userId string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[userId] = conn
	m.broadcast <- UserStatus{UserID: userId, Status: "online"}
}

func (m *OnlineManager) RemoveClient(userId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[userId]; ok {
		delete(m.clients, userId)
		m.broadcast <- UserStatus{UserID: userId, Status: "offline"}
	}
}

func (m *OnlineManager) BroadcastStatus() {
	for {
		status := <-m.broadcast
		for _, conn := range m.clients {
			if err := conn.WriteJSON(status); err != nil {
				log.Println("Error broadcasting status:", err)
			}
		}
	}
}
