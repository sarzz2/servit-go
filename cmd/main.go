package main

import (
	"fmt"
	"servit-go/internal/config"
	"servit-go/internal/db"
	"servit-go/internal/routes"
	"servit-go/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables and configuration
	cfg := config.LoadConfig()
	db.InitDB(cfg.DatabaseURL)

	// Initialize Gin router
	router := gin.Default()

	// Initialize ChatService
	chatService := services.NewChatService(db.DB)

	// Set up routes
	routes.SetupRoutes(router, chatService)

	fmt.Printf("Starting server on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		fmt.Printf("Could not start server: %v", err)
	}
}
