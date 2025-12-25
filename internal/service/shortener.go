package service

import (
	"fmt"
	"time"

	"github.com/akhilthirunalveli/GoURL/internal/store"
)

type ShortenerService struct {
	store store.Store
}

func NewShortenerService(s store.Store) *ShortenerService {
	return &ShortenerService{store: s}
}

func (s *ShortenerService) Shorten(originalURL string) (*store.Link, error) {
	// 1. Get Next ID from Store
	id, err := s.store.NextID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	// 2. Encode ID to Base62
	code := Encode(id)

	link := &store.Link{
		ID:          id,
		OriginalURL: originalURL,
		ShortCode:   code,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * 7 * time.Hour), // Default 1 week
	}

	// 3. Save Link with explicit ID
	if err := s.store.SaveLink(link); err != nil {
		return nil, fmt.Errorf("failed to save link: %w", err)
	}

	return link, nil
}

func (s *ShortenerService) GetOriginalURL(code string) (string, error) {
	link, err := s.store.GetLinkByCode(code)
	if err != nil {
		return "", err
	}
	return link.OriginalURL, nil
}
