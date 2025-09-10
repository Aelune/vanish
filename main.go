package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Config represents the TOML configuration
type Config struct {
	Cache struct {
		Directory string `toml:"directory"`
		Days      int    `toml:"days"`
	} `toml:"cache"`
	Logging struct {
		Enabled   bool   `toml:"enabled"`
		Directory string `toml:"directory"`
	} `toml:"logging"`
	UI struct {
		Theme string `toml:"theme"` // "default", "dark", "light", "cyberpunk", "minimal"
		Colors struct {
			Primary     string `toml:"primary"`
			Secondary   string `toml:"secondary"`
			Success     string `toml:"success"`
			Warning     string `toml:"warning"`
			Error       string `toml:"error"`
			Text        string `toml:"text"`
			Muted       string `toml:"muted"`
			Border      string `toml:"border"`
			Highlight   string `toml:"highlight"`
			// Background string `toml:"background"`
		} `toml:"colors"`
		Progress struct {
			Style      string `toml:"style"` // "gradient", "solid", "rainbow"
			ShowEmoji  bool   `toml:"show_emoji"`
			Animation  bool   `toml:"animation"`
		} `toml:"progress"`
		PaddingX int `toml:"padding_x"`
		PaddingY int `toml:"padding_y"`
	} `toml:"ui"`
}

// DeletedItem represents an item in the global index
type DeletedItem struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	DeleteDate   time.Time `json:"delete_date"`
	CachePath    string    `json:"cache_path"`
	IsDirectory  bool      `json:"is_directory"`
	FileCount    int       `json:"file_count,omitempty"`
	Size         int64     `json:"size"`
}

// Index represents the global index file
type Index struct {
	Items []DeletedItem `json:"items"`
}

// FileInfo holds information about a file to be deleted
type FileInfo struct {
	Path        string
	IsDirectory bool
	FileCount   int
	Exists      bool
	Error       string
}

// ThemeStyles holds all the styled components
type ThemeStyles struct {
	Root     lipgloss.Style
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
	List        lipgloss.Style
	StatusGood  lipgloss.Style
	StatusBad   lipgloss.Style
}

// Model represents the application state
type Model struct {
	filenames      []string
	fileInfos      []FileInfo
	currentIndex   int
	state          string
	progress       progress.Model
	progressVal    float64
	confirmed      bool
	errorMsg       string
	config         Config
	styles         ThemeStyles
	processedItems []DeletedItem
	clearAll       bool
	totalFiles     int
	processedFiles int
}

type filesExistMsg struct {
	fileInfos []FileInfo
}

type fileMoveMsg struct {
	item DeletedItem
	err  error
}

type cleanupMsg struct{}

type clearMsg struct {
	err error
}

type errorMsg string


func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // fallback width
	}
	return width
}

