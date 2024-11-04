package manager

import (
	"servit-go/internal/models"
	"sync"
)

type ChannelChatManager struct {
	channels map[string]map[string]chan models.ChannelMessage
	mu       sync.RWMutex
}

func NewChannelChatManager() *ChannelChatManager {
	return &ChannelChatManager{
		channels: make(map[string]map[string]chan models.ChannelMessage),
	}
}

func (m *ChannelChatManager) JoinChannel(channelID, userID string) chan models.ChannelMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.channels[channelID]; !ok {
		m.channels[channelID] = make(map[string]chan models.ChannelMessage)
	}

	userChan := make(chan models.ChannelMessage)
	m.channels[channelID][userID] = userChan
	return userChan
}

func (m *ChannelChatManager) LeaveChannel(channelID, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if users, ok := m.channels[channelID]; ok {
		if userChan, ok := users[userID]; ok {
			close(userChan)
			delete(users, userID)
		}
		if len(users) == 0 {
			delete(m.channels, channelID)
		}
	}
}

func (m *ChannelChatManager) SendMessage(channelID, userID, username, message string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if users, ok := m.channels[channelID]; ok {
		msg := models.ChannelMessage{
			UserId:   userID,
			Username: username,
			Content:  message,
		}
		for id, userChan := range users {
			if id != userID {
				userChan <- msg
			}
		}
	}
}
