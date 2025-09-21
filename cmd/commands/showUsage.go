package command

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"vanish/internal/helpers"
	"vanish/internal/types"
)

// ShowUsageSmart that detects color support of terminal
func ShowUsageSmart(config types.Config) {
	// Check if terminal supports colors
	if helpers.IsColorTerminal() {
		ShowUsage(config)
	} else {
		ShowUsageFallback(config)
	}
}

// ShowUsage prints how to use vanish with example
func ShowUsage(config types.Config) {
	// Create specific styles for help output using theme colors
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Primary)).
		Bold(true).
		Underline(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Success)).
		Bold(true).
		MarginTop(1)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Highlight)).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Text))

	flagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Secondary))

	exampleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Muted)).
		Italic(true).
		MarginLeft(2)

	configStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Warning)).
		MarginLeft(2)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.UI.Colors.Muted)).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(1)

	fmt.Println(titleStyle.Render("Vanish (vx) - Safe file/directory removal tool"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("USAGE:"))
	fmt.Printf("  %s %s\n",
		commandStyle.Render("vx"),
		descStyle.Render("<file|directory> [file2] [dir2] ...        Remove files/directories safely"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("FILE OPERATIONS:"))
	fmt.Printf("  %s  %s\n", flagStyle.Render("vx <files...>"), descStyle.Render("Remove files/directories safely"))
	fmt.Printf("  %s, %s     %s\n", flagStyle.Render("-r"), flagStyle.Render("--restore <pattern>..."), descStyle.Render("Restore files matching patterns"))
	fmt.Printf("  %s, %s         %s\n", flagStyle.Render("-c"), flagStyle.Render("--clear"), descStyle.Render("Clear all cached files immediately"))
	fmt.Printf("  %s, %s        %s\n", flagStyle.Render("-pr"), flagStyle.Render("--purge <days>"), descStyle.Render("Delete files older than N days"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("INFORMATION:"))
	fmt.Printf("  %s, %s          %s\n", flagStyle.Render("-l"), flagStyle.Render("--list"), descStyle.Render("Show all cached files"))
	fmt.Printf("  %s, %s     %s\n", flagStyle.Render("-i"), flagStyle.Render("--info <pattern>"), descStyle.Render("Show detailed info about cached item(s)"))
	fmt.Printf("  %s, %s         %s\n", flagStyle.Render("-s"), flagStyle.Render("--stats"), descStyle.Render("Show cache statistics"))
	fmt.Printf("  %s, %s          %s\n", flagStyle.Render("-p"), flagStyle.Render("--path"), descStyle.Render("Print cache directory path"))
	fmt.Printf("  %s, %s    %s\n", flagStyle.Render("-cp"), flagStyle.Render("--config-path"), descStyle.Render("Print config file path"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("CUSTOMIZATION:"))
	fmt.Printf("  %s, %s        %s\n", flagStyle.Render("-t"), flagStyle.Render("--themes"), descStyle.Render("Interactive theme selector"))
	// fmt.Printf("  %s            %s\n", flagStyle.Render("--completion <shell>"), descStyle.Render("Generate shell completion (bash,zsh,fish,powershell)"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("OPTIONS:"))
	fmt.Printf("  %s, %s     %s\n", flagStyle.Render("-f"), flagStyle.Render("--noconfirm"), descStyle.Render("Skip confirmation prompts"))
	fmt.Printf("  %s, %s          %s\n", flagStyle.Render("-h"), flagStyle.Render("--help"), descStyle.Render("Show this help message"))
	fmt.Printf("  %s, %s        %s\n", flagStyle.Render("-v"), flagStyle.Render("--version"), descStyle.Render("Show version information"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("EXAMPLES:"))
	fmt.Println(exampleStyle.Render("Basic usage:"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx file1.txt"), descStyle.Render("# Delete file safely"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx file1.txt dir1/ *.log"), descStyle.Render("# Delete multiple items"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -f *.tmp"), descStyle.Render("# Delete without confirmation"))
	fmt.Println()

	fmt.Println(exampleStyle.Render("Recovery operations:"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -r file1.txt"), descStyle.Render("# Restore specific file"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -r \"*project*\""), descStyle.Render("# Restore files matching pattern"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -r \"*.pdf\" \"backup-*\""), descStyle.Render("# Restore multiple patterns"))
	fmt.Println()

	fmt.Println(exampleStyle.Render("Maintenance:"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -pr 30"), descStyle.Render("# Purge files older than 30 days"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -s"), descStyle.Render("# Show cache statistics"))
	fmt.Printf("  %s %s\n", commandStyle.Render("vx -c"), descStyle.Render("# Clear entire cache"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("CURRENT CONFIGURATION:"))
	fmt.Printf("  %s %s\n", configStyle.Render("Cache location:"), descStyle.Render(config.Cache.Directory))
	fmt.Printf("  %s %s\n", configStyle.Render("Retention period:"), descStyle.Render(fmt.Sprintf("%d days", config.Cache.Days)))
	fmt.Printf("  %s %s\n", configStyle.Render("Skip confirmations:"), descStyle.Render(fmt.Sprintf("%v", config.Cache.NoConfirm)))
	fmt.Printf("  %s %s\n", configStyle.Render("Current theme:"), descStyle.Render(config.UI.Theme))

	if config.Logging.Enabled {
		fmt.Printf("  %s %s\n", configStyle.Render("Logging:"), descStyle.Render("enabled → "+config.Logging.Directory))
	} else {
		fmt.Printf("  %s %s\n", configStyle.Render("Logging:"), descStyle.Render("disabled"))
	}

	// if config.Notifications.DesktopEnabled {
	// 	fmt.Printf("  %s %s\n", configStyle.Render("Notifications:"), descStyle.Render("enabled"))
	// } else {
	// 	fmt.Printf("  %s %s\n", configStyle.Render("Notifications:"), descStyle.Render("disabled"))
	// }
	// fmt.Println()

	// Show shell completion section if available
	// fmt.Println(sectionStyle.Render("SHELL COMPLETION:"))
	// completionStyle := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color(config.UI.Colors.Secondary)).
	// 	MarginLeft(2)

	// fmt.Println(completionStyle.Render("Setup tab completion for better productivity:"))
	// fmt.Printf("  %s %s\n", commandStyle.Render("vx --completion bash"), descStyle.Render("# Generate Bash completion"))
	// fmt.Printf("  %s %s\n", commandStyle.Render("vx --completion zsh"), descStyle.Render("# Generate Zsh completion"))
	// fmt.Printf("  %s %s\n", commandStyle.Render("vx --completion fish"), descStyle.Render("# Generate Fish completion"))
	// fmt.Println()
	fmt.Println(footerStyle.Render("For more information visit: https://github.com/Aelune/vanish"))
}

// ShowUsageFallback is a alternative fallback to simple output if colors are not available
func ShowUsageFallback(config types.Config) {
	fmt.Println("Vanish (vx) - Safe file/directory removal tool")
	fmt.Println("=" + strings.Repeat("=", 47))
	fmt.Println()

	fmt.Println("USAGE:")
	fmt.Println("  vx <file|directory> [file2] [dir2] ...        Remove files/directories safely")
	fmt.Println()

	fmt.Println("FILE OPERATIONS:")
	fmt.Println("  vx <files...>                                 Remove files/directories safely")
	fmt.Println("  -r, --restore <pattern>...                   Restore files matching patterns")
	fmt.Println("  -c, --clear                                   Clear all cached files immediately")
	fmt.Println("  -pr, --purge <days>                           Delete files older than N days")
	fmt.Println()

	fmt.Println("INFORMATION:")
	fmt.Println("  -l, --list                                    Show all cached files")
	fmt.Println("  -i, --info <pattern>                          Show detailed info about cached item(s)")
	fmt.Println("  -s, --stats                                   Show cache statistics")
	fmt.Println("  -p, --path                                    Print cache directory path")
	fmt.Println("  -cp, --config-path                            Print config file path")
	fmt.Println()

	fmt.Println("CUSTOMIZATION:")
	fmt.Println("  -t, --themes                                  Interactive theme selector")
	// fmt.Println("  --completion <shell>                          Generate shell completion")
	fmt.Println()

	fmt.Println("OPTIONS:")
	fmt.Println("  -f, --noconfirm                               Skip confirmation prompts")
	fmt.Println("  -h, --help                                    Show this help message")
	fmt.Println("  -v, --version                                 Show version information")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  vx file1.txt                                  # Delete file safely")
	fmt.Println("  vx -r \"*project*\"                             # Restore files matching pattern")
	fmt.Println("  vx -pr 30                                     # Purge files older than 30 days")
	fmt.Println()

	fmt.Println("CURRENT CONFIGURATION:")
	fmt.Printf("  Cache location: %s\n", config.Cache.Directory)
	fmt.Printf("  Retention period: %d days\n", config.Cache.Days)
	fmt.Printf("  Current theme: %s\n", config.UI.Theme)
	fmt.Printf("  Logging: %v\n", config.Logging.Enabled)
	// fmt.Printf("  Notifications: %v\n", config.Notifications.DesktopEnabled)
	fmt.Println("\n\nFor more information visit: https://github.com/Aelune/vanish")
}