func getDefaultThemes() map[string]Config {
	themes := make(map[string]Config)

	// Default theme
	defaultTheme := Config{}
	defaultTheme.UI.Theme = "default"
	defaultTheme.UI.Colors.Primary = "#3B82F6"
	defaultTheme.UI.Colors.Secondary = "#6366F1"
	defaultTheme.UI.Colors.Success = "#10B981"
	defaultTheme.UI.Colors.Warning = "#F59E0B"
	defaultTheme.UI.Colors.Error = "#EF4444"
	defaultTheme.UI.Colors.Text = "#F9FAFB"
	defaultTheme.UI.Colors.Muted = "#9CA3AF"
	defaultTheme.UI.Colors.Border = "#374151"
	defaultTheme.UI.Colors.Highlight = "#FBBF24"
	defaultTheme.UI.Progress.Style = "gradient"
	defaultTheme.UI.Progress.ShowEmoji = true
	defaultTheme.UI.Progress.Animation = true
	themes["default"] = defaultTheme

	// Dark theme
	darkTheme := Config{}
	darkTheme.UI.Theme = "dark"
	darkTheme.UI.Colors.Primary = "#8B5CF6"
	darkTheme.UI.Colors.Secondary = "#A78BFA"
	darkTheme.UI.Colors.Success = "#34D399"
	darkTheme.UI.Colors.Warning = "#FBBF24"
	darkTheme.UI.Colors.Error = "#F87171"
	darkTheme.UI.Colors.Text = "#E2E8F0"
	darkTheme.UI.Colors.Muted = "#64748B"
	darkTheme.UI.Colors.Border = "#1E293B"
	darkTheme.UI.Colors.Highlight = "#FDE047"
	darkTheme.UI.Progress.Style = "gradient"
	darkTheme.UI.Progress.ShowEmoji = true
	darkTheme.UI.Progress.Animation = true
	themes["dark"] = darkTheme

	// Light theme
	lightTheme := Config{}
	lightTheme.UI.Theme = "light"
	lightTheme.UI.Colors.Primary = "#2563EB"
	lightTheme.UI.Colors.Secondary = "#4F46E5"
	lightTheme.UI.Colors.Success = "#059669"
	lightTheme.UI.Colors.Warning = "#D97706"
	lightTheme.UI.Colors.Error = "#DC2626"
	lightTheme.UI.Colors.Text = "#1F2937"
	lightTheme.UI.Colors.Muted = "#6B7280"
	lightTheme.UI.Colors.Border = "#E5E7EB"
	lightTheme.UI.Colors.Highlight = "#F59E0B"
	lightTheme.UI.Progress.Style = "solid"
	lightTheme.UI.Progress.ShowEmoji = true
	lightTheme.UI.Progress.Animation = false
	themes["light"] = lightTheme

	// Cyberpunk theme
	cyberpunkTheme := Config{}
	cyberpunkTheme.UI.Theme = "cyberpunk"
	cyberpunkTheme.UI.Colors.Primary = "#00FFFF"
	cyberpunkTheme.UI.Colors.Secondary = "#FF00FF"
	cyberpunkTheme.UI.Colors.Success = "#00FF00"
	cyberpunkTheme.UI.Colors.Warning = "#FFFF00"
	cyberpunkTheme.UI.Colors.Error = "#FF0040"
	cyberpunkTheme.UI.Colors.Text = "#00FFFF"
	cyberpunkTheme.UI.Colors.Muted = "#8A2BE2"
	cyberpunkTheme.UI.Colors.Border = "#FF00FF"
	cyberpunkTheme.UI.Colors.Highlight = "#FFFF00"
	cyberpunkTheme.UI.Progress.Style = "rainbow"
	cyberpunkTheme.UI.Progress.ShowEmoji = false
	cyberpunkTheme.UI.Progress.Animation = true
	themes["cyberpunk"] = cyberpunkTheme

	// Minimal theme
	minimalTheme := Config{}
	minimalTheme.UI.Theme = "minimal"
	minimalTheme.UI.Colors.Primary = "#000000"
	minimalTheme.UI.Colors.Secondary = "#404040"
	minimalTheme.UI.Colors.Success = "#008000"
	minimalTheme.UI.Colors.Warning = "#FFA500"
	minimalTheme.UI.Colors.Error = "#FF0000"
	minimalTheme.UI.Colors.Text = "#000000"
	minimalTheme.UI.Colors.Muted = "#808080"
	minimalTheme.UI.Colors.Border = "#C0C0C0"
	minimalTheme.UI.Colors.Highlight = "#0000FF"
	minimalTheme.UI.Progress.Style = "solid"
	minimalTheme.UI.Progress.ShowEmoji = false
	minimalTheme.UI.Progress.Animation = false
	themes["minimal"] = minimalTheme

	return themes
}

