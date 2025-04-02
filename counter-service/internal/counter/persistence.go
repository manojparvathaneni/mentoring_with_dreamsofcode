package counter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/metrics"
	"github.com/yourusername/counter-service/pkg/fileutils"
)

// CounterData is the structure used for serialization
type CounterData struct {
	Visits    int64     `json:"visits"`
	Timestamp time.Time `json:"last_updated"`
	Version   string    `json:"version"`
	CRC       uint32    `json:"crc,omitempty"`
}

// SaveCounter persists the counter to disk
func SaveCounter(counter *Counter, cfg *config.Config, logger *zerolog.Logger, metrics *metrics.Metrics) error {
	startTime := time.Now()
	defer func() {
		metrics.OperationDuration.WithLabelValues("save").Observe(time.Since(startTime).Seconds())
	}()
	
	// Increment operation counter
	metrics.CounterOperations.WithLabelValues("save").Inc()
	
	// Prepare data
	data := CounterData{
		Visits:    counter.GetValue(),
		Timestamp: time.Now(),
		Version:   config.Version,
	}
	
	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal counter data")
		metrics.PersistErrors.Inc()
		return err
	}
	
	// Calculate CRC
	crc := fileutils.CalculateCRC(jsonBytes)
	data.CRC = crc
	
	// Marshal again with CRC
	jsonBytes, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal counter data with CRC")
		metrics.PersistErrors.Inc()
		return err
	}
	
	// Implement retry logic
	var saveErr error
	for attempt := 0; attempt < cfg.SaveRetryAttempts; attempt++ {
		saveErr = writeCounterToDisk(jsonBytes, cfg, logger, metrics)
		if saveErr == nil {
			// Successfully saved, mark counter as clean
			counter.MarkClean()
			return nil
		}
		
		logger.Warn().
			Err(saveErr).
			Int("attempt", attempt+1).
			Int("maxAttempts", cfg.SaveRetryAttempts).
			Msg("Save attempt failed, retrying")
			
		metrics.PersistErrors.Inc()
		time.Sleep(cfg.SaveRetryDelay)
	}
	
	logger.Error().
		Err(saveErr).
		Int("attempts", cfg.SaveRetryAttempts).
		Msg("Failed to save counter after multiple attempts")
	
	return fmt.Errorf("failed to save counter after %d attempts: %w", cfg.SaveRetryAttempts, saveErr)
}

// writeCounterToDisk handles atomic file writing with proper locking
func writeCounterToDisk(data []byte, cfg *config.Config, logger *zerolog.Logger, metrics *metrics.Metrics) error {
	startTime := time.Now()
	defer func() {
		metrics.OperationDuration.WithLabelValues("write").Observe(time.Since(startTime).Seconds())
	}()
	
	metrics.CounterOperations.WithLabelValues("write").Inc()
	
	// Create temporary file for atomic writing
	tempFile := cfg.Filename + ".tmp"
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, cfg.FilePermissions)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	
	defer func() {
		f.Close()
		// Clean up temp file on error
		if err != nil {
			os.Remove(tempFile)
		}
	}()
	
	// Apply exclusive lock for writing
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire write lock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	
	// Write data
	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	
	// Ensure data is written to disk
	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	
	// Close file explicitly before rename
	f.Close()
	
	// Atomically replace the old file with the new one
	if err := os.Rename(tempFile, cfg.Filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

// LoadCounter reads the counter from disk
func LoadCounter(cfg *config.Config, logger *zerolog.Logger, metrics *metrics.Metrics) (*Counter, error) {
	startTime := time.Now()
	defer func() {
		metrics.OperationDuration.WithLabelValues("load").Observe(time.Since(startTime).Seconds())
	}()
	
	metrics.CounterOperations.WithLabelValues("load").Inc()
	
	// Check if file exists
	if _, err := os.Stat(cfg.Filename); os.IsNotExist(err) {
		logger.Info().Msg("Counter file does not exist, starting with zero")
		return NewCounter(0), nil
	}
	
	f, err := os.OpenFile(cfg.Filename, os.O_RDONLY, cfg.FilePermissions)
	if err != nil {
		return nil, fmt.Errorf("failed to open counter file: %w", err)
	}
	defer f.Close()
	
	// Apply shared lock for reading
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return nil, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	
	// Check if file is empty
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	if fi.Size() == 0 {
		logger.Info().Msg("Empty counter file, starting with zero")
		return NewCounter(0), nil
	}
	
	// Read file content
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}
	
	var data CounterData
	if err := json.Unmarshal(content, &data); err != nil {
		logger.Warn().Err(err).Msg("Failed to decode counter data, starting with zero")
		return NewCounter(0), nil
	}
	
	// Validate CRC if present
	if data.CRC > 0 {
		// Create a copy without CRC for validation
		dataCopy := data
		dataCopy.CRC = 0
		jsonBytes, err := json.MarshalIndent(dataCopy, "", "  ")
		if err == nil {
			calculatedCRC := fileutils.CalculateCRC(jsonBytes)
			if calculatedCRC != data.CRC {
				logger.Warn().
					Uint32("expected", data.CRC).
					Uint32("calculated", calculatedCRC).
					Msg("CRC validation failed, starting with zero")
				return NewCounter(0), nil
			}
		}
	}
	
	logger.Info().Int64("visits", data.Visits).Msg("Counter loaded successfully")
	return NewCounter(data.Visits), nil
}