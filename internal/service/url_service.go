package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/cache"
	"github.com/akhilthirunalveli/GoURL/internal/database"
	"github.com/akhilthirunalveli/GoURL/internal/models"
)

const (
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxRetries        = 10
	rateLimitExceeded = "rate limit exceeded"
)

type URLService struct {
	db              *database.Database
	cache           *cache.Cache
	shortCodeLength int
	baseURL         string
	rateLimitReqs   int
	rateLimitWindow int
	mu              sync.RWMutex
}

func NewURLService(db *database.Database, cache *cache.Cache, shortCodeLength int, baseURL string, rateLimitReqs, rateLimitWindow int) *URLService {
	return &URLService{
		db:              db,
		cache:           cache,
		shortCodeLength: shortCodeLength,
		baseURL:         baseURL,
		rateLimitReqs:   rateLimitReqs,
		rateLimitWindow: rateLimitWindow,
	}
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL, customCode, clientIP string) (*models.CreateURLResponse, error) {
	// Check rate limit
	allowed, err := s.cache.CheckRateLimit(ctx, clientIP, s.rateLimitReqs, s.rateLimitWindow)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}
	if !allowed {
		return nil, fmt.Errorf(rateLimitExceeded)
	}

	// Validate URL
	if !s.isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	var shortCode string
	if customCode != "" {
		// Validate custom code
		if !s.isValidShortCode(customCode) {
			return nil, fmt.Errorf("invalid custom code: must be alphanumeric and 3-20 characters")
		}

		// Check if custom code already exists
		exists, err := s.db.ShortCodeExists(ctx, customCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom code: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("custom code already exists")
		}
		shortCode = customCode
	} else {
		// Generate unique short code
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// Create URL entry
	urlEntry := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}

	if err := s.db.CreateURL(ctx, urlEntry); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	// Cache the URL
	if err := s.cache.Set(ctx, shortCode, urlEntry); err != nil {
		// Log error but don't fail the request
		log.Printf("Warning: failed to cache URL: %v", err)
	}

	return &models.CreateURLResponse{
		ShortCode: shortCode,
		ShortURL:  fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		LongURL:   originalURL,
	}, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string, clientIP string) (string, error) {
	// Check rate limit
	allowed, err := s.cache.CheckRateLimit(ctx, clientIP, s.rateLimitReqs, s.rateLimitWindow)
	if err != nil {
		return "", fmt.Errorf("rate limit check failed: %w", err)
	}
	if !allowed {
		return "", fmt.Errorf(rateLimitExceeded)
	}

	// Try to get from cache first
	cachedURL, err := s.cache.Get(ctx, shortCode)
	if err != nil {
		// Log error but continue to database
		log.Printf("Warning: cache get failed: %v", err)
	}

	if cachedURL != nil {
		// Increment click count asynchronously
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.db.IncrementClickCount(ctx, shortCode); err != nil {
				log.Printf("Warning: failed to increment click count: %v", err)
			}
		}()
		return cachedURL.OriginalURL, nil
	}

	// Get from database
	urlEntry, err := s.db.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	if urlEntry == nil {
		return "", nil
	}

	// Cache the result
	if err := s.cache.Set(ctx, shortCode, urlEntry); err != nil {
		// Log error but don't fail the request
		log.Printf("Warning: failed to cache URL: %v", err)
	}

	// Increment click count asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.db.IncrementClickCount(ctx, shortCode); err != nil {
			log.Printf("Warning: failed to increment click count: %v", err)
		}
	}()

	return urlEntry.OriginalURL, nil
}

func (s *URLService) generateUniqueShortCode(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < maxRetries; i++ {
		shortCode, err := s.generateRandomString(s.shortCodeLength)
		if err != nil {
			return "", err
		}

		exists, err := s.db.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d retries", maxRetries)
}

func (s *URLService) generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func (s *URLService) isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Must have a scheme (http/https) and host
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// Only allow http and https
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	return true
}

func (s *URLService) isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	for _, c := range code {
		if !strings.ContainsRune(charset, c) {
			return false
		}
	}

	return true
}

func (s *URLService) IsRateLimitError(err error) bool {
	return err != nil && err.Error() == rateLimitExceeded
}
