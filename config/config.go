package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

// Environment variable names.
const (
	EnvLogLevel               = "LOG_LEVEL"
	EnvResponseTimeout        = "RESPONSE_TIMEOUT"
	EnvPanicOnUnexpectedError = "PANIC_ON_UNEXPECTED_ERROR"
	EnvRootPath               = "ROOT_PATH"
	EnvDirIndexingEnabled     = "DIR_INDEXING_ENABLED"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	LogLevel               zerolog.Level // Logging level (debug, info, warn, error)
	ResponseTimeout        int           // Timeout for responses in seconds
	PanicOnUnexpectedError bool          // Panic on unexpected errors when visiting a URL
	RootPath               string        // Path to serve files from
	DirIndexingEnabled     bool          // Allow client to browse directories or not
}

var CONFIG Config //nolint:gochecknoglobals

// parsePositiveInt parses and validates positive integer values.
func parsePositiveInt(param, value string) (int, error) {
	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, ValidationError{
			Param:  param,
			Value:  value,
			Reason: "must be a valid integer",
		}
	}
	if val <= 0 {
		return 0, ValidationError{
			Param:  param,
			Value:  value,
			Reason: "must be positive",
		}
	}
	return val, nil
}

func parseBool(param, value string) (bool, error) {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return false, ValidationError{
			Param:  param,
			Value:  value,
			Reason: "cannot be converted to boolean",
		}
	}
	return val, nil
}

// GetConfig loads and validates configuration from environment variables
func GetConfig() *Config {
	config := &Config{}

	// Map of environment variables to their parsing functions
	parsers := map[string]func(string) error{
		EnvLogLevel: func(v string) error {
			level, err := zerolog.ParseLevel(v)
			if err != nil {
				return ValidationError{
					Param:  EnvLogLevel,
					Value:  v,
					Reason: "must be one of: debug, info, warn, error",
				}
			}
			config.LogLevel = level
			return nil
		},
		EnvResponseTimeout: func(v string) error {
			val, err := parsePositiveInt(EnvResponseTimeout, v)
			if err != nil {
				return err
			}
			config.ResponseTimeout = val
			return nil
		},
		EnvPanicOnUnexpectedError: func(v string) error {
			val, err := parseBool(EnvPanicOnUnexpectedError, v)
			if err != nil {
				return err
			}
			config.PanicOnUnexpectedError = val
			return nil
		},
		EnvRootPath: func(v string) error {
			config.RootPath = v
			return nil
		},
		EnvDirIndexingEnabled: func(v string) error {
			val, err := parseBool(EnvDirIndexingEnabled, v)
			if err != nil {
				return err
			}
			config.DirIndexingEnabled = val
			return nil
		},
	}

	// Process each environment variable
	for envVar, parser := range parsers {
		value, ok := os.LookupEnv(envVar)
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "Missing required environment variable: %s\n", envVar)
			os.Exit(1)
		}

		if err := parser(value); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(1)
		}
	}

	return config
}
