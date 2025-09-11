package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"vanish/internal/config"
	"vanish/internal/models"
)

// ThemeStyles holds all the styled components
type ThemeStyles struct {
	Title       lipgloss.Style
	Header      lipgloss.Style
	Question    lipgloss.Style
	Filename    lipgloss.Style
	Success     lipgloss.Style
	Error       lipgloss.Style
	Warning     lipgloss.Style
	Info        lipgloss.Style
	Help        lipgloss.Style
	Progress    lipgloss.Style
	Border      lipgloss.Style
	Background  lipgloss.Style
	List        lipgloss.Style
	StatusGood  lipgloss.Style
	StatusBad   lipgloss.Style
}

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return width
}

// Convert RGB hex colors to ANSI 256-color codes for better compatibility
func convertColorForTerminal(hexColor string) string {
	// Map of common hex colors to ANSI 256-color codes
	colorMap := map[string]string{
		"#FF0000": "196", // Red
		"#00FF00": "46",  // Green
		"#0000FF": "21",  // Blue
		"#FFFF00": "226", // Yellow
		"#FF00FF": "201", // Magenta
		"#00FFFF": "51",  // Cyan
		"#FFFFFF": "15",  // White
		"#000000": "0",   // Black
		"#808080": "244", // Gray
		"#FFA500": "214", // Orange
		"#800080": "129", // Purple
		"#008000": "28",  // Dark Green
		"#000080": "18",  // Dark Blue
		"#800000": "88",  // Dark Red
		// Add colors from your config themes
		"#ef4444": "203", // Red variant
		"#22c55e": "46",  // Green variant
		"#3b82f6": "75",  // Blue variant
		"#f59e0b": "214", // Amber/Orange
		"#8b5cf6": "135", // Purple/Violet
		"#06b6d4": "87",  // Cyan variant
		"#6b7280": "244", // Gray-500
		"#9ca3af": "249", // Gray-400
		"#d1d5db": "252", // Gray-300
		"#f3f4f6": "255", // Gray-100
	}

	// Check if it's a hex color that needs conversion
	if strings.HasPrefix(hexColor, "#") {
		if ansiColor, exists := colorMap[strings.ToUpper(hexColor)]; exists {
			return ansiColor
		}
		// If not in our map, try to convert to a close ANSI color
		return convertHexToAnsi256(hexColor)
	}

	// If it's already an ANSI color code, return as-is
	return hexColor
}

// Simple hex to ANSI 256 conversion
func convertHexToAnsi256(hex string) string {
	// Fallback color mappings based on common patterns
	hex = strings.ToUpper(hex)
	switch {
	case strings.Contains(hex, "FF") && strings.Contains(hex, "00"):
		if strings.HasSuffix(hex, "0000") {
			return "196" // Red
		} else if strings.HasPrefix(hex, "#00FF") {
			return "46" // Green
		} else if strings.Contains(hex, "00FF") {
			return "21" // Blue
		}
	case strings.Contains(hex, "80"):
		return "244" // Gray
	case strings.Contains(hex, "FF"):
		return "226" // Yellow/Bright
	}
	return "15" // Default to white
}

