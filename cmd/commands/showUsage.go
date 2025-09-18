package command

import (
	"fmt"
	"vanish/internal/types"
)

// ShowUsage prints how vanish is to be used
func ShowUsage(config types.Config) {
	fmt.Println("Vanish (vx) - Safe file/directory removal tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  vx <file|directory> [file2] [dir2] ...        Remove files/directories safely")
	fmt.Println("  vx --clear                                    Clear all cached files immediately")
	fmt.Println("  vx --restore <pattern> [pattern2] ...        Restore files matching patterns")
	fmt.Println("  vx --list                                     Show all cached files")
	fmt.Println("  vx --info <pattern>                           Show detailed info about cached item(s)")
	fmt.Println("  vx --stats                                    Show cache statistics")
	fmt.Println("  vx --purge <days>                             Delete files older than N days")
	fmt.Println("  vx --path                                     Print cache directory path")
	fmt.Println("  vx --config-path                              Print config file path")
	fmt.Println("  vx --themes                                   List available themes")
	fmt.Println("  vx --noconfirm                                Skip confirmation prompts")
	fmt.Println("  vx -h, --help                                 Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  vx file1.txt                                  Delete file1.txt safely")
	fmt.Println("  vx file1.txt file2.txt directory1/            Delete multiple items")
	fmt.Println("  vx --noconfirm *.log temp_folder/             Delete without confirmation")
	fmt.Println("  vx --restore file1.txt                        Restore file1.txt from cache")
	fmt.Println("  vx --restore \"*temp*\"                         Restore all files with 'temp' in name")
	fmt.Println("  vx --purge 5                                  Delete cached files older than 5 days")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Printf("  Cache location: %s\n", config.Cache.Directory)
	fmt.Printf("  Default retention: %d days\n", config.Cache.Days)
	fmt.Printf("  No confirm mode: %v\n", config.UI.NoConfirm)
	fmt.Printf("  Current theme: %s\n", config.UI.Theme)
	// fmt.Println("  Config file: ~/.config/vanish/vanish.toml")
	// fmt.Println("  cache location: ~/.cache/vanish")
	// fmt.Println("  Default retention: 10 days")
	// fmt.Println("  Available themes: default, dark, light, cyberpunk, minimal")
}