func createThemeStyles(cfg Config) ThemeStyles {
	border := lipgloss.NormalBorder()
	rounded := lipgloss.RoundedBorder()
	dashes := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	termWidth := getTerminalWidth()
	contentWidth := termWidth - (cfg.UI.PaddingX * 2) // Account for root padding

	return ThemeStyles{
		Root: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Text)).
			// Background(lipgloss.Color(cfg.UI.Colors.Background)).
			Width(termWidth).
			Padding(cfg.UI.PaddingY, cfg.UI.PaddingX),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Primary)).
			Bold(true).
			Underline(true).
			PaddingBottom(1).
			Width(contentWidth),

		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Secondary)).
			Bold(true).
			MarginBottom(1).
			Width(contentWidth),

		Filename: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Highlight)).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Muted)).
			Width(contentWidth),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Warning)).
			Bold(true).
			Italic(true).
			Width(contentWidth),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Error)).
			Bold(true).
			Border(dashes, true).
			Padding(0, 1).
			Width(contentWidth - 4). // Account for border and padding
			Align(lipgloss.Center),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Success)).
			Bold(true).
			Border(rounded, true).
			Padding(0, 1).
			Width(contentWidth - 4). // Account for border and padding
			Align(lipgloss.Center),

		List: lipgloss.NewStyle().
			MarginLeft(2).
			Border(border, false, false, false, true).
			BorderForeground(lipgloss.Color(cfg.UI.Colors.Border)).
			Width(contentWidth - 3), // Account for margin and border
			// Background(lipgloss.Color(cfg.UI.Colors.Background)), // Ensure background color

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Highlight)).
			Italic(true).
			MarginTop(1).
			Width(contentWidth),

		Progress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Primary)).
			Bold(true).
			Width(contentWidth),

		StatusBad: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.UI.Colors.Error)).
			Italic(true).
			Underline(true),
	}
}

// Helper function to render list
func renderList(content string, style lipgloss.Style, width int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	var renderedLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			// Skip completely empty lines to avoid extra background
			continue
		} else {
			// Pad each line to the full width minus border and margin
			lineWidth := lipgloss.Width(line)
			targetWidth := width - 3 // Account for margin (2) and border (1)
			if lineWidth < targetWidth {
				paddedLine := line + strings.Repeat(" ", targetWidth-lineWidth)
				renderedLines = append(renderedLines, paddedLine)
			} else {
				renderedLines = append(renderedLines, line)
			}
		}
	}

	// Join without adding extra newlines at the end
	return style.Render(strings.Join(renderedLines, "\n"))
}

func loadConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(homeDir, ".config", "saferm", "saferm.toml")

	// Default configuration
	config := Config{}
	config.Cache.Directory = filepath.Join(homeDir, ".cache", "saferm")
	config.Cache.Days = 10
	config.Logging.Enabled = true
	config.Logging.Directory = filepath.Join(homeDir, ".cache", "saferm", "logs")

	// Apply default theme
	themes := getDefaultThemes()
	defaultTheme := themes["default"]
	config.UI = defaultTheme.UI

	// Try to load config file
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return config, fmt.Errorf("error parsing config file: %v", err)
		}

		// If a theme is specified, apply it but preserve any custom color overrides
		if config.UI.Theme != "" && config.UI.Theme != "default" {
			if themeConfig, exists := themes[config.UI.Theme]; exists {
				// Save current custom colors
				customColors := config.UI.Colors
				// Apply theme
				config.UI = themeConfig.UI
				// Restore any custom colors that were explicitly set
				if customColors.Primary != "" {
					config.UI.Colors.Primary = customColors.Primary
				}
				// ... repeat for other colors as needed
			}
		}
	} else {
		// Create default config file
		if err := createDefaultConfig(configPath, config); err != nil {
			log.Printf("Warning: Could not create default config: %v", err)
		}
	}

	return config, nil
}

func createDefaultConfig(configPath string, config Config) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configContent := `[cache]
# Directory where deleted files are stored (relative to home directory)
directory = ".cache/saferm"
# Number of days to keep deleted files
days = 10

[logging]
# Enable logging
enabled = true
# Directory for log files (relative to cache directory)
directory = ".cache/saferm/logs"

[ui]
# Theme: "default", "dark", "light", "cyberpunk", "minimal"
theme = "default"

[ui.colors]
# Color customization (hex colors)
primary = "#3B82F6"      # Main accent color
secondary = "#6366F1"    # Secondary accent
success = "#10B981"      # Success messages
warning = "#F59E0B"      # Warning messages
error = "#EF4444"        # Error messages
text = "#F9FAFB"         # Main text color
muted = "#9CA3AF"        # Muted/help text
border = "#374151"       # Border color
highlight = "#FBBF24"    # Filename highlight

[ui.progress]
# Progress bar settings
style = "gradient"       # "gradient", "solid", "rainbow"
show_emoji = true        # Show emoji in progress messages
animation = true         # Enable progress animations
`

	return os.WriteFile(configPath, []byte(configContent), 0644)
}