func CreateThemeStyles(cfg config.Config) ThemeStyles {
	// Always create styles, but use terminal-compatible colors
	termWidth := GetTerminalWidth()
	contentWidth := termWidth - (cfg.UI.PaddingX * 2)

	// Convert all theme colors to terminal-compatible versions
	colors := struct {
		Primary    string
		Secondary  string
		Success    string
		Error      string
		Warning    string
		Highlight  string
		Muted      string
		Text       string
		Border     string
	}{
		Primary:    convertColorForTerminal(cfg.UI.Colors.Primary),
		Secondary:  convertColorForTerminal(cfg.UI.Colors.Secondary),
		Success:    convertColorForTerminal(cfg.UI.Colors.Success),
		Error:      convertColorForTerminal(cfg.UI.Colors.Error),
		Warning:    convertColorForTerminal(cfg.UI.Colors.Warning),
		Highlight:  convertColorForTerminal(cfg.UI.Colors.Highlight),
		Muted:      convertColorForTerminal(cfg.UI.Colors.Muted),
		Text:       convertColorForTerminal(cfg.UI.Colors.Text),
		Border:     convertColorForTerminal(cfg.UI.Colors.Border),
	}

	// Use simpler border styles that work better across terminals
	simpleBorder := lipgloss.Border{
		Top:    "─",
		Bottom: "─",
		Left:   "│",
		Right:  "│",
	}

	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Text))

	if cfg.UI.Compact {
		return ThemeStyles{
			Root: baseStyle.Width(termWidth).Padding(0, cfg.UI.PaddingX),

			Header: baseStyle.
				Foreground(lipgloss.Color(colors.Primary)).
				Bold(true).
				Width(contentWidth),

			Question: baseStyle.
				Foreground(lipgloss.Color(colors.Secondary)).
				Bold(true).
				Width(contentWidth),

			Filename: baseStyle.
				Foreground(lipgloss.Color(colors.Highlight)).
				Bold(true),

			Info: baseStyle.
				Foreground(lipgloss.Color(colors.Muted)).
				Width(contentWidth),

			Warning: baseStyle.
				Foreground(lipgloss.Color(colors.Warning)).
				Bold(true).
				Width(contentWidth),

			Error: baseStyle.
				Foreground(lipgloss.Color(colors.Error)).
				Bold(true).
				Border(simpleBorder).
				Width(contentWidth - 2).
				Align(lipgloss.Center),

			Success: baseStyle.
				Foreground(lipgloss.Color(colors.Success)).
				Bold(true).
				Border(simpleBorder).
				Width(contentWidth - 2).
				Align(lipgloss.Center),

			List: baseStyle.
				MarginLeft(1).
				Border(simpleBorder, false, false, false, true).
				BorderForeground(lipgloss.Color(colors.Border)).
				Width(contentWidth - 2),

			Help: baseStyle.
				Foreground(lipgloss.Color(colors.Highlight)).
				Italic(true).
				Width(contentWidth),

			Progress: baseStyle.
				Foreground(lipgloss.Color(colors.Primary)).
				Bold(true).
				Width(contentWidth),

			StatusBad: baseStyle.
				Foreground(lipgloss.Color(colors.Error)).
				Italic(true),

			Compact: baseStyle,
		}
	}

	// Standard styles
	return ThemeStyles{
		Root: baseStyle.
			Width(termWidth).
			Padding(cfg.UI.PaddingY, cfg.UI.PaddingX),

		Header: baseStyle.
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true).
			Underline(true).
			PaddingBottom(1).
			Width(contentWidth),

		Question: baseStyle.
			Foreground(lipgloss.Color(colors.Secondary)).
			Bold(true).
			MarginBottom(1).
			Width(contentWidth),

		Filename: baseStyle.
			Foreground(lipgloss.Color(colors.Highlight)).
			Bold(true),

		Info: baseStyle.
			Foreground(lipgloss.Color(colors.Muted)).
			Width(contentWidth),

		Warning: baseStyle.
			Foreground(lipgloss.Color(colors.Warning)).
			Bold(true).
			Italic(true).
			Width(contentWidth),

		Error: baseStyle.
			Foreground(lipgloss.Color(colors.Error)).
			Bold(true).
			Border(simpleBorder).
			Padding(0, 1).
			Width(contentWidth - 4).
			Align(lipgloss.Center),

		Success: baseStyle.
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true).
			Border(simpleBorder).
			Padding(0, 1).
			Width(contentWidth - 4).
			Align(lipgloss.Center),

		List: baseStyle.
			MarginLeft(2).
			Border(simpleBorder, false, false, false, true).
			BorderForeground(lipgloss.Color(colors.Border)).
			Width(contentWidth - 3),

		Help: baseStyle.
			Foreground(lipgloss.Color(colors.Highlight)).
			Italic(true).
			MarginTop(1).
			Width(contentWidth),

		Progress: baseStyle.
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true).
			Width(contentWidth),

		StatusBad: baseStyle.
			Foreground(lipgloss.Color(colors.Error)).
			Italic(true).
			Underline(true),

		Compact: baseStyle,
	}
}

func setupProgress(config Config) progress.Model {
	prog := progress.New()
	prog.Width = 50

	switch config.UI.Progress.Style {
	case "solid":
		prog = progress.New(progress.WithSolidFill(config.UI.Colors.Primary))
	case "rainbow":
		prog = progress.New(progress.WithGradient("#FF0000", "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#4B0082", "#9400D3"))
	default: // gradient
		prog = progress.New(progress.WithGradient(config.UI.Colors.Primary, config.UI.Colors.Secondary))
	}

	return prog
}


