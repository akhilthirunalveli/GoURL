package rest

import (
	"encoding/json"
	"net/http"

	"github.com/akhilthirunalveli/GoURL/internal/service"
	"github.com/akhilthirunalveli/GoURL/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.ShortenerService
}

func NewHandler(svc *service.ShortenerService) *Handler {
	return &Handler{svc: svc}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	link, err := h.svc.Shorten(req.URL)
	if err != nil {
		logger.Log.Error("Failed to shorten URL", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.svc.GetOriginalURL(code)
	if err != nil {
		logger.Log.Error("Failed to get original URL", "error", err)
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