func setupProgress(config Config) progress.Model {
	width := 50
	primary := config.UI.Colors.Primary
	secondary := config.UI.Colors.Secondary

	var prog progress.Model

	switch config.UI.Progress.Style {
	case "solid":
		prog = progress.New(
			progress.WithSolidFill(primary),
		)

	case "rainbow":
		// Fake a rainbow using red → violet
		// Note: only 2 colors allowed in WithGradient
		prog = progress.New(
			progress.WithGradient("#FF0000", "#8B00FF"),
		)

	case "striped":
		// Alternate gradient using user colors
		prog = progress.New(
			progress.WithGradient(primary, secondary),
		)

	case "simple":
		// Use bubbletea's built-in default gradient
		prog = progress.New(
			progress.WithDefaultGradient(),
		)

	default:
		// Fallback to gradient with user-defined primary → secondary
		prog = progress.New(
			progress.WithGradient(primary, secondary),
		)
	}

	prog.Width = width
	return prog
}

func initialModel(filenames []string, clearAll bool) (Model, error) {
	config, err := loadConfig()
	if err != nil {
		return Model{}, err
	}

	prog := setupProgress(config)
	styles := createThemeStyles(config)

	return Model{
		filenames:      filenames,
		fileInfos:      make([]FileInfo, len(filenames)),
		state:          "checking",
		progress:       prog,
		config:         config,
		styles:         styles,
		clearAll:       clearAll,
		processedItems: make([]DeletedItem, 0),
		totalFiles:     len(filenames),
	}, nil
}

