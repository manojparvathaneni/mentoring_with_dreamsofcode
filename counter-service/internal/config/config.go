package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Version is the application version
const Version = "1.0.0"

// Constants for default configuration
const (
	defaultPort           = "8090"
	defaultFilename       = "counter.json"
	defaultShutdownTimeout = 10 * time.Second
	defaultReadTimeout     = 5 * time.Second
	defaultWriteTimeout    = 10 * time.Second
	defaultIdleTimeout     = 120 * time.Second
	defaultFilePermissions = 0644
	defaultSaveRetryAttempts = 3
	defaultSaveRetryDelay    = 100 * time.Millisecond
	defaultRateLimit         = 10
	defaultRateBurst         = 20
	defaultPersistInterval   = 5 * time.Minute
	defaultLogLevel          = "info"
	defaultEnvironment       = "development"
)

// Config holds application configuration
type Config struct {
	// Server settings
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// File persistence settings
	Filename          string
	FilePermissions   os.FileMode
	SaveRetryAttempts int
	SaveRetryDelay    time.Duration
	PersistInterval   time.Duration

	// Rate limiting
	RateLimit int
	RateBurst int

	// Feature flags
	EnableMetrics bool
	EnableCORS    bool

	// CORS settings
	AllowedOrigins []string

	// Logging
	LogLevel    string
	Environment string
}

// Load loads the application configuration
func Load() (*Config, error) {
	// Set up default configuration
	viper.SetDefault("port", defaultPort)
	viper.SetDefault("readTimeout", defaultReadTimeout)
	viper.SetDefault("writeTimeout", defaultWriteTimeout)
	viper.SetDefault("idleTimeout", defaultIdleTimeout)
	viper.SetDefault("shutdownTimeout", defaultShutdownTimeout)
	viper.SetDefault("filename", defaultFilename)
	viper.SetDefault("filePermissions", defaultFilePermissions)
	viper.SetDefault("saveRetryAttempts", defaultSaveRetryAttempts)
	viper.SetDefault("saveRetryDelay", defaultSaveRetryDelay)
	viper.SetDefault("persistInterval", defaultPersistInterval)
	viper.SetDefault("rateLimit", defaultRateLimit)
	viper.SetDefault("rateBurst", defaultRateBurst)
	viper.SetDefault("enableMetrics", true)
	viper.SetDefault("enableCORS", true)
	viper.SetDefault("allowedOrigins", []string{"*"})
	viper.SetDefault("logLevel", defaultLogLevel)
	viper.SetDefault("environment", defaultEnvironment)

	// Set up configuration file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/counter/")

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("COUNTER")

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load configuration into struct
	config := &Config{
		Port:              viper.GetString("port"),
		ReadTimeout:       viper.GetDuration("readTimeout"),
		WriteTimeout:      viper.GetDuration("writeTimeout"),
		IdleTimeout:       viper.GetDuration("idleTimeout"),
		ShutdownTimeout:   viper.GetDuration("shutdownTimeout"),
		Filename:          viper.GetString("filename"),
		FilePermissions:   os.FileMode(viper.GetInt("filePermissions")),
		SaveRetryAttempts: viper.GetInt("saveRetryAttempts"),
		SaveRetryDelay:    viper.GetDuration("saveRetryDelay"),
		PersistInterval:   viper.GetDuration("persistInterval"),
		RateLimit:         viper.GetInt("rateLimit"),
		RateBurst:         viper.GetInt("rateBurst"),
		EnableMetrics:     viper.GetBool("enableMetrics"),
		EnableCORS:        viper.GetBool("enableCORS"),
		AllowedOrigins:    viper.GetStringSlice("allowedOrigins"),
		LogLevel:          viper.GetString("logLevel"),
		Environment:       viper.GetString("environment"),
	}

	return config, nil
}