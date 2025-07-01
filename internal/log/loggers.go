package log

import (
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
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
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
