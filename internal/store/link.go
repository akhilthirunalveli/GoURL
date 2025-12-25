package store

import "time"

type Link struct {
	ID          uint64     `json:"id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"` // Nil if no expiry
}

// Store defines the interface for link storage
type Store interface {
	SaveLink(link *Link) error
	GetLinkByCode(code string) (*Link, error)
	Close()
}