func (m Model) Init() tea.Cmd {
	if m.clearAll {
		m.state = "clearing"
		return tea.Batch(
			m.progress.SetPercent(0.1),
			clearAllCache(m.config),
		)
	}

	return tea.Batch(
		checkFilesExist(m.filenames),
		m.progress.SetPercent(0.1),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y", "Y":
			if m.state == "confirming" {
				m.confirmed = true
				m.state = "moving"
				m.currentIndex = 0
				return m, tea.Batch(
					m.progress.SetPercent(0.3),
					moveNextFileToCache(m.fileInfos, m.currentIndex, m.config),
				)
			}
		case "n", "N":
			if m.state == "confirming" {
				return m, tea.Quit
			}
		case "enter":
			if m.state == "done" || m.state == "error" {
				return m, tea.Quit
			}
		}

	case filesExistMsg:
		m.fileInfos = msg.fileInfos
		validFiles := 0
		for _, info := range m.fileInfos {
			if info.Exists {
				validFiles++
			}
		}

		if validFiles == 0 {
			m.state = "error"
			m.errorMsg = "No valid files or directories found"
			return m, nil
		}

		m.state = "confirming"
		return m, m.progress.SetPercent(0.2)

	case fileMoveMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error moving file: %v", msg.err)
			return m, nil
		}

		if msg.item.ID != "" {
			m.processedItems = append(m.processedItems, msg.item)
			m.processedFiles++
		}

		m.currentIndex++

		// Update progress
		progressPercent := 0.3 + (float64(m.currentIndex)/float64(len(m.fileInfos)))*0.4

		// Check if we have more files to process
		nextIndex := findNextValidFile(m.fileInfos, m.currentIndex)
		if nextIndex != -1 {
			m.currentIndex = nextIndex
			return m, tea.Batch(
				m.progress.SetPercent(progressPercent),
				moveNextFileToCache(m.fileInfos, m.currentIndex, m.config),
			)
		}

		// All files processed, move to cleanup
		m.state = "cleanup"
		return m, tea.Batch(
			m.progress.SetPercent(0.7),
			cleanupOldFiles(m.config),
		)

	case cleanupMsg:
		m.state = "done"
		return m, m.progress.SetPercent(1.0)

	case clearMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error clearing cache: %v", msg.err)
			return m, nil
		}
		m.state = "done"
		return m, m.progress.SetPercent(1.0)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case errorMsg:
		m.state = "error"
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var content strings.Builder
	termWidth := getTerminalWidth()
	contentWidth := termWidth - (m.config.UI.PaddingX * 2)

	switch m.state {
	case "checking":
		if m.config.UI.Progress.ShowEmoji {
			content.WriteString("🔍 ")
		}
		content.WriteString("Checking files and directories...\n\n")
		content.WriteString(m.styles.Progress.Render(m.progress.View()))

	case "confirming":
		content.WriteString(m.styles.Question.Render("Are you sure you want to delete the following items?"))
		content.WriteString("\n")

		validCount := 0
		invalidCount := 0
		totalFileCount := 0

		listContent := strings.Builder{}
		for _, info := range m.fileInfos {
			if info.Exists {
				validCount++
				if info.IsDirectory {
					if m.config.UI.Progress.ShowEmoji {
						listContent.WriteString("📁 ")
					} else {
						listContent.WriteString("DIR ")
					}
					listContent.WriteString(m.styles.Filename.Render(info.Path))
					if info.FileCount > 0 {
						listContent.WriteString(m.styles.Info.Render(fmt.Sprintf(" (%d items)", info.FileCount)))
						totalFileCount += info.FileCount
					} else {
						listContent.WriteString(m.styles.Info.Render(" (empty)"))
					}
				} else {
					if m.config.UI.Progress.ShowEmoji {
						listContent.WriteString("📄 ")
					} else {
						listContent.WriteString("FILE ")
					}
					listContent.WriteString(m.styles.Filename.Render(info.Path))
					totalFileCount++
				}
				listContent.WriteString("\n")
			} else {
				invalidCount++
				if m.config.UI.Progress.ShowEmoji {
					listContent.WriteString("❌ ")
				} else {
					listContent.WriteString("ERR ")
				}
				listContent.WriteString(m.styles.StatusBad.Render(info.Path))
				listContent.WriteString(m.styles.Warning.Render(" (does not exist)"))
				listContent.WriteString("\n")
			}
		}

		// Use the new helper function to render the list with proper background
		content.WriteString(renderList(listContent.String(), m.styles.List, contentWidth))

		if invalidCount > 0 {
			warningText := fmt.Sprintf("⚠ Warning: %d file(s) will be skipped", invalidCount)
			paddedWarning := padToWidth(warningText, contentWidth)
			content.WriteString(m.styles.Warning.Render(paddedWarning))
			content.WriteString("\n")
		}

		if validCount > 0 {
			infoText := fmt.Sprintf("Total items to delete: %d", validCount)
			if totalFileCount > validCount {
				infoText += fmt.Sprintf(" | Files affected: %d", totalFileCount)
			}
			infoText += fmt.Sprintf(" | Recoverable for %d days", m.config.Cache.Days)

			paddedInfo := padToWidth(infoText, contentWidth)
			content.WriteString(m.styles.Info.Render(paddedInfo))
			content.WriteString("\n\n")

			helpText := padToWidth("Press 'y' to confirm, 'n' to cancel, or 'q' to quit", contentWidth)
			content.WriteString(m.styles.Help.Render(helpText))
		} else {
			content.WriteString("\n")
			helpText := padToWidth("Press 'n' to cancel or 'q' to quit", contentWidth)
			content.WriteString(m.styles.Help.Render(helpText))
		}

	case "moving":
		if m.currentIndex < len(m.fileInfos) {
			currentFile := m.fileInfos[m.currentIndex]
			emoji := ""
			fileType := "file"
			if currentFile.IsDirectory {
				fileType = "directory"
				if m.config.UI.Progress.ShowEmoji {
					emoji = "📦 "
				}
			} else {
				if m.config.UI.Progress.ShowEmoji {
					emoji = "📦 "
				}
			}

			statusText := fmt.Sprintf("%sMoving %s '%s' to safe cache... (%d/%d)",
				emoji, fileType, currentFile.Path, m.processedFiles+1, m.totalFiles)
			paddedStatus := padToWidth(statusText, contentWidth)
			content.WriteString(m.styles.Info.Render(paddedStatus))
		} else {
			statusText := "Moving files to safe cache..."
			if m.config.UI.Progress.ShowEmoji {
				statusText = "📦 " + statusText
			}
			paddedStatus := padToWidth(statusText, contentWidth)
			content.WriteString(m.styles.Info.Render(paddedStatus))
		}
		content.WriteString("\n\n")
		content.WriteString(m.styles.Progress.Render(m.progress.View()))

	case "cleanup":
		statusText := "Cleaning up old cached files..."
		if m.config.UI.Progress.ShowEmoji {
			statusText = "🧹 " + statusText
		}
		paddedStatus := padToWidth(statusText, contentWidth)
		content.WriteString(m.styles.Info.Render(paddedStatus))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Progress.Render(m.progress.View()))

	case "clearing":
		statusText := "Clearing all cached files..."
		if m.config.UI.Progress.ShowEmoji {
			statusText = "🗑️ " + statusText
		}
		paddedStatus := padToWidth(statusText, contentWidth)
		content.WriteString(m.styles.Info.Render(paddedStatus))
		content.WriteString("\n\n")
		content.WriteString(m.styles.Progress.Render(m.progress.View()))

	case "done":
		successMsg := ""
		if m.clearAll {
			if m.config.UI.Progress.ShowEmoji {
				successMsg = "✅ All cached files cleared!"
			} else {
				successMsg = "SUCCESS: All cached files cleared!"
			}
		} else {
			emoji := ""
			if m.config.UI.Progress.ShowEmoji {
				emoji = "✅ "
			}
			successMsg = fmt.Sprintf("%sSuccessfully processed %d item(s)!", emoji, len(m.processedItems))
		}

		content.WriteString(m.styles.Success.Render(successMsg))
		content.WriteString("\n\n")

		if !m.clearAll && len(m.processedItems) > 0 {
			detailsBuilder := strings.Builder{}
			for i, item := range m.processedItems {
				if i < 5 {
					detailsBuilder.WriteString(fmt.Sprintf("• %s → %s\n",
						m.styles.Filename.Render(item.OriginalPath),
						filepath.Base(item.CachePath)))
				}
			}

			if len(m.processedItems) > 5 {
				detailsBuilder.WriteString(m.styles.Info.Render(fmt.Sprintf("... and %d more item(s)", len(m.processedItems)-5)))
				detailsBuilder.WriteString("\n")
			}

			// Use the helper function for the details list too
			content.WriteString(renderList(detailsBuilder.String(), m.styles.List, contentWidth))

			if len(m.processedItems) > 0 {
				deleteAfter := m.processedItems[0].DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
				deleteAfterText := fmt.Sprintf("Will be permanently deleted after: %s", deleteAfter.Format("2006-01-02 15:04:05"))
				paddedDeleteAfter := padToWidth(deleteAfterText, contentWidth)
				content.WriteString(m.styles.Info.Render(paddedDeleteAfter))
				content.WriteString("\n\n")
			}
		}

		content.WriteString(m.styles.Progress.Render(m.progress.View()))
		content.WriteString("\n\n")
		helpText := padToWidth("Press Enter or 'q' to exit", contentWidth)
		content.WriteString(m.styles.Help.Render(helpText))

	case "error":
		errorMsg := ""
		if m.config.UI.Progress.ShowEmoji {
			errorMsg = "❌ Error"
		} else {
			errorMsg = "ERROR"
		}

		content.WriteString(m.styles.Error.Render(errorMsg))
		content.WriteString("\n\n")

		errorText := padToWidth(m.errorMsg, contentWidth)
		content.WriteString(m.styles.Info.Render(errorText))
		content.WriteString("\n\n")

		helpText := padToWidth("Press Enter or 'q' to exit", contentWidth)
		content.WriteString(m.styles.Help.Render(helpText))
	}

	return m.styles.Root.Render(content.String())
}

