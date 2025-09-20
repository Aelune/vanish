// Package helpers have all the helper function.
package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"io"
	// "log"
	"os"
	// "os/exec"
	"path/filepath"
	// "runtime"
	"strconv"
	"strings"
	"time"
	"vanish/internal/types"
)

// GetConfigPath returns path to vanish.toml
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "could find Config File"
	}
	return filepath.Join(homeDir, ".config", "vanish", "vanish.toml")
}

// FormatBytes formats bytes and is used in
// cmd/commands[showList.go,showInfo.go]
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// SendNotification sends a desktop notification based on the provided title and message.
// It only sends notifications if the corresponding flags are enabled in the config.
// It's tested on only Linux tho it should also work on macOS, and Windows platforms.
// func SendNotification(title, message string, config types.Config) {
// 	if !config.Notifications.NotifySuccess && !config.Notifications.NotifyErrors {
// 		return
// 	}

// 	if config.Notifications.DesktopEnabled {
// 		// Run the notification in a separate goroutine to avoid blocking the UI.
// 		go func() {
// 			var err error

// 			switch runtime.GOOS {
// 			case "linux":
// 				err = exec.Command("notify-send", title, message).Run()
// 			case "darwin":
// 				script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
// 				err = exec.Command("osascript", "-e", script).Run()
// 			}

// 			if err != nil {
// 				log.Printf("failed to send notification: %v", err)
// 			}
// 		}()
// 	}
// }


// SetUpProgress defines progress bar style
func SetUpProgress(config types.Config) progress.Model {
	prog := progress.New()
	prog.Width = 50

	switch config.UI.Progress.Style {
	case "solid":
		prog = progress.New(progress.WithSolidFill(config.UI.Colors.Primary))
	case "rainbow":
		prog = progress.New(progress.WithGradient("#FF0000", "#9400D3")) //  "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#4B0082",
	default: // gradient
		prog = progress.New(progress.WithGradient(config.UI.Colors.Primary, config.UI.Colors.Secondary))
	}
	return prog
}

// CreateThemeStyles create lipgloss themes
func CreateThemeStyles(config types.Config) types.ThemeStyles {
	colors := config.UI.Colors
	return types.ThemeStyles{
		Root: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2),
		// PaddingBottom(2).
		// PaddingLeft(4),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Text)).
			Bold(true).
			Padding(0, 2, 0, 2).
			MarginBottom(1),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true).
			Underline(true).
			MarginBottom(1),
		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true),
		Filename: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Highlight)).
			Bold(true).
			Underline(true),
		IconStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Highlight)).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Success)),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Error)),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Warning)).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Secondary)).
			Padding(0, 1),
		// Removed border and made it responsive
		// Border(lipgloss.NormalBorder()).
		// BorderForeground(lipgloss.Color(colors.Border))

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Muted)).
			Italic(true).
			MarginTop(1),
		Progress: lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1),
		Border: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(colors.Border)).
			Padding(1),
		List: lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1),
		StatusGood: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)),
		StatusBad: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)),
	}
}

// RenderThemeAsString renders strings for dummy ui in --themes flags
// used in cmd/commands/showThemes.go
func RenderThemeAsString(cfg types.Config) string {
	styles := CreateThemeStyles(cfg)

	result := styles.Root.Render(
		styles.Question.Render("Are you sure you want to delete the following items?") + "\n\n" +
			styles.List.Render(
				"  📄 "+styles.Filename.Render("example.txt")+"\n"+
					"  📁 "+styles.Filename.Render("temp_folder/")+styles.Info.Render(" (5 items)")+"\n",
			) + "\n" +
			styles.Info.Render("Total items to delete: 2 | Recoverable for 10 days") + "\n\n" +
			func() string {
				prog := SetUpProgress(cfg)
				return styles.Progress.Render(prog.ViewAs(0.75))
			}() + "\n\n" +
			styles.Success.Render("✅ Success") + "\n" + " " +
			styles.Warning.Render("⚠ Warning") + "\n" +
			styles.Error.Render("❌ Error") + "\n" + " " +
			styles.Help.Render("Press 'y' to confirm, 'n' to cancel"),
	)

	return result + "\n"
}

// GetTerminalSize returns the current terminal width and height
func GetTerminalSize() (int, int) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			return width, height
		}
	}
	// Fallback to reasonable defaults
	return 80, 24
}

