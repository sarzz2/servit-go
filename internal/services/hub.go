package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"servit-go/internal/models"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients.
type Hub struct {
	Clients map[string]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[string]*Client),
	}
}

// Register adds a client to the Hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Clients[client.ID] = client
}

// Unregister removes a client from the Hub.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Clients, client.ID)

}

// GetClient returns a client by user ID.
func (h *Hub) GetClient(userID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.Clients[userID]
	return client, ok
}

// Client represents a connected user.
type Client struct {
	ID         string
	Username   string
	Conn       *websocket.Conn
	Send       chan []byte
	Hub        *Hub
	ActiveChat *models.ActiveChat // current active chat window
	Unread     map[string]int     // key: chat id, value: unread count
	mu         sync.Mutex         // protects ActiveChat and Unread
}

// ReadPump reads messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}
		switch wsMsg.Type {
		case "switch_chat":
			// Client is switching chats (channel or DM).
			var active models.ActiveChat
			if err := json.Unmarshal(wsMsg.Data, &active); err != nil {
				log.Println("Invalid switch_chat data:", err)
				continue
			}
			c.mu.Lock()
			c.ActiveChat = &active
			// Reset unread count for the newly active chat.
			c.Unread[active.ChatID] = 0
			c.mu.Unlock()
			log.Printf("User %s switched to %s chat: %s", c.Username, active.ChatType, active.ChatID)
		case "channel_message":
			// Process a channel message.
			var msg models.ChannelMessage
			if err := json.Unmarshal(wsMsg.Data, &msg); err != nil {
				log.Println("Invalid channel_message data:", err)
				continue
			}
			msg.Timestamp = time.Now()
			SaveChannelMessage(msg)
			BroadcastChannelMessage(msg, c.Hub)
		case "direct_message":
			// Process a direct message.
			var msg models.DMMessage
			if err := json.Unmarshal(wsMsg.Data, &msg); err != nil {
				log.Println("Invalid direct_message data:", err)
				continue
			}
			msg.Timestamp = time.Now()
			SaveDMMessage(msg)
			SendDirectMessage(msg, c.Hub)
		case "typing":
			// Process a typing indicator.
			var te models.TypingEvent
			if err := json.Unmarshal(wsMsg.Data, &te); err != nil {
				log.Println("Invalid typing data:", err)
				continue
			}
			if te.ChatType == "dm" {
				if target, ok := c.Hub.GetClient(te.ToUserID); ok {
					data, err := json.Marshal(te)
					if err != nil {
						log.Println("Error marshalling typing event:", err)
						continue
					}
					wrapped := models.WSMessage{
						Type: "typing",
						Data: data,
					}
					wsData, err := json.Marshal(wrapped)
					if err != nil {
						log.Println("Error wrapping typing event:", err)
						continue
					}
					target.Send <- wsData
				}
			} else if te.ChatType == "channel" {
				// Broadcast to all clients in the channel except the sender.
				c.Hub.mu.RLock()
				for _, client := range c.Hub.Clients {
					client.mu.Lock()
					if client.ActiveChat != nil &&
						client.ActiveChat.ChatType == "channel" &&
						client.ActiveChat.ChatID == te.ChatID &&
						client.ID != te.FromUserID {
						data, err := json.Marshal(te)
						if err != nil {
							client.mu.Unlock()
							continue
						}
						wrapped := models.WSMessage{
							Type: "typing",
							Data: data,
						}
						wsData, err := json.Marshal(wrapped)
						if err != nil {
							client.mu.Unlock()
							continue
						}
						client.Send <- wsData
					}
					client.mu.Unlock()
				}
				c.Hub.mu.RUnlock()
			}
		case "not_typing":
			// Process a "not_typing" indicator.
			var te models.TypingEvent
			if err := json.Unmarshal(wsMsg.Data, &te); err != nil {
				log.Println("Invalid not_typing data:", err)
				continue
			}
			if te.ChatType == "dm" {
				if target, ok := c.Hub.GetClient(te.ToUserID); ok {
					data, err := json.Marshal(te)
					if err != nil {
						log.Println("Error marshalling not_typing event:", err)
						continue
					}
					wrapped := models.WSMessage{
						Type: "not_typing",
						Data: data,
					}
					wsData, err := json.Marshal(wrapped)
					if err != nil {
						log.Println("Error wrapping not_typing event:", err)
						continue
					}
					target.Send <- wsData
				}
			} else if te.ChatType == "channel" {
				c.Hub.mu.RLock()
				for _, client := range c.Hub.Clients {
					client.mu.Lock()
					if client.ActiveChat != nil &&
						client.ActiveChat.ChatType == "channel" &&
						client.ActiveChat.ChatID == te.ChatID &&
						client.ID != te.FromUserID {
						data, err := json.Marshal(te)
						if err != nil {
							client.mu.Unlock()
							continue
						}
						wrapped := models.WSMessage{
							Type: "not_typing",
							Data: data,
						}
						wsData, err := json.Marshal(wrapped)
						if err != nil {
							client.mu.Unlock()
							continue
						}
						client.Send <- wsData
					}
					client.mu.Unlock()
				}
				c.Hub.mu.RUnlock()
			}
		default:
			log.Println("Unknown message type:", wsMsg.Type)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The channel is closed.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error for client %s: %v", c.ID, err)
				return
			} else {
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error for client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// BroadcastChannelMessage sends a channel message to all connected clients.
// If a client isn’t actively viewing that channel, it sends a notification with an unread count.
func BroadcastChannelMessage(msg models.ChannelMessage, hub *Hub) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling channel message:", err)
		return
	}

	// Create a WSMessage wrapper
	wsMsg := models.WSMessage{
		Type: "channel_message",
		Data: data,
	}

	// Marshal the entire WSMessage
	wrappedData, err := json.Marshal(wsMsg)
	if err != nil {
		log.Println("Error marshalling WSMessage:", err)
		return
	}

	clients := make([]*Client, 0, len(hub.Clients))
	for _, c := range hub.Clients {
		clients = append(clients, c)
	}

	for _, client := range clients {
		client.mu.Lock()

		if client.ActiveChat != nil &&
			client.ActiveChat.ChatType == "channel" &&
			client.ActiveChat.ChatID == msg.ChannelID &&
			client.ID != msg.SenderID {
			select {
			case client.Send <- wrappedData:
			default:
				log.Printf("Send channel full for client %s", client.ID)
			}
		} else if client.ID != msg.SenderID {
			// The client is not active in the channel—send a notification.
			client.Unread[msg.ChannelID]++
			notif := map[string]interface{}{
				"type":      "notification",
				"chat_type": "channel",
				"chat_id":   msg.ChannelID,
				"unread":    client.Unread[msg.ChannelID],
				"message":   "New message in channel " + msg.ChannelID,
			}
			notifData, _ := json.Marshal(notif)
			select {
			case client.Send <- notifData:
			default:
				log.Printf("Send channel full for client %s, skipping notification", client.ID)
			}
		}
		client.mu.Unlock()
	}
}

