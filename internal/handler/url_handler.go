package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/akhilthirunalveli/GoURL/internal/models"
	"github.com/akhilthirunalveli/GoURL/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{
		service: service,
	}
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.respondError(w, "url is required", http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)
	resp, err := h.service.CreateShortURL(r.Context(), req.URL, req.CustomCode, clientIP)
	if err != nil {
		if h.service.IsRateLimitError(err) {
			h.respondError(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "already exists") {
			h.respondError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.respondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, resp, http.StatusCreated)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract short code from path
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		h.respondError(w, "short code is required", http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)
	originalURL, err := h.service.GetOriginalURL(r.Context(), shortCode, clientIP)
	if err != nil {
		if h.service.IsRateLimitError(err) {
			h.respondError(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		h.respondError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if originalURL == "" {
		h.respondError(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status": "healthy",
	}
	h.respondJSON(w, response, http.StatusOK)
}

func (h *URLHandler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *URLHandler) respondError(w http.ResponseWriter, message string, status int) {
	h.respondJSON(w, models.ErrorResponse{Error: message}, status)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
