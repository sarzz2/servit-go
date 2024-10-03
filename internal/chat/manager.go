// servit-go/internal/chat/manager.go
package chat

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// ConnectionManager manages WebSocket connections for users
type ConnectionManager struct {
	// Map of user ID to WebSocket connections
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

// NewConnectionManager creates a new ConnectionManager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*websocket.Conn),
	}
}

// AddConnection adds a new WebSocket connection for a user
func (m *ConnectionManager) AddConnection(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[userID] = conn
	log.Printf("User %s connected", userID)
}

// RemoveConnection removes a WebSocket connection for a user
func (m *ConnectionManager) RemoveConnection(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, userID)
	log.Printf("User %s disconnected", userID)
}

// GetConnection retrieves the WebSocket connection for a user
func (m *ConnectionManager) GetConnection(userID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[userID]
	return conn, exists
}