// Helper function to pad text to a specific width
func padToWidth(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth < width {
		return text + strings.Repeat(" ", width-textWidth)
	}
	return text
}


// Commands (unchanged from previous version)
func checkFilesExist(filenames []string) tea.Cmd {
	return func() tea.Msg {
		fileInfos := make([]FileInfo, len(filenames))

		for i, filename := range filenames {
			stat, err := os.Stat(filename)
			if err != nil {
				fileInfos[i] = FileInfo{
					Path:    filename,
					Exists:  false,
					Error:   err.Error(),
				}
				continue
			}

			isDir := stat.IsDir()
			fileCount := 0

			if isDir {
				fileCount = countFilesInDirectory(filename)
			}

			fileInfos[i] = FileInfo{
				Path:        filename,
				IsDirectory: isDir,
				FileCount:   fileCount,
				Exists:      true,
			}
		}

		return filesExistMsg{fileInfos: fileInfos}
	}
}

func findNextValidFile(fileInfos []FileInfo, startIndex int) int {
	for i := startIndex; i < len(fileInfos); i++ {
		if fileInfos[i].Exists {
			return i
		}
	}
	return -1
}

func moveNextFileToCache(fileInfos []FileInfo, index int, config Config) tea.Cmd {
	return func() tea.Msg {
		// Skip non-existent files
		if index >= len(fileInfos) || !fileInfos[index].Exists {
			return fileMoveMsg{item: DeletedItem{}, err: nil}
		}

		filename := fileInfos[index].Path
		return moveFileToCache(filename, config)()
	}
}

