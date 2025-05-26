package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config holds the application configuration loaded from CLI flags.
type Config struct {
	LogLevel           slog.Level // Logging level (debug, info, warn, error)
	ResponseTimeout    int        // Timeout for responses in seconds
	RootPath           string     // Path to serve files from
	DirIndexingEnabled bool       // Allow client to browse directories or not
	Listen             string     // Address to listen on
}

var CONFIG Config //nolint:gochecknoglobals

// parseLogLevel parses a log level string into slog.Level
func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}

// GetConfig loads and validates configuration from CLI flags
func GetConfig() *Config {
	// Define CLI flags with defaults
	logLevel := flag.String("log-level", "info", "Logging level (debug, info, warn, error)")
	responseTimeout := flag.Int("response-timeout", 30, "Timeout for responses in seconds")
	rootPath := flag.String("root-path", "", "Path to serve files from")
	dirIndexing := flag.Bool("dir-indexing", false, "Allow client to browse directories")
	listen := flag.String("listen", "localhost:1965", "Address to listen on")

	flag.Parse()

	// Parse and validate log level
	level, err := parseLogLevel(*logLevel)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid log level '%s': must be one of: debug, info, warn, error\n", *logLevel)
		os.Exit(1)
	}

	// Validate response timeout
	if *responseTimeout <= 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Invalid response timeout '%d': must be positive\n", *responseTimeout)
		os.Exit(1)
	}

	return &Config{
		LogLevel:           level,
		ResponseTimeout:    *responseTimeout,
		RootPath:           *rootPath,
		DirIndexingEnabled: *dirIndexing,
		Listen:             *listen,
	}
}
