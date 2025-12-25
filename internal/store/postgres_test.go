package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/config"
	"github.com/akhilthirunalveli/GoURL/internal/store"
	"github.com/akhilthirunalveli/GoURL/pkg/logger"
)

// TestPostgresIntegration runs a real database test.
// It skips if INTEGRATION_TEST is not set to "true" or if the database connection fails.
func TestPostgresIntegration(t *testing.T) {
	// Skip if INTEGRATION_TEST environment variable is not set to "true"
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TEST environment variable not set to 'true'")
	}

	// Initialize logger for tests
	logger.InitLogger("development")

	// Load config purely from env or use defaults.
	// To run this locally, ensure you have set env vars or rely on the defaults which point to localhost
	cfg := config.DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password", // Change logic to read from env if needed for local manual run
		Name:     "gourl",
		SSLMode:  "disable",
	}

	// Allow overriding via env vars
	if h := os.Getenv("DB_HOST"); h != "" {
		cfg.Host = h
	}
	if p := os.Getenv("DB_PASSWORD"); p != "" {
		cfg.Password = p
	}

	s, err := store.NewPostgresStore(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: database connection failed: %v", err)
	}
	defer s.Close()

	// Initial Migration (Quick & Dirty for Test)
	_, _ = s.Pool().Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS links (
			id BIGSERIAL PRIMARY KEY,
			original_url TEXT NOT NULL,
			short_code VARCHAR(10) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP WITH TIME ZONE
		);
	`)

	// Test Save
	expiresAt := time.Now().Add(24 * time.Hour)
	link := &store.Link{
		OriginalURL: "https://google.com",
		ShortCode:   "test1",
		CreatedAt:   time.Now(),
		ExpiresAt:   &expiresAt,
	}

	if err := s.SaveLink(link); err != nil {
		t.Fatalf("Failed to save link: %v", err)
	}

	if link.ID == 0 {
		t.Fatal("Expected ID to be set after save")
	}

	// Test Get
	retrieved, err := s.GetLinkByCode("test1")
	if err != nil {
		t.Fatalf("Failed to get link: %v", err)
	}

	if retrieved.OriginalURL != link.OriginalURL {
		t.Errorf("Expected URL %s, got %s", link.OriginalURL, retrieved.OriginalURL)
	}

	// Clean up
	_, _ = s.Pool().Exec(context.Background(), "DELETE FROM links WHERE short_code = 'test1'")
}