func moveFileToCache(filename string, config Config) tea.Cmd {
	return func() tea.Msg {
		// Ensure cache directory exists
		cacheDir := expandPath(config.Cache.Directory)
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return fileMoveMsg{err: err}
		}

		// Get file info
		stat, err := os.Stat(filename)
		if err != nil {
			return fileMoveMsg{err: err}
		}

		// Get absolute path
		absPath, err := filepath.Abs(filename)
		if err != nil {
			return fileMoveMsg{err: err}
		}

		// Generate unique ID and cache filename
		now := time.Now()
		id := fmt.Sprintf("%d", now.UnixNano())
		timestamp := now.Format("2006-01-02-15-04-05")
		baseFilename := filepath.Base(filename)
		cacheFilename := fmt.Sprintf("%s-%s-%s", id, timestamp, baseFilename)
		cachePath := filepath.Join(cacheDir, cacheFilename)

		isDir := stat.IsDir()
		fileCount := 0
		size := stat.Size()

		// Move file or directory
		if isDir {
			fileCount = countFilesInDirectory(filename)
			size = getDirectorySize(filename)
			if err := moveDirectory(filename, cachePath); err != nil {
				return fileMoveMsg{err: err}
			}
		} else {
			if err := moveFile(filename, cachePath); err != nil {
				return fileMoveMsg{err: err}
			}
		}

		// Create deleted item
		item := DeletedItem{
			ID:           id,
			OriginalPath: absPath,
			DeleteDate:   now,
			CachePath:    cachePath,
			IsDirectory:  isDir,
			FileCount:    fileCount,
			Size:         size,
		}

		// Update index
		if err := addToIndex(item, config); err != nil {
			return fileMoveMsg{err: err}
		}

		// Log the operation
		if config.Logging.Enabled {
			logOperation("DELETE", item, config)
		}

		return fileMoveMsg{item: item, err: nil}
	}
}

func cleanupOldFiles(config Config) tea.Cmd {
	return func() tea.Msg {
		cutoffDays := time.Duration(config.Cache.Days) * 24 * time.Hour
		cutoff := time.Now().Add(-cutoffDays)

		index, err := loadIndex(config)
		if err != nil {
			return errorMsg(fmt.Sprintf("Error loading index: %v", err))
		}

		var remainingItems []DeletedItem
		for _, item := range index.Items {
			if item.DeleteDate.Before(cutoff) {
				// Remove the actual file or directory
				if item.IsDirectory {
					os.RemoveAll(item.CachePath)
				} else {
					os.Remove(item.CachePath)
				}

				// Log cleanup
				if config.Logging.Enabled {
					logOperation("CLEANUP", item, config)
				}
			} else {
				remainingItems = append(remainingItems, item)
			}
		}

		// Update index
		index.Items = remainingItems
		if err := saveIndex(index, config); err != nil {
			return errorMsg(fmt.Sprintf("Error updating index: %v", err))
		}

		return cleanupMsg{}
	}
}