// ExpandPath expands a given path by resolving '~/' to the user's home directory
// and converting relative paths to absolute paths. If the path is already absolute,
// it is returned unchanged.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	if !filepath.IsAbs(path) {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path)
	}
	return path
}

// --- Index Helpers ---

// SaveIndex serializes the provided index to JSON and writes it to disk
// at the location specified by the given config. Returns an error if
// marshalling or writing to file fails.
func SaveIndex(index types.Index, config types.Config) error {
	indexPath := GetIndexPath(config)
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

// GetIndexPath returns the full path to the index.json file used to
// store metadata about cached files, based on the provided config.
func GetIndexPath(config types.Config) string {
	cacheDir := ExpandPath(config.Cache.Directory)
	return filepath.Join(cacheDir, "index.json")
}

// LoadIndex reads and unmarshals the index.json file into an Index struct.
// If the file does not exist, it returns an empty Index. Returns an error
// if reading or unmarshalling fails.
func LoadIndex(config types.Config) (types.Index, error) {
	var index types.Index
	indexPath := GetIndexPath(config)

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return types.Index{Items: []types.DeletedItem{}}, nil
		}
		return index, err
	}

	err = json.Unmarshal(data, &index)
	return index, err
}

// AddToIndex adds a DeletedItem to the index and saves the updated
// index to disk using the provided config. Returns an error if loading
// or saving the index fails.
func AddToIndex(item types.DeletedItem, config types.Config) error {
	index, err := LoadIndex(config)
	if err != nil {
		return err
	}

	index.Items = append(index.Items, item)
	return SaveIndex(index, config)
}

// RemoveFromIndex removes a DeletedItem with the specified ID from the
// index and saves the updated index to disk. Returns an error if loading
// or saving the index fails.
func RemoveFromIndex(itemID string, config types.Config) error {
	index, err := LoadIndex(config)
	if err != nil {
		return err
	}

	var remainingItems []types.DeletedItem
	for _, item := range index.Items {
		if item.ID != itemID {
			remainingItems = append(remainingItems, item)
		}
	}

	index.Items = remainingItems
	return SaveIndex(index, config)
}

// --- Logging ---

// LogOperation writes a log entry for the given operation and DeletedItem
// to the configured logging directory. If the log directory or file does
// not exist, it attempts to create them.
func LogOperation(operation string, item types.DeletedItem, config types.Config) {
	logDir := ExpandPath(config.Logging.Directory)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	logPath := filepath.Join(logDir, "vanish.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	logEntry := fmt.Sprintf("%s %s %s -> %s (Size: %d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		operation,
		item.OriginalPath,
		item.CachePath,
		item.Size)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return
	}
}

// INIT()

// ClearAllCache removes all cached files and directories, recreates the
// cache directory, resets the index, and logs the operation if logging
// is enabled. Returns a tea.Msg with any error encountered.
func ClearAllCache(config types.Config) tea.Cmd {
	return func() tea.Msg {
		cacheDir := ExpandPath(config.Cache.Directory)

		// Remove all files in cache directory
		if err := os.RemoveAll(cacheDir); err != nil {
			return types.ClearMsg{Err: err}
		}

		// Recreate cache directory
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return types.ClearMsg{Err: err}
		}

		// Create empty index
		index := types.Index{Items: []types.DeletedItem{}}
		if err := SaveIndex(index, config); err != nil {
			return types.ClearMsg{Err: err}
		}

		// Log clear operation
		if config.Logging.Enabled {
			logDir := ExpandPath(config.Logging.Directory)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return types.ClearMsg{Err: err}
			}
			logPath := filepath.Join(logDir, "vanish.log")
			logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer logFile.Close()
				if _, err := logFile.WriteString(fmt.Sprintf("%s CLEAR_ALL Cache cleared\n", time.Now().Format("2006-01-02 15:04:05"))); err != nil {
					return types.ClearMsg{Err: err}
				}
			}
		}

		return types.ClearMsg{Err: nil}
	}
}

