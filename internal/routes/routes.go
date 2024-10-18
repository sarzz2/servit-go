package routes

import (
	"servit-go/internal/handlers"
	"servit-go/internal/middleware"
	"servit-go/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, chatService *services.ChatService, onlineService *services.OnlineService) {
	router.GET("/", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.ChatHandler(c.Writer, c.Request, chatService)
	})
	router.GET("/fetch_paginated_messages", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.FetchPaginatedMessagesHandler(c.Writer, c.Request, chatService)
	})
	router.GET("/ws/online", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.OnlineHandler(c.Writer, c.Request, onlineService)
	})

	router.GET("/friends/online", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.GetFriendsOnlineStatus(c.Writer, c.Request, onlineService)
	})
}
