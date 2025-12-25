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
// It skips if DB_HOST is not set or if connection fails (optional, but good for CI/CD separation)
func TestPostgresIntegration(t *testing.T) {
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
		t.Logf("Skipping integration test: %v", err)
		return
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

	// Test Save with ExpiresAt
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

	// Test Save with NULL ExpiresAt
	linkNoExpiry := &store.Link{
		OriginalURL: "https://example.com",
		ShortCode:   "test2",
		CreatedAt:   time.Now(),
		ExpiresAt:   nil, // NULL expiry
	}

	if err := s.SaveLink(linkNoExpiry); err != nil {
		t.Fatalf("Failed to save link with NULL ExpiresAt: %v", err)
	}

	if linkNoExpiry.ID == 0 {
		t.Fatal("Expected ID to be set after save")
	}

	// Test Get for link with NULL ExpiresAt
	retrievedNoExpiry, err := s.GetLinkByCode("test2")
	if err != nil {
		t.Fatalf("Failed to get link with NULL ExpiresAt: %v", err)
	}

	if retrievedNoExpiry.OriginalURL != linkNoExpiry.OriginalURL {
		t.Errorf("Expected URL %s, got %s", linkNoExpiry.OriginalURL, retrievedNoExpiry.OriginalURL)
	}

	if retrievedNoExpiry.ExpiresAt != nil {
		t.Errorf("Expected ExpiresAt to be nil, got %v", retrievedNoExpiry.ExpiresAt)
	}

	// Clean up test data even if the test fails, and do not ignore errors.
	t.Cleanup(func() {
		if _, err := s.Pool().Exec(context.Background(), "DELETE FROM links WHERE short_code IN ('test1', 'test2')"); err != nil {
			t.Fatalf("failed to clean up test data: %v", err)
		}
	})
}
