package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // <-- 1. ADD THIS IMPORT

	"github.com/sujalbistaa/whispr/internal/db"
	routes "github.com/sujalbistaa/whispr/internal/http"
	"github.com/sujalbistaa/whispr/internal/models"
	"github.com/sujalbistaa/whispr/internal/ws"
)

func main() {
	// 2. LOAD .env FILE
	// This MUST be the first thing we do.
	if err := godotenv.Load(); err != nil {
		// We don't panic, but we log it. This allows running in production
		// (where env vars are set directly) without a .env file.
		log.Println("No .env file found, reading from environment")
	}

	// 1. Initialize Database
	database, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Run Migrations
	log.Println("Running database migrations...")
	if err := database.AutoMigrate(&models.Post{}, &models.Vote{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations complete.")

	// 3. Initialize WebSocket Hub
	hub := ws.NewHub()
	go hub.Run() // Run the hub in a separate goroutine

	// 4. Initialize Gin Router
	router := gin.Default()

	// 5. Setup Routes
	// This is where the panic was happening. Now it will find the env var.
	routes.SetupRoutes(router, database, hub)

	// 6. Start Server with Graceful Shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to start the server
	go func() {
		log.Printf("Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Block until a signal is received
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}