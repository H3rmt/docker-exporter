package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	DebugLogger   *log.Logger
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	// By default, debug logging is disabled
	DebugLogger = log.New(io.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLogger = log.New(os.Stdout, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stderr, "WARN:  ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetVerbose enables or disables debug logging based on the verbose flag
func SetVerbose(verbose bool) {
	if verbose {
		DebugLogger.SetOutput(os.Stdout)
	} else {
		DebugLogger.SetOutput(io.Discard)
	}
}

func SetQuiet(quiet bool) {
	if quiet {
		InfoLogger.SetOutput(io.Discard)
	} else {
		InfoLogger.SetOutput(os.Stdout)
	}
}

func Debug(format string, v ...any) {
	_ = DebugLogger.Output(2, fmt.Sprintf(format, v...))
}

func Info(format string, v ...any) {
	_ = InfoLogger.Output(2, fmt.Sprintf(format, v...))
}

func Warning(format string, v ...any) {
	_ = WarningLogger.Output(2, fmt.Sprintf(format, v...))
}

func Error(format string, v ...any) {
	_ = ErrorLogger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
