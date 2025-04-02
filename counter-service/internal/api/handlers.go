package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/counter"
)

// HTTPResponse standardizes API responses
type HTTPResponse struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	Error        string      `json:"error,omitempty"`
	ErrorCode    string      `json:"error_code,omitempty"`
	RequestID    string      `json:"request_id,omitempty"`
	ResponseTime float64     `json:"response_time_ms,omitempty"`
}

// Handler contains the HTTP handlers for the API
type Handler struct {
	counterService *counter.Service
	logger         *zerolog.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(counterService *counter.Service, logger *zerolog.Logger) *Handler {
	return &Handler{
		counterService: counterService,
		logger:         logger,
	}
}

// HealthCheck handles the health check endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := r.Context().Value(requestIDKey).(string)

	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, r, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED", requestID, start)
		return
	}

	// Basic health check
	health := map[string]interface{}{
		"status":    "UP",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   config.Version,
		"buildInfo": map[string]string{
			"goVersion": runtime.Version(),
			"platform":  runtime.GOOS + "/" + runtime.GOARCH,
		},
	}

	h.sendJSONResponse(w, http.StatusOK, HTTPResponse{
		Success:      true,
		Data:         health,
		RequestID:    requestID,
		ResponseTime: float64(time.Since(start).Microseconds()) / 1000.0,
	})
}

// IncrementCounter handles the counter increment endpoint
func (h *Handler) IncrementCounter(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := r.Context().Value(requestIDKey).(string)

	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, r, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED", requestID, start)
		return
	}

	// Increment counter
	newValue, err := h.counterService.Increment()
	if err != nil {
		h.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to increment counter", "COUNTER_ERROR", requestID, start)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, HTTPResponse{
		Success: true,
		Data: map[string]interface{}{
			"visits": newValue,
		},
		RequestID:    requestID,
		ResponseTime: float64(time.Since(start).Microseconds()) / 1000.0,
	})
}

// GetCounter handles the counter get endpoint
func (h *Handler) GetCounter(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := r.Context().Value(requestIDKey).(string)

	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, r, http.StatusMethodNotAllowed, "Method not allowed", "METHOD_NOT_ALLOWED", requestID, start)
		return
	}

	// Get counter value
	value, err := h.counterService.GetValue()
	if err != nil {
		h.sendErrorResponse(w, r, http.StatusInternalServerError, "Failed to get counter", "COUNTER_ERROR", requestID, start)
		return
	}

	h.sendJSONResponse(w, http.StatusOK, HTTPResponse{
		Success: true,
		Data: map[string]interface{}{
			"visits": value,
		},
		RequestID:    requestID,
		ResponseTime: float64(time.Since(start).Microseconds()) / 1000.0,
	})
}

// sendJSONResponse sends a JSON response with the provided status code
func (h *Handler) sendJSONResponse(w http.ResponseWriter, statusCode int, response HTTPResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse sends an error response with the provided status code
func (h *Handler) sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string, errorCode string, requestID string, start time.Time) {
	h.logger.Error().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("error", message).
		Str("errorCode", errorCode).
		Str("requestID", requestID).
		Int("status", statusCode).
		Msg("Request error")

	h.sendJSONResponse(w, statusCode, HTTPResponse{
		Success:      false,
		Error:        message,
		ErrorCode:    errorCode,
		RequestID:    requestID,
		ResponseTime: float64(time.Since(start).Microseconds()) / 1000.0,
	})
}