// PurgeOldFiles removes cached files and directories that are older than
// the specified number of days. Updates the index and logs each purge
// if logging is enabled. Returns a tea.Msg containing the purge results.
func PurgeOldFiles(config types.Config, daysStr string) tea.Cmd {
	return func() tea.Msg {
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return types.PurgeMsg{Err: fmt.Errorf("invalid days value: %s", daysStr)}
		}

		cutoffDays := time.Duration(days) * 24 * time.Hour
		cutoff := time.Now().Add(-cutoffDays)

		index, err := LoadIndex(config)
		if err != nil {
			return types.PurgeMsg{Err: fmt.Errorf("error loading index: %v", err)}
		}

		var remainingItems []types.DeletedItem
		purgedCount := 0

		for _, item := range index.Items {
			if item.DeleteDate.Before(cutoff) {
				// Remove the actual file or directory
				if item.IsDirectory {
					os.RemoveAll(item.CachePath)
				} else {
					os.Remove(item.CachePath)
				}
				purgedCount++

				// Log purge
				if config.Logging.Enabled {
					LogOperation("PURGE", item, config)
				}
			} else {
				remainingItems = append(remainingItems, item)
			}
		}

		// Update index
		index.Items = remainingItems
		if err := SaveIndex(index, config); err != nil {
			return types.PurgeMsg{Err: fmt.Errorf("error updating index: %v", err)}
		}

		return types.PurgeMsg{PurgedCount: purgedCount, Err: nil}
	}
}

// CheckRestoreItems searches the index for deleted items that match
// any of the given patterns (case-insensitive substring match).
// Returns a tea.Msg containing the matched items.
func CheckRestoreItems(patterns []string, config types.Config) tea.Cmd {
	return func() tea.Msg {
		index, err := LoadIndex(config)
		if err != nil {
			return types.ErrorMsg(fmt.Sprintf("Error loading index: %v", err))
		}

		var matchingItems []types.DeletedItem
		for _, pattern := range patterns {
			for _, item := range index.Items {
				// Simple pattern matching - check if pattern is contained in original path
				if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(pattern)) {
					matchingItems = append(matchingItems, item)
				}
			}
		}

		return types.RestoreItemsMsg{Items: matchingItems}
	}
}

// CheckFilesExist checks if the specified files or directories exist on disk,
// gathers metadata about each, and returns a tea.Msg with the results.
func CheckFilesExist(filenames []string) tea.Cmd {
	return func() tea.Msg {
		fileInfos := make([]types.FileInfo, len(filenames))

		for i, filename := range filenames {
			stat, err := os.Stat(filename)
			if err != nil {
				fileInfos[i] = types.FileInfo{
					Path:   filename,
					Exists: false,
					Error:  err.Error(),
				}
				continue
			}

			isDir := stat.IsDir()
			fileCount := 0

			if isDir {
				fileCount, _ = CountFilesInDirectory(filename)
			}

			fileInfos[i] = types.FileInfo{
				Path:        filename,
				IsDirectory: isDir,
				FileCount:   fileCount,
				Exists:      true,
			}
		}

		return types.FilesExistMsg{FileInfos: fileInfos}
	}
}

// CountFilesInDirectory returns the number of files (not including directories)
// in the specified directory and its subdirectories. Errors during walking
// the directory tree are ignored.
func CountFilesInDirectory(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return nil // skips problematic files
		}
		if path != dir {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err // return the error from filepath.Walk
	}
	return count, nil
}

// CountValidFiles returns the number of FileInfo entries that represent
// existing files or directories.
func CountValidFiles(fileInfos []types.FileInfo) int {
	count := 0
	for _, info := range fileInfos {
		if info.Exists {
			count++
		}
	}
	return count
}

// FindNextValidFile returns the index of the next valid file (i.e., one that exists)
// in the given FileInfo slice, starting from startIndex. Returns -1 if none found.
func FindNextValidFile(fileInfos []types.FileInfo, startIndex int) int {
	for i := startIndex; i < len(fileInfos); i++ {
		if fileInfos[i].Exists {
			return i
		}
	}
	return -1
}

// MoveFile moves a file from the source path to the destination path by copying
// its contents and removing the original. Returns an error if any operation fails.
func MoveFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Remove(src)
}

// MoveDirectory moves a directory from src to dst. Attempts an atomic move
// using os.Rename first, and falls back to a copy-and-remove approach
// if that fails
func MoveDirectory(src, dst string) error {
	// Use os.Rename for atomic operation when possible (same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback to copy + remove for cross-filesystem moves
	if err := CopyDirectory(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// CopyDirectory recursively copies the contents of the source directory to the
// destination directory. Preserves file and directory modes. Returns an error
// if any operation fails.
func CopyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a file from src to dst, preserving its permissions.
// Returns an error if opening, copying, or creating fails.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GetDirectorySize returns the total size in bytes of all non-directory
// files within the specified directory and its subdirectories.
func GetDirectorySize(dir string) (int64, error) {
	var size int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err // return the error from filepath.Walk
	}
	return size, nil
}