// RenderList renders a list with proper  and padding
func RenderList(content string, style lipgloss.Style, width int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	var renderedLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineWidth := lipgloss.Width(line)
		targetWidth := width - 3
		if lineWidth < targetWidth {
			paddedLine := line + strings.Repeat(" ", targetWidth-lineWidth)
			renderedLines = append(renderedLines, paddedLine)
		} else {
			renderedLines = append(renderedLines, line)
		}
	}

	return style.Render(strings.Join(renderedLines, "\n"))
}

// PadToWidth pads text to a specific width
func PadToWidth(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth < width {
		return text + strings.Repeat(" ", width-textWidth)
	}
	return text
}

// ShowThemes displays available themes with visual previews
func ShowThemes() {
	themes := config.GetDefaultThemes()

	fmt.Println("Available Vanish Themes")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	themeOrder := []string{"default", "dark", "light", "cyberpunk", "minimal"}

	for _, name := range themeOrder {
		if theme, exists := themes[name]; exists {
			displayThemePreview(name, theme)
			fmt.Println()
		}
	}

	fmt.Println("Usage:")
	fmt.Println("  Set theme in your config file: ~/.config/vanish/vanish.toml")
	fmt.Println("  [ui]")
	fmt.Println("  theme = \"dark\"  # or any theme name above")
	fmt.Println()
	fmt.Println("  You can also override individual colors in [ui.colors] section.")
}

func displayThemePreview(name string, theme config.Config) {
	// Convert colors to terminal-compatible versions
	colors := map[string]string{
		"primary":   convertColorForTerminal(theme.UI.Colors.Primary),
		"success":   convertColorForTerminal(theme.UI.Colors.Success),
		"warning":   convertColorForTerminal(theme.UI.Colors.Warning),
		"error":     convertColorForTerminal(theme.UI.Colors.Error),
		"highlight": convertColorForTerminal(theme.UI.Colors.Highlight),
		"muted":     convertColorForTerminal(theme.UI.Colors.Muted),
	}

	// Create styles for this theme
	primaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["primary"])).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["success"])).Bold(true)
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["warning"])).Bold(true)
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["error"])).Bold(true)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["highlight"])).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors["muted"]))

	// Theme header
	fmt.Printf("┌─ %s ─┐\n", primaryStyle.Render(strings.ToUpper(name)))

	// Color preview line
	colorLine := fmt.Sprintf("  %s %s %s %s %s",
		primaryStyle.Render("●"),
		successStyle.Render("●"),
		warningStyle.Render("●"),
		errorStyle.Render("●"),
		highlightStyle.Render("●"))
	fmt.Println(colorLine)

	// Sample content preview
	fmt.Printf("  %s %s\n",
		highlightStyle.Render("file.txt"),
		mutedStyle.Render("→ cached"))
	fmt.Printf("  %s %s\n",
		successStyle.Render("✓"),
		mutedStyle.Render("Operation completed"))

	// Theme details
	fmt.Printf("  %s\n", mutedStyle.Render(fmt.Sprintf(
		"Progress: %s | Emoji: %v | Animation: %v",
		theme.UI.Progress.Style,
		theme.UI.Progress.ShowEmoji,
		theme.UI.Progress.Animation)))
}

// DiagnoseTerminal prints terminal capability information
func DiagnoseTerminal() {
	fmt.Println("Terminal Diagnostics")
	fmt.Println(strings.Repeat("=", 40))

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	fmt.Printf("TTY Detection: %v\n", isTTY)
	fmt.Printf("TERM: %s\n", os.Getenv("TERM"))
	fmt.Printf("COLORTERM: %s\n", os.Getenv("COLORTERM"))

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w, h = 80, 24
	}
	fmt.Printf("Terminal Size: %dx%d\n", w, h)

	fmt.Println("\nColor Test (ANSI 256):")
	colors := []string{"1", "2", "3", "4", "5", "6", "46", "196", "226", "87"}
	for i, color := range colors {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		fmt.Printf("%s ", style.Render("●"))
		if i == len(colors)-1 {
			fmt.Println()
		}
	}

	fmt.Println("\nTesting problematic RGB conversion:")
	testColors := []string{"#ef4444", "#22c55e", "#3b82f6"}
	for _, hexColor := range testColors {
		ansiColor := convertColorForTerminal(hexColor)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(ansiColor))
		fmt.Printf("%s -> %s %s\n", hexColor, ansiColor, style.Render("●"))
	}
}
