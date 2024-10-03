package routes

import (
	"servit-go/internal/handlers"
	"servit-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/", middleware.JWTAuthMiddleware(), func(c *gin.Context) {
		handlers.ChatHandler(c.Writer, c.Request)
	})
}
