package routes

import (
	"servit-go/internal/handlers"
	"servit-go/internal/middleware"
	"servit-go/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, chatService *services.ChatService) {
	router.GET("/", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.ChatHandler(c.Writer, c.Request, chatService)
	})
}