func clearAllCache(config Config) tea.Cmd {
	return func() tea.Msg {
		cacheDir := expandPath(config.Cache.Directory)

		// Remove all files in cache directory
		if err := os.RemoveAll(cacheDir); err != nil {
			return clearMsg{err: err}
		}

		// Recreate cache directory
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return clearMsg{err: err}
		}

		// Create empty index
		index := Index{Items: []DeletedItem{}}
		if err := saveIndex(index, config); err != nil {
			return clearMsg{err: err}
		}

		// Log clear operation
		if config.Logging.Enabled {
			logDir := expandPath(config.Logging.Directory)
			os.MkdirAll(logDir, 0755)
			logPath := filepath.Join(logDir, "saferm.log")
			logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer logFile.Close()
				logFile.WriteString(fmt.Sprintf("%s CLEAR_ALL Cache cleared\n", time.Now().Format("2006-01-02 15:04:05")))
			}
		}

		return clearMsg{err: nil}
	}
}

// Helper functions
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	if !filepath.IsAbs(path) {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path)
	}
	return path
}

func getIndexPath(config Config) string {
	cacheDir := expandPath(config.Cache.Directory)
	return filepath.Join(cacheDir, "index.json")
}

func loadIndex(config Config) (Index, error) {
	var index Index
	indexPath := getIndexPath(config)

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return Index{Items: []DeletedItem{}}, nil
		}
		return index, err
	}

	err = json.Unmarshal(data, &index)
	return index, err
}

func saveIndex(index Index, config Config) error {
	indexPath := getIndexPath(config)
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

func addToIndex(item DeletedItem, config Config) error {
	index, err := loadIndex(config)
	if err != nil {
		return err
	}

	index.Items = append(index.Items, item)
	return saveIndex(index, config)
}

func logOperation(operation string, item DeletedItem, config Config) {
	logDir := expandPath(config.Logging.Directory)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	logPath := filepath.Join(logDir, "saferm.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	logEntry := fmt.Sprintf("%s %s %s -> %s (Size: %d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"),
		operation,
		item.OriginalPath,
		item.CachePath,
		item.Size)

	logFile.WriteString(logEntry)
}

func moveFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Remove(src)
}

func moveDirectory(src, dst string) error {
	// Use os.Rename for atomic operation when possible (same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback to copy + remove for cross-filesystem moves
	if err := copyDirectory(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

func copyDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func countFilesInDirectory(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path != dir {
			count++
		}
		return nil
	})
	return count
}

func getDirectorySize(dir string) int64 {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func showUsage() {
	fmt.Println("SafeRM - Safe file/directory removal tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  saferm <file|directory> [file2] [dir2] ...    Remove multiple files or directories safely")
	fmt.Println("  saferm --clear                                Clear all cached files immediately")
	fmt.Println("  saferm --themes                               List available themes")
	fmt.Println("  saferm -h, --help                             Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  saferm file1.txt")
	fmt.Println("  saferm file1.txt file2.txt directory1/")
	fmt.Println("  saferm *.log temp_folder/")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.config/saferm/saferm.toml")
	fmt.Println("  Default cache location: ~/.cache/saferm")
	fmt.Println("  Default retention: 10 days")
	fmt.Println("  Available themes: default, dark, light, cyberpunk, minimal")
}


func showThemes() {
	themes := getDefaultThemes()
	fmt.Println("Available SafeRM Themes:")
	fmt.Println("=" + strings.Repeat("=", 40))

	for name, theme := range themes {
		fmt.Printf("\nTheme: %s\n", strings.ToUpper(name))
		fmt.Printf("  Primary:    %s\n", theme.UI.Colors.Primary)
		fmt.Printf("  Success:    %s\n", theme.UI.Colors.Success)
		fmt.Printf("  Warning:    %s\n", theme.UI.Colors.Warning)
		fmt.Printf("  Error:      %s\n", theme.UI.Colors.Error)
		fmt.Printf("  Progress:   %s\n", theme.UI.Progress.Style)
	}

	fmt.Println("\nTo use a theme, set 'theme = \"name\"' in your saferm.toml config file.")
	fmt.Println("You can also override individual colors in the [ui.colors] section.")
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

	// Handle themes flag
	if arg == "--themes" {
		showThemes()
		return
	}

	// Handle clear flag
	if arg == "--clear" {
		model, err := initialModel([]string{}, true)
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			os.Exit(1)
		}
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Normal file deletion - collect all arguments as filenames
	filenames := os.Args[1:]
	model, err := initialModel(filenames, false)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
