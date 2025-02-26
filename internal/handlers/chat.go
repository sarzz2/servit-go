package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"servit-go/internal/middleware"
	"servit-go/internal/models"
	"servit-go/internal/services"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// FetchPaginatedMessagesHandler retrieves messages between two users using paging state.
func FetchPaginatedMessagesHandler(w http.ResponseWriter, r *http.Request, chatService services.ChatServiceInterface) {
	toUserID := r.URL.Query().Get("to_user_id")
	pagingStateStr := r.URL.Query().Get("paging_state")
	var pagingState []byte
	if pagingStateStr != "" {
		// Decode the paging state from its base64 string representation.
		var err error
		pagingState, err = base64.StdEncoding.DecodeString(pagingStateStr)
		if err != nil {
			log.Print(err)
			http.Error(w, "Invalid paging_state", http.StatusBadRequest)
			return
		}
	}

	// Optionally, allow the frontend to specify a page size; default to 50.
	pageSize := 10
	if psStr := r.URL.Query().Get("page_size"); psStr != "" {
		if ps, err := strconv.Atoi(psStr); err == nil {
			pageSize = ps
		}
	}

	// Get fromUserID from context (authenticated user)
	fromUserID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || fromUserID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Fetch paginated messages using conversation-based query.
	messages, newPagingState, err := chatService.QueryMessages(fromUserID, toUserID, pageSize, pagingState)
	if err != nil {
		log.Print(err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	// Build response struct containing messages and the new paging state.
	response := struct {
		Messages    []models.DMMessage `json:"messages"`
		PagingState []byte             `json:"paging_state"`
	}{
		Messages:    messages,
		PagingState: newPagingState, // Will be encoded as base64 in JSON
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
		return
	}
}

func FetchPaginatedChannelMessagesHandler(w http.ResponseWriter, r *http.Request, chatService services.ChatServiceInterface) {
	// Retrieve required query parameter: channel_id.
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "Missing channel_id", http.StatusBadRequest)
		return
	}

	// Retrieve optional paging_state parameter and decode it from base64.
	pagingStateStr := r.URL.Query().Get("paging_state")
	var pagingState []byte
	if pagingStateStr != "" {
		var err error
		pagingState, err = base64.StdEncoding.DecodeString(pagingStateStr)
		if err != nil {
			log.Print(err)
			http.Error(w, "Invalid paging_state", http.StatusBadRequest)
			return
		}
	}

	// Optionally allow the frontend to specify a page size; default to 10.
	pageSize := 10
	if psStr := r.URL.Query().Get("page_size"); psStr != "" {
		if ps, err := strconv.Atoi(psStr); err == nil {
			pageSize = ps
		}
	}

	// Fetch paginated channel messages.
	messages, newPagingState, err := chatService.QueryChannelMessages(channelID, pageSize, pagingState)
	if err != nil {
		log.Print(err)
		http.Error(w, fmt.Sprintf("Failed to fetch channel messages: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response struct containing messages and the new paging state.
	response := struct {
		Messages    []models.ChannelMessage `json:"messages"`
		PagingState []byte                  `json:"paging_state"`
	}{
		Messages:    messages,
		PagingState: newPagingState,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode messages", http.StatusInternalServerError)
		return
	}
}
