// Package types contains all the shared types used across the Vanish (vx) application.
// It excludes TUI-specific models defined in the tui package.
package types

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Config holds the user configuration loaded from the config file.
type Config struct {
	Cache struct {
		Directory string `toml:"directory"`
		Days      int    `toml:"days"`
		NoConfirm bool   `toml:"no_confirm"`
	} `toml:"cache"`
	Logging struct {
		Enabled   bool   `toml:"enabled"`
		Directory string `toml:"directory"`
	} `toml:"logging"`
	// Notifications struct {
	// 	DesktopEnabled bool `toml:"desktop_enabled"`
	// 	NotifySuccess  bool `toml:"notify_success"`
	// 	NotifyErrors   bool `toml:"notify_errors"`
	// } `toml:"notifications"`
	UI struct {
		Theme  string `toml:"theme"` // "default", "dark", "light", "cyberpunk", "minimal"
		Colors struct {
			Primary   string `toml:"primary"`
			Secondary string `toml:"secondary"`
			Success   string `toml:"success"`
			Warning   string `toml:"warning"`
			Error     string `toml:"error"`
			Text      string `toml:"text"`
			Muted     string `toml:"muted"`
			Border    string `toml:"border"`
			Highlight string `toml:"highlight"`
		} `toml:"colors"`
		Progress struct {
			Style     string `toml:"style"` // "gradient", "solid", "rainbow"
			ShowEmoji bool   `toml:"show_emoji"`
			Animation bool   `toml:"animation"`
		} `toml:"progress"`
	} `toml:"ui"`
}

// DeletedItem represents an item that has been moved to cache
type DeletedItem struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	DeleteDate   time.Time `json:"delete_date"`
	CachePath    string    `json:"cache_path"`
	IsDirectory  bool      `json:"is_directory"`
	IsSymlink    bool      `json:"is_symlink"`
	LinkTarget   string    `json:"link_target,omitempty"` // Only populated for symlinks
	FileCount    int       `json:"file_count,omitempty"`
	Size         int64     `json:"size"`
}

// Index represents the global index file
type Index struct {
	Items []DeletedItem `json:"items"`
}

// FileInfo holds information about a file to be deleted
type FileInfo struct {
	Path        string
	IsDirectory bool
	FileCount   int
	Exists      bool
	Error       string
}

// ThemeStyles holds all the styled components used in the TUI
type ThemeStyles struct {
	Root       lipgloss.Style
	Base       lipgloss.Style
	Title      lipgloss.Style
	Header     lipgloss.Style
	Question   lipgloss.Style
	Filename   lipgloss.Style
	IconStyle  lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style
	Info       lipgloss.Style
	Help       lipgloss.Style
	Progress   lipgloss.Style
	Border     lipgloss.Style
	List       lipgloss.Style
	StatusGood lipgloss.Style
	StatusBad  lipgloss.Style
}

// ---- Types for Messages ----

// FilesExistMsg is sent when checking whether specific files or directories exist
type FilesExistMsg struct {
	FileInfos []FileInfo
}

// RestoreItemsMsg contains a list of items to be restored.
type RestoreItemsMsg struct {
	Items []DeletedItem
}

// FileMoveMsg represents the result of a file move operation.
type FileMoveMsg struct {
	Item DeletedItem
	Err  error
}

// RestoreMsg represents the result of restoring a deleted item.
type RestoreMsg struct {
	Item DeletedItem
	Err  error
}

// CleanupMsg indicates that a cleanup action has occurred.
type CleanupMsg struct{}

// ClearMsg represents the result of clearing cached files.
type ClearMsg struct {
	Err error
}

// PurgeMsg contains information about files purged from the cache.
type PurgeMsg struct {
	PurgedCount int
	Err         error
}

// ErrorMsg is a generic error message used across the application.
type ErrorMsg string

// ItemType returns a human-readable string describing the item type
func (item DeletedItem) ItemType() string {
	if item.IsSymlink {
		return "symlink"
	}
	if item.IsDirectory {
		return "directory"
	}
	return "file"
}
