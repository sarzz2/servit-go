package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"servit-go/internal/config"
	"servit-go/internal/db"
	"servit-go/internal/routes"
)

func main() {
	// Load environment variables and configuration
	cfg := config.LoadConfig()

	// Initialize the database connection
	db.InitDB(cfg.DatabaseURL)

	// Initialize Gin router
	router := gin.Default()

	// Set up routes
	routes.SetupRoutes(router)

	// Start the server
	fmt.Printf("Starting server on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		fmt.Printf("Could not start server: %v", err)
	}
}
