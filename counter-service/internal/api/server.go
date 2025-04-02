package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/counter"
	"github.com/yourusername/counter-service/internal/metrics"
	"golang.org/x/time/rate"
)

// Server represents the HTTP server
type Server struct {
	config         *config.Config
	logger         *zerolog.Logger
	counterService *counter.Service
	metrics        *metrics.Metrics
	server         *http.Server
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, logger *zerolog.Logger, counterService *counter.Service, metrics *metrics.Metrics) *Server {
	return &Server{
		config:         cfg,
		logger:         logger,
		counterService: counterService,
		metrics:        metrics,
	}
}

// setupRoutes configures the HTTP routes with middleware
func (s *Server) setupRoutes() http.Handler {
	// Create a new router
	mux := http.NewServeMux()

	// Create handler
	handler := NewHandler(s.counterService, s.logger)

	// Register API routes
	mux.HandleFunc("/api/counter/increment", handler.IncrementCounter)
	mux.HandleFunc("/api/counter", handler.GetCounter)
	mux.HandleFunc("/health", handler.HealthCheck)

	// Register metrics endpoint
	if s.config.EnableMetrics {
		mux.Handle("/metrics", promhttp.Handler())
	}

	// Apply middleware stack
	var middleware http.Handler = mux

	// Rate limiting
	limiter := rate.NewLimiter(rate.Limit(s.config.RateLimit), s.config.RateBurst)
	middleware = rateLimitMiddleware(s.logger, limiter)(middleware)

	// Metrics middleware
	middleware = metricsMiddleware(s.metrics)(middleware)

	// Request logging
	middleware = requestLogMiddleware(s.logger, s.metrics)(middleware)

	// Panic recovery
	middleware = recoverMiddleware(s.logger)(middleware)

	// CORS if enabled
	if s.config.EnableCORS {
		corsMiddleware := cors.New(cors.Options{
			AllowedOrigins:   s.config.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		middleware = corsMiddleware.Handler(middleware)
	}

	return middleware
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	// Create HTTP server
	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.setupRoutes(),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Start the server
	s.logger.Info().Str("port", s.config.Port).Msg("Server listening")
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	// Persist counter state before shutdown
	if err := s.counterService.Persist(); err != nil {
		s.logger.Error().Err(err).Msg("Error persisting counter during shutdown")
	}

	// Create a context with timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}