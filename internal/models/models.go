package models

import (
	"time"
)

// DeletedItem represents an item in the global index
type DeletedItem struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	DeleteDate   time.Time `json:"delete_date"`
	CachePath    string    `json:"cache_path"`
	IsDirectory  bool      `json:"is_directory"`
	FileCount    int       `json:"file_count,omitempty"`
	Size         int64     `json:"size"`
	IsProtected  bool      `json:"is_protected,omitempty"`
	BackupPath   string    `json:"backup_path,omitempty"`
}

// Index represents the global index file
type Index struct {
	Items   []DeletedItem `json:"items"`
	Version string        `json:"version"`
	Created time.Time     `json:"created"`
	Updated time.Time     `json:"updated"`
}

// FileInfo holds information about a file to be deleted
type FileInfo struct {
	Path        string
	IsDirectory bool
	FileCount   int
	Size        int64
	Exists      bool
	Error       string
	IsProtected bool
	IsLarge     bool
	NeedsConfirm bool
}

// OperationStats tracks statistics for the current operation
type OperationStats struct {
	TotalFiles     int
	TotalDirs      int
	ProcessedFiles int
	ProcessedDirs  int
	SkippedFiles   int
	ErrorCount     int
	TotalSize      int64
	ProcessedSize  int64
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	Path      string    `json:"path"`
	CachePath string    `json:"cache_path,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Error     string    `json:"error,omitempty"`
}
