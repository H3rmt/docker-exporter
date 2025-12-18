package log

import (
	"fmt"
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
		Level: level,
		AddSource: true,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "logfmt", "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
}

// SetVerbose enables or disables debug logging based on the verbose flag
// Deprecated: Use InitLogger instead
func SetVerbose(verbose bool) {
	// For backward compatibility, reinitialize with current settings
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		AddSource: true,
	})
	logger = slog.New(handler)
}

// SetQuiet sets quiet mode
// Deprecated: Use InitLogger instead
func SetQuiet(quiet bool) {
	var level slog.Level
	if quiet {
		level = slog.LevelWarn
	} else {
		level = slog.LevelInfo
	}
	
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		AddSource: true,
	})
	logger = slog.New(handler)
}

func Debug(format string, v ...any) {
	logger.Debug(fmt.Sprintf(format, v...))
}

func DebugWith(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(format string, v ...any) {
	logger.Info(fmt.Sprintf(format, v...))
}

func InfoWith(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Warning(format string, v ...any) {
	logger.Warn(fmt.Sprintf(format, v...))
}

func WarningWith(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Error(format string, v ...any) {
	logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func ErrorWith(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}

// GetStdLogger returns a standard library logger that writes to the slog logger
// This is useful for libraries that expect a *log.Logger
func GetStdLogger() *slog.Logger {
	return logger
}
