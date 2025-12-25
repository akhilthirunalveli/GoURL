package models

import "time"

type URL struct {
	ID          int       `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ClickCount  int       `json:"click_count" db:"click_count"`
}

type CreateURLRequest struct {
	URL        string `json:"url" binding:"required"`
	CustomCode string `json:"custom_code,omitempty"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	LongURL   string `json:"long_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
