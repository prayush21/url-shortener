package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prayushdave/url-shortener/internal/http"
	"github.com/prayushdave/url-shortener/internal/id"
	"github.com/prayushdave/url-shortener/internal/storage"
)

func main() {
	// Get configuration from environment variables
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0 // Using default DB
	serverPort := getEnv("SERVER_PORT", "8080")
	baseURL := getEnv("BASE_URL", fmt.Sprintf("http://localhost:%s", serverPort))

	// Initialize Redis store
	store := storage.NewRedisStore(redisAddr, redisPassword, redisDB)
	defer store.Close()

	// Initialize ID generator
	generator := id.NewGenerator()

	// Initialize HTTP handler
	handler := http.NewHandler(store, generator, baseURL)

	// Set up Gin router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // Vite's default dev server port
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	handler.SetupRoutes(router)

	// Start server
	log.Printf("Starting server on port %s...\n", serverPort)
	if err := router.Run(fmt.Sprintf(":%s", serverPort)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
