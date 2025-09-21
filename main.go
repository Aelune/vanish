// Package main is the entry point for the vanish program
package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"vanish/cmd/commands"
	"vanish/internal/config"
	"vanish/internal/tui"
)

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
	m, err := tui.InitialModel(parsed.Filenames, parsed.Operation, parsed.NoConfirm)
	if err != nil {
		log.Fatalf("Error initializing: %v", err)
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
