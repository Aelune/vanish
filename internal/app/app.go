package app

import (
	"errors"
	"fmt"

	"vanish/internal/config"
	"vanish/internal/filesystem"
	"vanish/internal/tui"
)

func RunApp(filenames []string, clear bool, autoConfirm bool) error {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Warning: Failed to load config, using defaults")
		cfg = config.GetDefaultConfig()
	}

	// Set up logging
	// if cfg.Logging.Enabled {
	// 	if err := logging.Init(cfg); err != nil {
	// 		fmt.Println("Warning: Logging initialization failed:", err)
	// 	}
	// }

	// If --clear flag was passed
	if clear {
		if err := filesystem.ClearAllCache(cfg); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared.")
		return nil
	}

	// Validate file list
	if len(filenames) == 0 {
		return errors.New("no files or directories specified")
	}

	// Convert raw filenames into internal representations
	targets := filesystem.BuildTargets(filenames)

	// Run without confirmation (non-interactive)
	if autoConfirm {
		return filesystem.SafeDelete(cfg, targets, false /* showProgress */)
	}

	// Run TUI confirmation and deletion
	model := tui.NewModel(filenames, false, !autoConfirm, cfg)
	if err := tui.Start(model); err != nil {
    return fmt.Errorf("TUI error: %w", err)
	}


	return nil
}
