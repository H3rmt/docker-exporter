package log

import (
	"log/slog"
	"os"
)

const (
	LevelTrace = slog.Level(-8) // More verbose than DEBUG
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
func InitLogger(format string, verbose bool, trace bool, quiet bool) {
	var handler slog.Handler
	var level slog.Level

	// Determine log level
	if trace {
		level = LevelTrace
	} else if verbose {
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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				switch level {
				case LevelTrace:
					a.Value = slog.StringValue("TRACE")
				}
			}
			return a
		},
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
