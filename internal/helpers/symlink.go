package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

// IsSymlink checks if the given path is a symbolic link
func IsSymlink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path) // Use Lstat instead of Stat to not follow symlinks
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

// MoveSymlink handles moving a symbolic link to cache
// It reads the link target and recreates the symlink at the destination
func MoveSymlink(src, dst string) error {
	// Read the link target
	linkTarget, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("failed to read symlink: %w", err)
	}

	// Create the symlink at destination
	if err := os.Symlink(linkTarget, dst); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Remove the original symlink
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove original symlink: %w", err)
	}

	return nil
}





// CopyDirectory recursively copies the contents of the source directory to the
// destination directory. Preserves file and directory modes, and handles symlinks.



// GetFileInfo returns detailed information about a file, including whether it's a symlink
func GetFileInfo(path string) (os.FileInfo, bool, error) {
	// Use Lstat to not follow symlinks
	info, err := os.Lstat(path)
	if err != nil {
		return nil, false, err
	}

	isSymlink := info.Mode()&os.ModeSymlink != 0
	return info, isSymlink, nil
}

// RestoreSymlink restores a symbolic link from cache back to its original location
func RestoreSymlink(cachePath, originalPath string) error {
	// Read the link target from cache
	linkTarget, err := os.Readlink(cachePath)
	if err != nil {
		return fmt.Errorf("failed to read cached symlink: %w", err)
	}

	// Create directory for original path if needed
	originalDir := filepath.Dir(originalPath)
	if err := os.MkdirAll(originalDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Recreate the symlink at original location
	if err := os.Symlink(linkTarget, originalPath); err != nil {
		return fmt.Errorf("failed to restore symlink: %w", err)
	}

	// Remove from cache
	if err := os.Remove(cachePath); err != nil {
		return fmt.Errorf("failed to remove cached symlink: %w", err)
	}

	return nil
}
