// Package command is responsibe for parsing command line args and perform operations based on args
package command

import (
	"fmt"
	"log"
	"os"

	"vanish/internal/helpers"
	"vanish/internal/types"
)

// ParsedArgs holds the result of parsing CLI arguments
type ParsedArgs struct {
	Operation string
	Filenames []string
	NoConfirm bool
}

// ParseArgs parses the command-line arguments and returns the operation, filenames, and flags
func ParseArgs(args []string, cfg types.Config) ParsedArgs {
	var operation string
	var filenames []string
	var noConfirm bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			ShowUsage(cfg)
			os.Exit(0)
		case "--themes":
			displayer := &MainThemeDisplayer{}
			ShowThemesWithTuiPreview(displayer)
			os.Exit(0)
		case "--path":
			fmt.Println(helpers.ExpandPath(cfg.Cache.Directory))
			os.Exit(0)
		case "--config-path":
			fmt.Println(helpers.GetConfigPath())
			os.Exit(0)
		case "--list":
			if err := ShowList(cfg); err != nil {
				log.Fatalf("Error: %v", err)
			}
			os.Exit(0)
		case "--stats":
			if err := ShowStats(cfg); err != nil {
				log.Fatalf("Error: %v", err)
			}
			os.Exit(0)
		case "--clear":
			operation = "clear"
			filenames = []string{""}
		case "--noconfirm":
			noConfirm = true
		case "--restore":
			operation = "restore"
			if i+1 < len(args) {
				filenames = args[i+1:]
				i = len(args) // consume remaining args
			} else {
				log.Fatal("Error: --restore requires at least one pattern")
			}
		case "--info":
			if i+1 < len(args) {
				if err := ShowInfo(args[i+1], cfg); err != nil {
					log.Fatalf("Error: %v", err)
				}
			} else {
				log.Fatal("Error: --info requires a pattern")
			}
			os.Exit(0)
		case "--purge":
			if i+1 < len(args) {
				operation = "purge"
				filenames = []string{args[i+1]}
				i++ // skip value
			} else {
				log.Fatal("Error: --purge requires number of days")
			}
		default:
			// If no operation is set yet, assume delete
			if operation == "" {
				operation = "delete"
				filenames = args[i:]
				i = len(args) // consume all
			}
		}

		if operation == "restore" || operation == "delete" {
			break
		}
	}

	// Fallback to delete if nothing else is matched
	if operation == "" && len(filenames) == 0 && len(args) > 0 {
		operation = "delete"
		for _, arg := range args {
			if arg != "--noconfirm" {
				filenames = append(filenames, arg)
			}
		}
	}

	return ParsedArgs{
		Operation: operation,
		Filenames: filenames,
		NoConfirm: noConfirm,
	}
}
