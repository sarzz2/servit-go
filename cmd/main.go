package main

import (
	"fmt"
	"servit-go/internal/config"
	"servit-go/internal/db"
	"servit-go/internal/routes"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables and configuration
	cfg := config.LoadConfig()
	db.InitDB(cfg.DatabaseURL)

	// Initialize Gin router
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	routes.SetupRoutes(router)

	fmt.Printf("Starting server on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		fmt.Printf("Could not start server: %v", err)
	}
}
