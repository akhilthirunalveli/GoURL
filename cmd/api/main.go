package main

import (
	"net/http"
	"os"

	"github.com/akhilthirunalveli/GoURL/internal/config"
	"github.com/akhilthirunalveli/GoURL/internal/service"
	"github.com/akhilthirunalveli/GoURL/internal/store"
	"github.com/akhilthirunalveli/GoURL/internal/transport/rest"
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

	// Initialize Store
	dbStore, err := store.NewPostgresStore(cfg.DB)
	if err != nil {
		logger.Log.Error("Failed to initialize database store", "error", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	// Initialize Service
	svc := service.NewShortenerService(dbStore)

	// Initialize HTTP Handler & Router
	handler := rest.NewHandler(svc)
	router := rest.SetupRoutes(handler)

	// Start Server
	serverAddr := ":" + cfg.Server.Port
	logger.Log.Info("Server listening...", "address", serverAddr)
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Log.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
