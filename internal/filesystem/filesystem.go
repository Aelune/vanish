package filesystem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vanish/internal/config"
	"vanish/internal/models"
	"vanish/internal/logging"
)

// ExpandPath expands ~ and relative paths to absolute paths
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

// GetIndexPath returns the path to the index file
func GetIndexPath(cfg config.Config) string {
	cacheDir := ExpandPath(cfg.Cache.Directory)
	return filepath.Join(cacheDir, "index.json")
}

// LoadIndex loads the global index file
func LoadIndex(cfg config.Config) (models.Index, error) {
	var index models.Index
	indexPath := GetIndexPath(cfg)

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return models.Index{
				Items:   []models.DeletedItem{},
				Version: "1.0",
				Created: time.Now(),
				Updated: time.Now(),
			}, nil
		}
		return index, err
	}

	err = json.Unmarshal(data, &index)
	if err != nil {
		return index, err
	}

	// Ensure index has required fields
	if index.Version == "" {
		index.Version = "1.0"
	}
	if index.Created.IsZero() {
		index.Created = time.Now()
	}

	return index, nil
}

// SaveIndex saves the global index file
func SaveIndex(index models.Index, cfg config.Config) error {
	index.Updated = time.Now()
	indexPath := GetIndexPath(cfg)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

// AddToIndex adds an item to the global index
func AddToIndex(item models.DeletedItem, cfg config.Config) error {
	index, err := LoadIndex(cfg)
	if err != nil {
		return err
	}

	index.Items = append(index.Items, item)
	return SaveIndex(index, cfg)
}

// CheckFileInfo analyzes a file/directory for deletion
func CheckFileInfo(filename string, cfg config.Config) models.FileInfo {
	stat, err := os.Stat(filename)
	if err != nil {
		return models.FileInfo{
			Path:   filename,
			Exists: false,
			Error:  err.Error(),
		}
	}

	absPath, _ := filepath.Abs(filename)
	isDir := stat.IsDir()
	fileCount := 0
	size := stat.Size()

	if isDir {
		fileCount = CountFilesInDirectory(filename)
		size = GetDirectorySize(filename)
	}

	// Check if protected
	isProtected := IsProtectedPath(absPath, cfg.Safety.ProtectedPaths)

	// Check if large
	isLarge := (size > cfg.Behavior.LargeSizeLimit) ||
		       (isDir && fileCount > cfg.Behavior.LargeCountLimit)

	// Check if needs confirmation
	needsConfirm := isProtected ||
		           (cfg.Behavior.ConfirmOnLarge && isLarge) ||
		           MatchesConfirmPatterns(filename, cfg.Safety.RequireConfirm)

	return models.FileInfo{
		Path:        filename,
		IsDirectory: isDir,
		FileCount:   fileCount,
		Size:        size,
		Exists:      true,
		IsProtected: isProtected,
		IsLarge:     isLarge,
		NeedsConfirm: needsConfirm,
	}
}

// IsProtectedPath checks if a path is in the protected paths list
func IsProtectedPath(path string, protectedPaths []string) bool {
	absPath, _ := filepath.Abs(path)
	for _, protected := range protectedPaths {
		protectedAbs, _ := filepath.Abs(protected)
		if strings.HasPrefix(absPath, protectedAbs) {
			return true
		}
	}
	return false
}

// MatchesConfirmPatterns checks if filename matches any confirmation patterns
func MatchesConfirmPatterns(filename string, patterns []string) bool {
	base := filepath.Base(filename)
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

// MoveFileToCache moves a file or directory to the cache
func MoveFileToCache(filename string, cfg config.Config) (models.DeletedItem, error) {
	// Ensure cache directory exists
	cacheDir := ExpandPath(cfg.Cache.Directory)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return models.DeletedItem{}, err
	}

	// Get file info
	stat, err := os.Stat(filename)
	if err != nil {
		return models.DeletedItem{}, err
	}

	// Get absolute path
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return models.DeletedItem{}, err
	}

	// Generate unique ID and cache filename
	now := time.Now()
	id := fmt.Sprintf("%d", now.UnixNano())
	timestamp := now.Format("2006-01-02-15-04-05")
	baseFilename := filepath.Base(filename)
	cacheFilename := fmt.Sprintf("%s-%s-%s", id, timestamp, baseFilename)
	cachePath := filepath.Join(cacheDir, cacheFilename)

	isDir := stat.IsDir()
	fileCount := 0
	size := stat.Size()
	isProtected := IsProtectedPath(absPath, cfg.Safety.ProtectedPaths)

	// Calculate size and file count for directories
	if isDir {
		fileCount = CountFilesInDirectory(filename)
		size = GetDirectorySize(filename)
	}

	// Move file or directory
	if isDir {
		if err := MoveDirectory(filename, cachePath); err != nil {
			return models.DeletedItem{}, err
		}
	} else {
		if err := MoveFile(filename, cachePath); err != nil {
			return models.DeletedItem{}, err
		}
	}

	// Create backup if needed
	var backupPath string
	if cfg.Safety.BackupImportant && isProtected {
		backupPath = cachePath + ".backup"
		if isDir {
			CopyDirectory(cachePath, backupPath)
		} else {
			CopyFile(cachePath, backupPath)
		}
	}

	// Create deleted item
	item := models.DeletedItem{
		ID:           id,
		OriginalPath: absPath,
		DeleteDate:   now,
		CachePath:    cachePath,
		IsDirectory:  isDir,
		FileCount:    fileCount,
		Size:         size,
		IsProtected:  isProtected,
		BackupPath:   backupPath,
	}

	return item, nil
}

