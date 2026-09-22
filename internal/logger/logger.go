package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

var Log *slog.Logger
var handler *LogHandler

// Init initializes logging to a directory relative to the binary (or given dir).
func Init(logDir string) error {
	var err error
	handler, err = NewLogHandler(logDir)
	if err != nil {
		return err
	}
	Log = slog.New(handler)
	slog.SetDefault(Log)
	return nil
}

// GetLogs returns the most recent log lines (newest first) for UI display.
func GetLogs() []string {
	if handler == nil {
		return nil
	}
	return handler.GetLogs()
}

// Close closes the log file.
func Close() error {
	if handler != nil && handler.file != nil {
		return handler.file.Close()
	}
	return nil
}

// DefaultLogDir returns a consistent log directory in the current working directory.
func DefaultLogDir() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "logs")
}