// SendDirectMessage delivers a direct message to the recipient.
func SendDirectMessage(msg models.DMMessage, hub *Hub) {
	hub.mu.RLock()
	receiver, ok := hub.Clients[msg.ReceiverID]
	hub.mu.RUnlock()

	// Marshal the DMMessage into the Data field of a WSMessage
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling direct message:", err)
		return
	}

	// Create a WSMessage wrapper
	wsMsg := models.WSMessage{
		Type: "direct_message",
		Data: msgData,
	}

	// Marshal the entire WSMessage
	wrappedData, err := json.Marshal(wsMsg)
	if err != nil {
		log.Println("Error marshalling WSMessage:", err)
		return
	}

	if ok {
		receiver.mu.Lock()
		defer receiver.mu.Unlock()
		log.Print("Active chat, reciever id", receiver.ActiveChat, msg.ReceiverID, msg.SenderID)
		if receiver.ActiveChat != nil &&
			receiver.ActiveChat.ChatType == "dm" &&
			receiver.ActiveChat.ChatID == msg.SenderID {
			// Send the wrapped WSMessage
			select {
			case receiver.Send <- wrappedData:
			default:
				log.Printf("Send channel full for client %s", receiver.ID)
			}
		} else {
			// Send notification (already correctly formatted)
			receiver.Unread[msg.SenderID]++
			notif := map[string]interface{}{
				"type":      "notification",
				"chat_type": "dm",
				"chat_id":   msg.SenderID,
				"unread":    receiver.Unread[msg.SenderID],
				"message":   "New direct message from " + msg.SenderID,
			}
			notifData, _ := json.Marshal(notif)
			select {
			case receiver.Send <- notifData:
			default:
				log.Printf("Send channel full for client %s", receiver.ID)
			}
		}
	} else {
		log.Printf("User %s not connected. Storing DM notification.", msg.ReceiverID)
	}
}
