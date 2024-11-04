package routes

import (
	"servit-go/internal/db"
	"servit-go/internal/handlers"
	"servit-go/internal/manager"
	"servit-go/internal/middleware"
	"servit-go/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	chatService := services.NewChatService(db.DB)
	channelChatManager := manager.NewChannelChatManager()
	channelChatService := services.NewChannelChatService(db.DB, channelChatManager)
	onlineService := services.NewOnlineService()

	// Set up routes
	router.GET("/", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.ChatHandler(c.Writer, c.Request, chatService)
	})
	router.GET("/fetch_paginated_messages", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.FetchPaginatedMessagesHandler(c.Writer, c.Request, chatService)
	})
	router.GET("/fetch_chat_history", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.FetchUserChatHistory(c.Writer, c.Request, chatService)
	})
	router.GET("/ws/online", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.OnlineHandler(c.Writer, c.Request, onlineService)
	})

	router.GET("/friends/online", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.GetFriendsOnlineStatus(c.Writer, c.Request, onlineService)
	})

	// Channel chat
	router.GET("/ws/channel", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.ChannelChatHandler(c.Writer, c.Request, channelChatService)
	})

	router.GET("/fetch_channel_messages", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.FetchPaginatedChannelMessages(c.Writer, c.Request, channelChatService)
	})
}
