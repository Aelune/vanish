package main

import (
	"fmt"
	"os"
	"golang.org/x/term"
	"vanish/internal/app"
	"vanish/internal/config"
	"vanish/internal/ui"
)

func showUsage() {
	fmt.Println("Vanish - Safe file/directory removal tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  vx <file|directory> [file2] [dir2] ...    Remove multiple files or directories safely")
	fmt.Println("  vx --noconfirm <files...>                 Remove files without confirmation")
	fmt.Println("  vx --clear                                Clear all cached files immediately")
	fmt.Println("  vx --themes                               List available themes")
	fmt.Println("  vx --debug                                Show terminal diagnostics")
	fmt.Println("  vx -h, --help                             Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  vx file1.txt")
	fmt.Println("  vx file1.txt file2.txt directory1/")
	fmt.Println("  vx --noconfirm *.log temp_folder/")
	fmt.Println("  vx *.log temp_folder/")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.config/vanish/vanish.toml")
	fmt.Println("  Default cache location: ~/.cache/vanish")
	fmt.Println("  Default retention: 10 days")
	fmt.Println("  Available themes: default, dark, light, cyberpunk, minimal")
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	arg := os.Args[1]

	// Handle help flags
	if arg == "-h" || arg == "--help" {
		showUsage()
		return
	}

	// Handle debug flag - NEW
	if arg == "--debug" {
		ui.DiagnoseTerminal()
		return
	}

	// Handle themes flag
	if arg == "--themes" {
		ui.ShowThemes()
		return
	}

	// Handle clear flag
	if arg == "--clear" {
		if err := app.RunApp([]string{}, true, false); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle noconfirm flag
	if arg == "--noconfirm" {
		if len(os.Args) < 3 {
			fmt.Println("Error: --noconfirm requires at least one file argument")
			showUsage()
			os.Exit(1)
		}
		filenames := os.Args[2:]
		if err := app.RunApp(filenames, false, true); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Normal file deletion - collect all arguments as filenames
	filenames := os.Args[1:]

	// Load config to check if auto-confirm is enabled
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config, using defaults: %v\n", err)
		cfg = config.GetDefaultConfig()
	}

	// Detect if stdout is a terminal → if not, disable TUI
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	autoConfirm := cfg.Behavior.AutoConfirm

	if !isTTY {
		// Fallback: run without TUI, always auto-confirm
		if err := app.RunApp(filenames, false, true); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Normal run
	if err := app.RunApp(filenames, false, autoConfirm); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
