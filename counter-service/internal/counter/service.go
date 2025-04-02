package counter

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/metrics"
)

// Service handles business logic for the counter
type Service struct {
	counter        *Counter
	config         *config.Config
	logger         *zerolog.Logger
	metrics        *metrics.Metrics
	persistMu      sync.Mutex
	shutdownCh     chan struct{}
	backgroundDone chan struct{}
}

// NewService creates a new counter service
func NewService(cfg *config.Config, logger *zerolog.Logger, metrics *metrics.Metrics) (*Service, error) {
	// Load counter from disk
	counter, err := LoadCounter(cfg, logger, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to load counter: %w", err)
	}

	// Update metric for current counter value
	metrics.CounterValue.Set(float64(counter.GetValue()))

	// Create service
	service := &Service{
		counter:        counter,
		config:         cfg,
		logger:         logger,
		metrics:        metrics,
		shutdownCh:     make(chan struct{}),
		backgroundDone: make(chan struct{}),
	}

	// Start background persistence
	go service.backgroundPersistence()

	return service, nil
}

// Increment increments the counter and returns the new value
func (s *Service) Increment() (int64, error) {
	// Increment counter
	newValue := s.counter.Increment()

	// Update metric
	s.metrics.CounterValue.Set(float64(newValue))
	s.metrics.CounterOperations.WithLabelValues("increment").Inc()

	return newValue, nil
}

// GetValue returns the current counter value
func (s *Service) GetValue() (int64, error) {
	value := s.counter.GetValue()
	s.metrics.CounterOperations.WithLabelValues("get").Inc()
	return value, nil
}

// Persist forces the counter to be persisted to disk
func (s *Service) Persist() error {
	s.persistMu.Lock()
	defer s.persistMu.Unlock()

	// Only persist if counter is dirty
	if !s.counter.IsDirty() {
		return nil
	}

	s.logger.Debug().Msg("Persisting counter to disk")
	return SaveCounter(s.counter, s.config, s.logger, s.metrics)
}

// backgroundPersistence periodically saves the counter to disk
func (s *Service) backgroundPersistence() {
	ticker := time.NewTicker(s.config.PersistInterval)
	defer ticker.Stop()
	defer close(s.backgroundDone)

	s.logger.Debug().Dur("interval", s.config.PersistInterval).Msg("Starting background persistence")

	for {
		select {
		case <-ticker.C:
			if s.counter.IsDirty() {
				s.logger.Debug().Msg("Performing scheduled counter persistence")
				s.persistMu.Lock()
				if err := SaveCounter(s.counter, s.config, s.logger, s.metrics); err != nil {
					s.logger.Error().Err(err).Msg("Failed to persist counter in background")
				}
				s.persistMu.Unlock()
			}
		case <-s.shutdownCh:
			s.logger.Debug().Msg("Background persistence stopping")
			return
		}
	}
}

// Shutdown stops the background persistence
func (s *Service) Shutdown() error {
	close(s.shutdownCh)
	<-s.backgroundDone
	return s.Persist()
}