// Package main is the entry point for the vanish program
package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"vanish/cmd/commands"
	"vanish/internal/config"
	"vanish/internal/helpers"
	"vanish/internal/tui"
	"vanish/internal/types"
)

func initialModel(filenames []string, operation string, noConfirm bool) (*tui.Model, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	prog := helpers.SetUpProgress(cfg)
	styles := helpers.CreateThemeStyles(cfg)

	// Check if no_confirm is set in config and not overridden by flag
	if cfg.UI.NoConfirm && !noConfirm {
		noConfirm = true
	}

	return &tui.Model{
		Filenames:      filenames,
		FileInfos:      make([]types.FileInfo, len(filenames)),
		State:          "checking",
		Progress:       prog,
		Config:         cfg,
		Styles:         styles,
		Operation:      operation,
		ProcessedItems: make([]types.DeletedItem, 0),
		TotalFiles:     len(filenames),
		NoConfirm:      noConfirm,
	}, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		command.ShowUsage(cfg)
		return
	}

	parsed := command.ParseArgs(args, cfg)

	// Validate
	if parsed.Operation == "" || len(parsed.Filenames) == 0 {
		if parsed.Operation != "clear" {
			command.ShowUsage(cfg)
			os.Exit(1)
		}
	}

	// Initialize TUI
	m, err := initialModel(parsed.Filenames, parsed.Operation, parsed.NoConfirm)
	if err != nil {
		log.Fatalf("Error initializing: %v", err)
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