// MoveFile moves a single file
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

	// Copy permissions
	if srcInfo, err := os.Stat(src); err == nil {
		destFile.Chmod(srcInfo.Mode())
	}

	return os.Remove(src)
}

// MoveDirectory moves a directory
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

// CopyFile copies a single file
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

// CopyDirectory copies a directory recursively
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

// CountFilesInDirectory counts files in a directory recursively
func CountFilesInDirectory(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path != dir {
			count++
		}
		return nil
	})
	return count
}

// GetDirectorySize calculates the total size of a directory
func GetDirectorySize(dir string) int64 {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// CleanupOldFiles removes files older than the configured retention period
func CleanupOldFiles(cfg config.Config) error {
	cutoffDays := time.Duration(cfg.Cache.Days) * 24 * time.Hour
	cutoff := time.Now().Add(-cutoffDays)

	index, err := LoadIndex(cfg)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	var remainingItems []models.DeletedItem
	cleanedCount := 0

	for _, item := range index.Items {
		if item.DeleteDate.Before(cutoff) {
			// Remove the actual file or directory
			if item.IsDirectory {
				os.RemoveAll(item.CachePath)
			} else {
				os.Remove(item.CachePath)
			}

			// Remove backup if exists
			if item.BackupPath != "" {
				os.RemoveAll(item.BackupPath)
			}

			cleanedCount++

			// Log cleanup if logging is enabled
			if cfg.Logging.Enabled {
				logging.LogOperation("CLEANUP", item, cfg)
			}
		} else {
			remainingItems = append(remainingItems, item)
		}
	}

	// Update index
	index.Items = remainingItems
	return SaveIndex(index, cfg)
}

// ClearAllCache removes all cached files
func ClearAllCache(cfg config.Config) error {
	cacheDir := ExpandPath(cfg.Cache.Directory)

	// Remove all files in cache directory
	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}

	// Recreate cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	// Create empty index
	index := models.Index{
		Items:   []models.DeletedItem{},
		Version: "1.0",
		Created: time.Now(),
		Updated: time.Now(),
	}
	if err := SaveIndex(index, cfg); err != nil {
		return err
	}

	// Log clear operation
	if cfg.Logging.Enabled {
		logging.LogOperation("CLEAR_ALL", models.DeletedItem{
			OriginalPath: "cache",
			DeleteDate:   time.Now(),
		}, cfg)
	}

	return nil
}

func SafeDelete(cfg config.Config, items []models.DeletedItem, showProgress bool) error {
	for _, item := range items {
		// Real logic would involve checks and user prompts
		moved, err := MoveFileToCache(item.OriginalPath, cfg)
		if err != nil {
			logging.LogError("DELETE_FAIL", item.OriginalPath, err, cfg)
			fmt.Printf("Failed to delete: %s\n", item.OriginalPath)
			continue
		}
		logging.LogOperation("DELETE", moved, cfg)
		fmt.Printf("Deleted: %s -> %s\n", moved.OriginalPath, moved.CachePath)
	}
	return nil
}

func BuildTargets(filenames []string) []models.DeletedItem {
	var targets []models.DeletedItem
	for _, f := range filenames {
		abs, _ := filepath.Abs(f)
		targets = append(targets, models.DeletedItem{
			OriginalPath: abs,
		})
	}
	return targets
}
