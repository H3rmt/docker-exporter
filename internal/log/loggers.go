package log

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	// Default to text format (logfmt-like) with INFO level
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
}

// InitLogger initializes the logger with the specified format and verbosity
func InitLogger(format string, verbose bool, quiet bool) {
	var handler slog.Handler
	var level slog.Level

	// Determine log level
	if verbose {
		level = slog.LevelDebug
	} else if quiet {
		level = slog.LevelWarn
	} else {
		level = slog.LevelInfo
	}

	// Create handler based on format
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "logfmt":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
}

// GetLogger returns the underlying slog.Logger instance
// Use with slog.NewLogLogger() to get a *log.Logger for stdlib compatibility
func GetLogger() *slog.Logger {
	return logger
}
