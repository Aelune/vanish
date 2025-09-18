package command

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"vanish/internal/helpers"
	"vanish/internal/types"
)

// ShowList displays a sorted list of cached files and directories,
// including metadata such as delete date, size, expiry status, and
// original path. Returns an error if loading the cache index fails.
func ShowList(config types.Config) error {
	index, err := helpers.LoadIndex(config)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	if len(index.Items) == 0 {
		fmt.Println("No cached files found.")
		return nil
	}

	// Sort by delete date (newest first)
	sort.Slice(index.Items, func(i, j int) bool {
		return index.Items[i].DeleteDate.After(index.Items[j].DeleteDate)
	})

	fmt.Printf("Cached Files (%d items):\n", len(index.Items))
	fmt.Println(strings.Repeat("=", 80))

	for _, item := range index.Items {
		fileType := "FILE"
		if item.IsDirectory {
			fileType = "DIR "
		}

		expiryDate := item.DeleteDate.Add(time.Duration(config.Cache.Days) * 24 * time.Hour)
		daysLeft := int(time.Until(expiryDate).Hours() / 24)

		status := "OK"
		if daysLeft <= 0 {
			status = "EXPIRED"
		} else if daysLeft <= 2 {
			status = "EXPIRING"
		}

		fmt.Printf("%s | %s | %8s | %s | %d days left | %s\n",
			fileType,
			item.DeleteDate.Format("2006-01-02 15:04"),
			helpers.FormatBytes(item.Size),
			status,
			daysLeft,
			item.OriginalPath,
		)
	}

	return nil
}
