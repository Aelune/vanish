package command

import (
	"fmt"
	"strings"
	"time"

	"vanish/internal/helpers"
	"vanish/internal/types"
)

// ShowInfo searches for cached items matching the given pattern.
// If no matches are found, it notifies the user accordingly.
// If matches are found, it displays detailed metadata for each item.
func ShowInfo(pattern string, config types.Config) error {
	index, err := helpers.LoadIndex(config)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	var matchingItems []types.DeletedItem
	for _, item := range index.Items {
		if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(pattern)) {
			matchingItems = append(matchingItems, item)
		}
	}

	if len(matchingItems) == 0 {
		fmt.Printf("Oops: %s wasn't found try \"vx --list\" to check for exact name \n", pattern)
		return nil
	}

	fmt.Printf("Ahoy! Spotted %d treasure(s) in the cache map:\n", len(matchingItems))
	fmt.Println(strings.Repeat("=", 60))

	for _, item := range matchingItems {
		fmt.Printf("\nID: %s\n", item.ID)
		fmt.Printf("Original Path: %s\n", item.OriginalPath)
		fmt.Printf("Cache Path: %s\n", item.CachePath)
		fmt.Printf("Deleted: %s\n", item.DeleteDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("Type: %s\n", func() string {
			if item.IsDirectory {
				return "Directory"
			}
			return "File"
		}())
		fmt.Printf("Size: %s\n", helpers.FormatBytes(item.Size))
		if item.FileCount > 0 {
			fmt.Printf("Files Inside: %d\n", item.FileCount)
		}

		expiryDate := item.DeleteDate.Add(time.Duration(config.Cache.Days) * 24 * time.Hour)
		daysLeft := int(time.Until(expiryDate).Hours() / 24)

		if daysLeft > 0 {
			fmt.Printf("Expires: %s (%d days left)\n",
				expiryDate.Format("2006-01-02 15:04:05"), daysLeft)
		} else {
			fmt.Printf("Status: EXPIRED (can be purged)\n")
		}
		fmt.Printf("\n - To bring it back from the void, run: \"vx --restore %s\"\n", pattern)
	}

	return nil
}
