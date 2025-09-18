package command

import (
	"fmt"
	"time"
	"vanish/internal/helpers"
	"vanish/internal/types"
)

// ShowStats prints summary statistics about the cached items, such as
// total size, counts of files and directories, and number of expired items.
// Returns an error if loading the cache index fails.
func ShowStats(config types.Config) error {
	index, err := helpers.LoadIndex(config)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	if len(index.Items) == 0 {
		fmt.Println("Cache is empty.")
		return nil
	}

	var totalSize int64
	var fileCount, dirCount int
	var expiredCount int

	cutoff := time.Now().Add(-time.Duration(config.Cache.Days) * 24 * time.Hour)

	for _, item := range index.Items {
		totalSize += item.Size
		if item.IsDirectory {
			dirCount++
		} else {
			fileCount++
		}

		if item.DeleteDate.Before(cutoff) {
			expiredCount++
		}
	}

	fmt.Printf("Vanish Cache Statistics\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Cache Directory: %s\n", helpers.ExpandPath(config.Cache.Directory))
	fmt.Printf("Total Items: %d\n", len(index.Items))
	fmt.Printf("  Files: %d\n", fileCount)
	fmt.Printf("  Directories: %d\n", dirCount)
	fmt.Printf("Total Size: %s\n", helpers.FormatBytes(totalSize))
	fmt.Printf("Retention Period: %d days\n", config.Cache.Days)
	fmt.Printf("Expired Items: %d\n", expiredCount)

	if expiredCount > 0 {
		fmt.Printf("\nRun 'vx --purge %d' to clean up expired items.\n", config.Cache.Days)
	}

	return nil
}
