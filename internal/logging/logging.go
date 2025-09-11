package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"vanish/internal/config"
	"vanish/internal/models"
)

// Logger handles all logging operations
type Logger struct {
	config config.Config
	logDir string
}

// NewLogger creates a new logger instance
func NewLogger(cfg config.Config) *Logger {
	logDir := expandPath(cfg.Logging.Directory)
	return &Logger{
		config: cfg,
		logDir: logDir,
	}
}

// LogOperation logs a file operation
func LogOperation(operation string, item models.DeletedItem, cfg config.Config) {
	if !cfg.Logging.Enabled {
		return
	}

	logger := NewLogger(cfg)
	logger.logOperation(operation, item)
}

func (l *Logger) logOperation(operation string, item models.DeletedItem) {
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return
	}

	// Create log entry
	entry := models.LogEntry{
		Timestamp: time.Now(),
		Operation: operation,
		Path:      item.OriginalPath,
		CachePath: item.CachePath,
		Size:      item.Size,
	}

	// Write to text log
	l.writeTextLog(entry)

	// Write to JSON log if debug level
	if l.config.Logging.Level == "debug" {
		l.writeJSONLog(entry)
	}
}

func (l *Logger) writeTextLog(entry models.LogEntry) {
	logPath := filepath.Join(l.logDir, "vanish.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	var logEntry string
	if entry.CachePath != "" {
		logEntry = fmt.Sprintf("%s %s %s -> %s (Size: %d bytes)\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Operation,
			entry.Path,
			entry.CachePath,
			entry.Size)
	} else {
		logEntry = fmt.Sprintf("%s %s %s\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Operation,
			entry.Path)
	}

	logFile.WriteString(logEntry)
}

func (l *Logger) writeJSONLog(entry models.LogEntry) {
	logPath := filepath.Join(l.logDir, "vanish.json")

	var entries []models.LogEntry

	// Read existing entries
	if data, err := os.ReadFile(logPath); err == nil {
		json.Unmarshal(data, &entries)
	}

	// Add new entry
	entries = append(entries, entry)

	// Keep only last 1000 entries to prevent file from growing too large
	if len(entries) > 1000 {
		entries = entries[len(entries)-1000:]
	}

	// Write back to file
	if data, err := json.MarshalIndent(entries, "", "  "); err == nil {
		os.WriteFile(logPath, data, 0644)
	}
}

// LogError logs an error
func LogError(operation string, path string, err error, cfg config.Config) {
	if !cfg.Logging.Enabled {
		return
	}

	logger := NewLogger(cfg)
	entry := models.LogEntry{
		Timestamp: time.Now(),
		Operation: operation,
		Path:      path,
		Error:     err.Error(),
	}

	logger.writeTextLog(entry)
	if cfg.Logging.Level == "debug" {
		logger.writeJSONLog(entry)
	}
}

// LogInfo logs general information
func LogInfo(operation string, message string, cfg config.Config) {
	if !cfg.Logging.Enabled || cfg.Logging.Level == "error" {
		return
	}

	logger := NewLogger(cfg)
	entry := models.LogEntry{
		Timestamp: time.Now(),
		Operation: operation,
		Path:      message,
	}

	logger.writeTextLog(entry)
	if cfg.Logging.Level == "debug" {
		logger.writeJSONLog(entry)
	}
}

// GetLogStats returns statistics about the logs
func GetLogStats(cfg config.Config) (map[string]interface{}, error) {
	if !cfg.Logging.Enabled {
		return nil, fmt.Errorf("logging is disabled")
	}

	logger := NewLogger(cfg)
	jsonLogPath := filepath.Join(logger.logDir, "vanish.json")

	data, err := os.ReadFile(jsonLogPath)
	if err != nil {
		return nil, err
	}

	var entries []models.LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := make(map[string]interface{})
	opCounts := make(map[string]int)
	var totalSize int64

	for _, entry := range entries {
		opCounts[entry.Operation]++
		totalSize += entry.Size
	}

	stats["total_operations"] = len(entries)
	stats["operations_by_type"] = opCounts
	stats["total_size_processed"] = totalSize

	if len(entries) > 0 {
		stats["first_entry"] = entries[0].Timestamp
		stats["last_entry"] = entries[len(entries)-1].Timestamp
	}

	return stats, nil
}

func expandPath(path string) string {
	if path[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}
	if !filepath.IsAbs(path) {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path)
	}
	return path
}
