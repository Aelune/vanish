// Package helpers have all the helper function.
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"vanish/internal/types"
)

// --- Logging ---

// LogOperation writes a log entry for the given operation and DeletedItem
// to the configured logging directory. If the log directory or file does
// not exist, it attempts to create them.
func LogOperation(operation string, item types.DeletedItem, config types.Config) error {
	if !config.Logging.Enabled {
		return nil
	}

	logDir := ExpandPath(config.Logging.Directory)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "vanish.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	itemType := "FILE"
	if item.IsDirectory {
		itemType = "DIR"
	}

	logEntry := fmt.Sprintf("%s [%s] %s: %s -> %s\n",
		timestamp,
		itemType,
		operation,
		item.OriginalPath,
		item.CachePath,
	)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}
func LogSimpleOperation(operation, message string, config types.Config) error {
	if !config.Logging.Enabled {
		return nil
	}

	logDir := ExpandPath(config.Logging.Directory)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "vanish.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s [%s] %s\n", timestamp, operation, message)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}
