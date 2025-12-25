package main

import (
	"fmt"

	"github.com/akhilthirunalveli/GoURL/internal/config"
	"github.com/akhilthirunalveli/GoURL/pkg/logger"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

	// Initialize Logger
	logger.InitLogger(cfg.Server.Env)

	logger.Log.Info("Starting GoURL Service",
		"env", cfg.Server.Env,
		"port", cfg.Server.Port,
	)

	// Placeholder for server start
	fmt.Printf("Server is configured to run on port %s\n", cfg.Server.Port)
	fmt.Printf("Database Host: %s\n", cfg.DB.Host)
	fmt.Printf("Redis Addr: %s\n", cfg.Redis.Addr)
}
