package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/cache"
	"github.com/akhilthirunalveli/GoURL/internal/config"
	"github.com/akhilthirunalveli/GoURL/internal/database"
	"github.com/akhilthirunalveli/GoURL/internal/handler"
	"github.com/akhilthirunalveli/GoURL/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL database")

	// Initialize cache
	redisCache, err := cache.New(
		cfg.Redis.Address(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.App.CacheTTL,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()
	log.Println("Connected to Redis cache")

	// Initialize service
	urlService := service.NewURLService(
		db,
		redisCache,
		cfg.App.ShortCodeLength,
		cfg.App.BaseURL,
		cfg.App.RateLimitRequests,
		cfg.App.RateLimitWindow,
	)

	// Initialize handler
	urlHandler := handler.NewURLHandler(urlService)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shorten", urlHandler.CreateShortURL)
	mux.HandleFunc("/health", urlHandler.HealthCheck)
	mux.HandleFunc("/", urlHandler.RedirectURL)

	// Create server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
