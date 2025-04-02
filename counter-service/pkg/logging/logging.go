package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/yourusername/counter-service/pkg/fileutils"
)

// NewLogger creates a new zerolog logger with appropriate configuration
func NewLogger(logLevel string, environment string) *zerolog.Logger {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Add caller info to log
	zerolog.CallerMarshalFunc = shortenCallerPath

	// Create logger
	var logger zerolog.Logger

	// Set up pretty console logging for development
	if environment == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		// JSON logging for production
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	return &logger
}

// SetupFileLogging configures logging to a file in addition to stdout
func SetupFileLogging(logger *zerolog.Logger, logPath string) error {
	// Ensure log directory exists
	if err := fileutils.EnsureDirectory(logPath); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to log to both console and file
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)

	// Update logger to use multi-writer
	newLogger := zerolog.New(multi).With().Timestamp().Caller().Logger()
	*logger = newLogger

	return nil
}

// shortenCallerPath shortens the file path in logs
func shortenCallerPath(pc uintptr, file string, line int) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == filepath.Separator {
			short = file[i+1:]
			break
		}
	}
	return fmt.Sprintf("%s:%d", short, line)
}

// LoggerWithRequestID creates a logger with request ID context
func LoggerWithRequestID(logger *zerolog.Logger, requestID string) *zerolog.Logger {
	newLogger := logger.With().Str("requestID", requestID).Logger()
	return &newLogger
}

// RecoveryFn creates a function to recover from panics with logging
func RecoveryFn(logger *zerolog.Logger) func() {
	return func() {
		if r := recover(); r != nil {
			// Get stack trace
			buf := make([]byte, 8192)
			n := runtime.Stack(buf, false)
			stack := string(buf[:n])

			// Log panic
			logger.Error().
				Interface("panic", r).
				Str("stack", stack).
				Msg("Recovered from panic")
		}
	}
}