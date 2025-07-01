package log

import (
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
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
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

func Debug(v ...any) {
	DebugLogger.Println(v...)
}

func Info(v ...any) {
	InfoLogger.Println(v...)
}

func Warning(v ...any) {
	WarningLogger.Println(v...)
}

func Error(v ...any) {
	ErrorLogger.Println(v...)
	os.Exit(1)
}
