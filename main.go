package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
		} `toml:"colors"`
		Progress struct {
			Style      string `toml:"style"` // "gradient", "solid", "rainbow"
			ShowEmoji  bool   `toml:"show_emoji"`
			Animation  bool   `toml:"animation"`
		} `toml:"progress"`
		NoConfirm bool `toml:"no_confirm"`
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
	Root 		lipgloss.Style
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
	noConfirm      bool
	operation      string // "delete", "restore", "clear", "purge"
	restoreItems   []DeletedItem
}

// Messages
type filesExistMsg struct {
	fileInfos []FileInfo
}

type restoreItemsMsg struct {
	items []DeletedItem
}

type fileMoveMsg struct {
	item DeletedItem
	err  error
}

type restoreMsg struct {
	item DeletedItem
	err  error
}

type cleanupMsg struct{}

type clearMsg struct {
	err error
}

type purgeMsg struct {
	purgedCount int
	err         error
}

type errorMsg string

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
	defaultTheme.UI.NoConfirm = false
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
	darkTheme.UI.NoConfirm = false
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
	lightTheme.UI.NoConfirm = false
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
	cyberpunkTheme.UI.NoConfirm = false
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
	minimalTheme.UI.NoConfirm = false
	themes["minimal"] = minimalTheme

	return themes
}

func createThemeStyles(config Config) ThemeStyles {
	colors := config.UI.Colors
	return ThemeStyles{
		Root: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(2),
			// PaddingBottom(2).
			// PaddingLeft(4),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Text)).
			Bold(true).
			Padding(0, 2, 0, 2).
			MarginBottom(1),
		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true).
			Underline(true).
			MarginBottom(1),
		Question: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true),
		Filename: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Highlight)).
			Bold(true).
			Underline(true),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Success)),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Error)),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Warning)).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Secondary)).
			Padding(0, 1),
			// Removed border and made it responsive
			// Border(lipgloss.NormalBorder()).
			// BorderForeground(lipgloss.Color(colors.Border))

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Muted)).
			Italic(true).
			MarginTop(1),
		Progress: lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1),
		Border: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(colors.Border)).
			Padding(1),
		List: lipgloss.NewStyle().
			MarginLeft(2).
			MarginTop(1).
			MarginBottom(1),
		StatusGood: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)),
		StatusBad: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)),
	}
}

func loadConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(homeDir, ".config", "vanish", "vanish.toml")

	// Default configuration
	config := Config{}
	config.Cache.Directory = filepath.Join(homeDir, ".cache", "vanish")
	config.Cache.Days = 10
	config.Logging.Enabled = true
	config.Logging.Directory = filepath.Join(homeDir, ".cache", "vanish", "logs")

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
				customProgress := config.UI.Progress
				customNoConfirm := config.UI.NoConfirm
				// Apply theme
				config.UI = themeConfig.UI
				// Restore any custom colors that were explicitly set
				if customColors.Primary != "" {
					config.UI.Colors.Primary = customColors.Primary
				}
				if customColors.Secondary != "" {
					config.UI.Colors.Secondary = customColors.Secondary
				}
				if customColors.Success != "" {
					config.UI.Colors.Success = customColors.Success
				}
				if customColors.Warning != "" {
					config.UI.Colors.Warning = customColors.Warning
				}
				if customColors.Error != "" {
					config.UI.Colors.Error = customColors.Error
				}
				if customColors.Text != "" {
					config.UI.Colors.Text = customColors.Text
				}
				if customColors.Muted != "" {
					config.UI.Colors.Muted = customColors.Muted
				}
				if customColors.Border != "" {
					config.UI.Colors.Border = customColors.Border
				}
				if customColors.Highlight != "" {
					config.UI.Colors.Highlight = customColors.Highlight
				}
				// Restore custom progress settings
				if customProgress.Style != "" {
					config.UI.Progress.Style = customProgress.Style
				}
				// Restore no confirm setting
				config.UI.NoConfirm = customNoConfirm
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
// getTerminalSize returns the current terminal width and height
func getTerminalSize() (int, int) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			return width, height
		}
	}
	// Fallback to reasonable defaults
	return 80, 24
}
func createDefaultConfig(configPath string, config Config) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configContent := `[cache]
# Directory where deleted files are stored (relative to home directory)
directory = ".cache/vanish"
# Number of days to keep deleted files
days = 10

[logging]
# Enable logging
enabled = true
# Directory for log files (relative to cache directory)
directory = ".cache/vanish/logs"

[ui]
# Theme: "default", "dark", "light", "cyberpunk", "minimal"
theme = "default"
# Skip confirmation prompts by default
no_confirm = false

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
	prog := progress.New()
	prog.Width = 50

	switch config.UI.Progress.Style {
	case "solid":
		prog = progress.New(progress.WithSolidFill(config.UI.Colors.Primary))
	case "rainbow":
		prog = progress.New(progress.WithGradient("#FF0000", "#9400D3")) //  "#FF7F00", "#FFFF00", "#00FF00", "#0000FF", "#4B0082",
	default: // gradient
		prog = progress.New(progress.WithGradient(config.UI.Colors.Primary, config.UI.Colors.Secondary))
	}

	return prog
}

func initialModel(filenames []string, operation string, noConfirm bool) (Model, error) {
	config, err := loadConfig()
	if err != nil {
		return Model{}, err
	}

	prog := setupProgress(config)
	styles := createThemeStyles(config)

	// Check if no_confirm is set in config and not overridden by flag
	if config.UI.NoConfirm && !noConfirm {
		noConfirm = true
	}

	return Model{
		filenames:      filenames,
		fileInfos:      make([]FileInfo, len(filenames)),
		state:          "checking",
		progress:       prog,
		config:         config,
		styles:         styles,
		operation:      operation,
		processedItems: make([]DeletedItem, 0),
		totalFiles:     len(filenames),
		noConfirm:      noConfirm,
	}, nil
}

func (m Model) Init() tea.Cmd {
	switch m.operation {
	case "clear":
		m.state = "clearing"
		return tea.Batch(
			m.progress.SetPercent(0.1),
			clearAllCache(m.config),
		)
	case "purge":
		m.state = "purging"
		return tea.Batch(
			m.progress.SetPercent(0.1),
			purgeOldFiles(m.config, m.filenames[0]), // days passed as first filename
		)
	case "restore":
		m.state = "checking"
		return tea.Batch(
			checkRestoreItems(m.filenames, m.config),
			m.progress.SetPercent(0.1),
		)
	default: // delete
		return tea.Batch(
			checkFilesExist(m.filenames),
			m.progress.SetPercent(0.1),
		)
	}
}
func countValidFiles(fileInfos []FileInfo) int {
	count := 0
	for _, info := range fileInfos {
		if info.Exists {
			count++
		}
	}
	return count
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
				if m.operation == "restore" {
					m.state = "restoring"
				} else {
					m.state = "moving"
				}
				m.currentIndex = 0
				return m, tea.Batch(
					m.progress.SetPercent(0.3),
					processNextItem(m),
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

		if m.noConfirm {
			m.confirmed = true
			m.state = "moving"
			m.currentIndex = 0
			return m, tea.Batch(
				m.progress.SetPercent(0.3),
				processNextItem(m),
			)
		} else {
			m.state = "confirming"
			return m, m.progress.SetPercent(0.2)
		}

	case restoreItemsMsg:
		m.restoreItems = msg.items
		if len(m.restoreItems) == 0 {
			m.state = "error"
			m.errorMsg = "No matching items found in cache for restoration"
			return m, nil
		}

		if m.noConfirm {
			m.confirmed = true
			m.state = "restoring"
			m.currentIndex = 0
			return m, tea.Batch(
				m.progress.SetPercent(0.3),
				processNextItem(m),
			)
		} else {
			m.state = "confirming"
			return m, m.progress.SetPercent(0.2)
		}

	case fileMoveMsg:
	if msg.err != nil {
		m.state = "error"
		m.errorMsg = fmt.Sprintf("Error processing item: %v", msg.err)
		return m, nil
	}

	if msg.item.ID != "" {
		m.processedItems = append(m.processedItems, msg.item)
		m.processedFiles++
	}

	// Find the next valid file index, starting from current + 1
	nextIndex := findNextValidFile(m.fileInfos, m.currentIndex+1)

	// Update progress based on processed files vs total valid files
	validFileCount := countValidFiles(m.fileInfos)
	progressPercent := 0.3 + (float64(m.processedFiles)/float64(validFileCount))*0.4

	// Check if we have more valid items to process
	if nextIndex != -1 {
		m.currentIndex = nextIndex
		return m, tea.Batch(
			m.progress.SetPercent(progressPercent),
			processNextItem(m),
		)
	}

	// All items processed, move to cleanup
	m.state = "cleanup"
	return m, tea.Batch(
		m.progress.SetPercent(0.7),
		cleanupOldFiles(m.config),
	)


	case restoreMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error restoring item: %v", msg.err)
			return m, nil
		}

		if msg.item.ID != "" {
			m.processedItems = append(m.processedItems, msg.item)
			m.processedFiles++
		}

		m.currentIndex++

		// Update progress
		progressPercent := 0.3 + (float64(m.currentIndex)/float64(len(m.restoreItems)))*0.4

		// Check if we have more items to restore
		if m.currentIndex < len(m.restoreItems) {
			return m, tea.Batch(
				m.progress.SetPercent(progressPercent),
				processNextItem(m),
			)
		}

		// All items restored
		m.state = "done"
		return m, m.progress.SetPercent(1.0)

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

	case purgeMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error purging cache: %v", msg.err)
			return m, nil
		}
		m.processedFiles = msg.purgedCount
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

	// Get terminal dimensions for responsive design
	termWidth, termHeight := lipgloss.Size(m.styles.Root.String())
	if termWidth == 0 {
		termWidth = 80 // fallback width
	}
	if termHeight == 0 {
		termHeight = 24 // fallback height
	}

	// Calculate content width (leaving some margin)
	contentWidth := termWidth - 8 // 4 chars margin on each side

	switch m.state {
	case "checking":
		m.renderCheckingState(&content)
	case "confirming":
		m.renderConfirmingState(&content, contentWidth)
	case "moving":
		m.renderMovingState(&content, contentWidth)
	case "restoring":
		m.renderRestoringState(&content, contentWidth)
	case "cleanup":
		m.renderCleanupState(&content)
	case "clearing":
		m.renderClearingState(&content)
	case "purging":
		m.renderPurgingState(&content)
	case "done":
		m.renderDoneState(&content, contentWidth)
	case "error":
		m.renderErrorState(&content)
	}

	return m.styles.Root.Render(content.String())
}

func (m Model) renderCheckingState(content *strings.Builder) {
	if m.config.UI.Progress.ShowEmoji {
		content.WriteString("🔍 ")
	}

	message := "Checking files and directories...\n"
	if m.operation == "restore" {
		message = "Checking items for restoration...\n"
	}

	content.WriteString(message)
	content.WriteString(m.styles.Progress.Render(m.progress.View()))
}

func (m Model) renderConfirmingState(content *strings.Builder, contentWidth int) {
	if m.operation == "restore" {
		m.renderRestoreConfirmation(content)
	} else {
		m.renderDeleteConfirmation(content, contentWidth)
	}

	content.WriteString("\n")
	content.WriteString(m.styles.Help.Render("Press 'y' to confirm, 'n' to cancel, or 'q' to quit"))
}

func (m Model) renderRestoreConfirmation(content *strings.Builder) {
	content.WriteString(m.styles.Question.Render("Are you sure you want to restore the following items?"))
	content.WriteString("\n")

	listContent := m.buildRestoreItemsList()
	content.WriteString(m.styles.List.Render(listContent))

	infoStyle := m.styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MarginTop(1)
	content.WriteString(infoStyle.Render(fmt.Sprintf("Total items to restore: %d", len(m.restoreItems))))
}

func (m Model) buildRestoreItemsList() string {
	var listContent strings.Builder

	for _, item := range m.restoreItems {
		icon := m.getFileIcon(item.IsDirectory)
		listContent.WriteString(icon)
		listContent.WriteString(m.styles.Filename.Render(item.OriginalPath))
		listContent.WriteString(m.styles.Info.Render(fmt.Sprintf(" (deleted: %s)", item.DeleteDate.Format("2006-01-02 15:04"))))
		listContent.WriteString("\n")
	}

	return listContent.String()
}

func (m Model) renderDeleteConfirmation(content *strings.Builder, contentWidth int) {
	content.WriteString(m.styles.Question.Render("Are you sure you want to delete the following items?"))
	content.WriteString("\n")

	validCount, invalidCount, totalFileCount := m.analyzeFileInfos()
	listContent := m.buildFileInfosList(validCount, invalidCount, &totalFileCount)

	content.WriteString(m.styles.List.Render(listContent))

	if invalidCount > 0 {
		m.renderInvalidFilesWarning(content, invalidCount)
	}

	if validCount > 0 {
		m.renderDeleteSummary(content, validCount, totalFileCount, contentWidth)
	}
}

func (m Model) analyzeFileInfos() (validCount, invalidCount, totalFileCount int) {
	for _, info := range m.fileInfos {
		if info.Exists {
			validCount++
			if info.IsDirectory {
				totalFileCount += info.FileCount
			} else {
				totalFileCount++
			}
		} else {
			invalidCount++
		}
	}
	return
}

func (m Model) buildFileInfosList(validCount, invalidCount int, totalFileCount *int) string {
	var listContent strings.Builder

	for _, info := range m.fileInfos {
		if info.Exists {
			m.appendValidFileInfo(&listContent, info, totalFileCount)
		} else {
			m.appendInvalidFileInfo(&listContent, info)
		}
	}

	return listContent.String()
}

func (m Model) appendValidFileInfo(listContent *strings.Builder, info FileInfo, totalFileCount *int) {
	icon := m.getFileIcon(info.IsDirectory)
	listContent.WriteString(icon)
	listContent.WriteString(m.styles.Filename.Render(info.Path))

	if info.IsDirectory {
		inlineInfoStyle := m.styles.Info.Border(lipgloss.Border{}).Padding(0)
		if info.FileCount > 0 {
			listContent.WriteString(inlineInfoStyle.Render(fmt.Sprintf(" (%d items)", info.FileCount)))
		} else {
			listContent.WriteString(inlineInfoStyle.Render(" (empty)"))
		}
	}

	listContent.WriteString("\n")
}

func (m Model) appendInvalidFileInfo(listContent *strings.Builder, info FileInfo) {
	icon := "ERR "
	if m.config.UI.Progress.ShowEmoji {
		icon = "❌ "
	}

	listContent.WriteString(icon)
	listContent.WriteString(m.styles.StatusBad.Render(info.Path))
	listContent.WriteString(m.styles.Warning.Render(" (does not exist)"))
}

func (m Model) renderInvalidFilesWarning(content *strings.Builder, invalidCount int) {
	content.WriteString("\n")
	warningText := fmt.Sprintf("⚠ Warning: %d file(s) will be skipped", invalidCount)
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.config.UI.Colors.Warning)).
		Bold(true)
	content.WriteString(warningStyle.Render(warningText))
}

func (m Model) renderDeleteSummary(content *strings.Builder, validCount, totalFileCount, contentWidth int) {
	content.WriteString("\n")

	infoText := fmt.Sprintf("Total items to delete: %d", validCount)
	if totalFileCount > validCount {
		infoText += fmt.Sprintf(" | Files affected: %d", totalFileCount)
	}
	infoText += fmt.Sprintf(" | Recoverable for %d days", m.config.Cache.Days)

	infoStyle := m.styles.Info.MaxWidth(contentWidth).Align(lipgloss.Left)
	content.WriteString(infoStyle.Render(infoText))
}

func (m Model) renderMovingState(content *strings.Builder, contentWidth int) {
	statusText := m.buildProgressStatusText("Moving", "📦")
	m.renderProgressState(content, statusText, contentWidth)
}

func (m Model) renderRestoringState(content *strings.Builder, contentWidth int) {
	statusText := m.buildRestoreStatusText()
	m.renderProgressState(content, statusText, contentWidth)
}

func (m Model) buildProgressStatusText(action, emoji string) string {
	if m.currentIndex < len(m.fileInfos) {
		currentFile := m.fileInfos[m.currentIndex]
		fileType := m.getFileTypeString(currentFile.IsDirectory)

		emojiPrefix := ""
		if m.config.UI.Progress.ShowEmoji {
			emojiPrefix = emoji + " "
		}

		return fmt.Sprintf("%s%s %s '%s' to safe cache... (%d/%d)",
			emojiPrefix, action, fileType, currentFile.Path, m.processedFiles+1, m.totalFiles)
	}

	fallback := fmt.Sprintf("%s files to safe cache...", action)
	if m.config.UI.Progress.ShowEmoji {
		fallback = emoji + " " + fallback
	}
	return fallback
}

func (m Model) buildRestoreStatusText() string {
	if m.currentIndex < len(m.restoreItems) {
		currentItem := m.restoreItems[m.currentIndex]
		fileType := m.getFileTypeString(currentItem.IsDirectory)

		emojiPrefix := ""
		if m.config.UI.Progress.ShowEmoji {
			emojiPrefix = "♻️ "
		}

		return fmt.Sprintf("%sRestoring %s '%s'... (%d/%d)",
			emojiPrefix, fileType, currentItem.OriginalPath, m.processedFiles+1, len(m.restoreItems))
	}

	fallback := "Restoring files from cache..."
	if m.config.UI.Progress.ShowEmoji {
		fallback = "♻️ " + fallback
	}
	return fallback
}

func (m Model) renderProgressState(content *strings.Builder, statusText string, contentWidth int) {
	statusStyle := m.styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MaxWidth(contentWidth)
	content.WriteString(statusStyle.Render(statusText))
	content.WriteString("\n")
	content.WriteString(m.styles.Progress.Render(m.progress.View()))
}

func (m Model) renderCleanupState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🧹", "Cleaning up old cached files...")
}

func (m Model) renderClearingState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🗑️", "Clearing all cached files...")
}

func (m Model) renderPurgingState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🔥", "Purging old cached files...")
}

func (m Model) renderSimpleProgressState(content *strings.Builder, emoji, message string) {
	if m.config.UI.Progress.ShowEmoji {
		content.WriteString(emoji + " ")
	}
	content.WriteString(message + "\n")
	content.WriteString(m.styles.Progress.Render(m.progress.View()))
}

func (m Model) renderDoneState(content *strings.Builder, contentWidth int) {
	successMsg := m.buildSuccessMessage()
	content.WriteString(m.styles.Success.Render(successMsg))
	content.WriteString("\n")

	if m.shouldShowItemDetails() {
		m.renderItemDetails(content, contentWidth)
	}

	content.WriteString(m.styles.Progress.Render(m.progress.View()))
	content.WriteString("\n")
	content.WriteString(m.styles.Help.Render("Press Enter or 'q' to exit"))
}

func (m Model) buildSuccessMessage() string {
	var successMsg string
	emoji := ""

	if m.config.UI.Progress.ShowEmoji {
		emoji = "✅ "
	}

	switch m.operation {
	case "clear":
		if m.config.UI.Progress.ShowEmoji {
			successMsg = "✅ All cached files cleared!"
		} else {
			successMsg = "SUCCESS: All cached files cleared!"
		}
	case "purge":
		if m.config.UI.Progress.ShowEmoji {
			successMsg = fmt.Sprintf("✅ Purged %d old cached files!", m.processedFiles)
		} else {
			successMsg = fmt.Sprintf("SUCCESS: Purged %d old cached files!", m.processedFiles)
		}
	case "restore":
		successMsg = fmt.Sprintf("%sSuccessfully restored %d item(s)!", emoji, len(m.processedItems))
	default:
		successMsg = fmt.Sprintf("%sSuccessfully processed %d item(s)!", emoji, len(m.processedItems))
	}

	return successMsg
}

func (m Model) shouldShowItemDetails() bool {
	return (m.operation == "delete" || m.operation == "restore") && len(m.processedItems) > 0
}

func (m Model) renderItemDetails(content *strings.Builder, contentWidth int) {
	detailsBuilder := m.buildItemDetailsText()
	content.WriteString(m.styles.List.Render(detailsBuilder))

	if m.operation == "delete" && len(m.processedItems) > 0 {
		m.renderDeletionInfo(content, contentWidth)
	}
}

func (m Model) buildItemDetailsText() string {
	var detailsBuilder strings.Builder

	maxItems := 5
	for i, item := range m.processedItems {
		if i >= maxItems {
			break
		}

		if m.operation == "restore" {
			detailsBuilder.WriteString(fmt.Sprintf("• %s ← %s\n",
				m.styles.Filename.Render(item.OriginalPath), "cache"))
		} else {
			detailsBuilder.WriteString(fmt.Sprintf("• %s → %s\n",
				m.styles.Filename.Render(item.OriginalPath), filepath.Base(item.CachePath)))
		}
	}

	if len(m.processedItems) > maxItems {
		infoStyle := m.styles.Info.Border(lipgloss.Border{}).Padding(0)
		detailsBuilder.WriteString(infoStyle.Render(fmt.Sprintf("... and %d more item(s)", len(m.processedItems)-maxItems)))
		detailsBuilder.WriteString("\n")
	}

	return detailsBuilder.String()
}

func (m Model) renderDeletionInfo(content *strings.Builder, contentWidth int) {
	deleteAfter := m.processedItems[0].DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
	infoStyle := m.styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MaxWidth(contentWidth)

	content.WriteString("\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Will be permanently deleted after: %s", deleteAfter.Format("2006-01-02 15:04:05"))))
	content.WriteString("\n")
}

func (m Model) renderErrorState(content *strings.Builder) {
	errorMsg := "ERROR"
	if m.config.UI.Progress.ShowEmoji {
		errorMsg = "❌ Error"
	}

	content.WriteString(m.styles.Error.Render(errorMsg))
	content.WriteString("\n")
	content.WriteString(m.errorMsg)
	content.WriteString("\n")
	content.WriteString(m.styles.Help.Render("Press Enter or 'q' to exit"))
}

// Helper functions
func (m Model) getFileIcon(isDirectory bool) string {
	if m.config.UI.Progress.ShowEmoji {
		if isDirectory {
			return "📁 "
		}
		return "📄 "
	}

	if isDirectory {
		return "DIR "
	}
	return "FILE "
}

func (m Model) getFileTypeString(isDirectory bool) string {
	if isDirectory {
		return "directory"
	}
	return "file"
}


// Commands
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

func checkRestoreItems(patterns []string, config Config) tea.Cmd {
	return func() tea.Msg {
		index, err := loadIndex(config)
		if err != nil {
			return errorMsg(fmt.Sprintf("Error loading index: %v", err))
		}

		var matchingItems []DeletedItem
		for _, pattern := range patterns {
			for _, item := range index.Items {
				// Simple pattern matching - check if pattern is contained in original path
				if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(pattern)) {
					matchingItems = append(matchingItems, item)
				}
			}
		}

		return restoreItemsMsg{items: matchingItems}
	}
}

func processNextItem(m Model) tea.Cmd {
	if m.operation == "restore" {
		if m.currentIndex >= len(m.restoreItems) {
			return nil
		}
		return restoreFromCache(m.restoreItems[m.currentIndex], m.config)
	} else {
		// Make sure we have a valid index
		if m.currentIndex < 0 || m.currentIndex >= len(m.fileInfos) {
			return nil
		}

		// Make sure the file at current index exists
		if !m.fileInfos[m.currentIndex].Exists {
			return nil
		}

		return moveFileToCache(m.fileInfos[m.currentIndex].Path, m.config)
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

func restoreFromCache(item DeletedItem, config Config) tea.Cmd {
	return func() tea.Msg {
		// Check if cache file exists
		if _, err := os.Stat(item.CachePath); os.IsNotExist(err) {
			return restoreMsg{err: fmt.Errorf("cached file not found: %s", item.CachePath)}
		}

		// Create directory for original path if needed
		originalDir := filepath.Dir(item.OriginalPath)
		if err := os.MkdirAll(originalDir, 0755); err != nil {
			return restoreMsg{err: fmt.Errorf("failed to create directory %s: %v", originalDir, err)}
		}

		// Check if original path already exists
		if _, err := os.Stat(item.OriginalPath); !os.IsNotExist(err) {
			return restoreMsg{err: fmt.Errorf("destination already exists: %s", item.OriginalPath)}
		}

		// Move file back
		if item.IsDirectory {
			if err := moveDirectory(item.CachePath, item.OriginalPath); err != nil {
				return restoreMsg{err: err}
			}
		} else {
			if err := moveFile(item.CachePath, item.OriginalPath); err != nil {
				return restoreMsg{err: err}
			}
		}

		// Remove from index
		if err := removeFromIndex(item.ID, config); err != nil {
			// Log error but don't fail the restore
			if config.Logging.Enabled {
				logDir := expandPath(config.Logging.Directory)
				os.MkdirAll(logDir, 0755)
				logPath := filepath.Join(logDir, "vanish.log")
				logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					defer logFile.Close()
					logFile.WriteString(fmt.Sprintf("%s ERROR Failed to remove from index: %s\n",
						time.Now().Format("2006-01-02 15:04:05"), item.ID))
				}
			}
		}

		// Log the restore operation
		if config.Logging.Enabled {
			logOperation("RESTORE", item, config)
		}

		return restoreMsg{item: item, err: nil}
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
			logPath := filepath.Join(logDir, "vanish.log")
			logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				defer logFile.Close()
				logFile.WriteString(fmt.Sprintf("%s CLEAR_ALL Cache cleared\n", time.Now().Format("2006-01-02 15:04:05")))
			}
		}

		return clearMsg{err: nil}
	}
}

func purgeOldFiles(config Config, daysStr string) tea.Cmd {
	return func() tea.Msg {
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return purgeMsg{err: fmt.Errorf("invalid days value: %s", daysStr)}
		}

		cutoffDays := time.Duration(days) * 24 * time.Hour
		cutoff := time.Now().Add(-cutoffDays)

		index, err := loadIndex(config)
		if err != nil {
			return purgeMsg{err: fmt.Errorf("Error loading index: %v", err)}
		}

		var remainingItems []DeletedItem
		purgedCount := 0

		for _, item := range index.Items {
			if item.DeleteDate.Before(cutoff) {
				// Remove the actual file or directory
				if item.IsDirectory {
					os.RemoveAll(item.CachePath)
				} else {
					os.Remove(item.CachePath)
				}
				purgedCount++

				// Log purge
				if config.Logging.Enabled {
					logOperation("PURGE", item, config)
				}
			} else {
				remainingItems = append(remainingItems, item)
			}
		}

		// Update index
		index.Items = remainingItems
		if err := saveIndex(index, config); err != nil {
			return purgeMsg{err: fmt.Errorf("Error updating index: %v", err)}
		}

		return purgeMsg{purgedCount: purgedCount, err: nil}
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

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "vanish", "vanish.toml")
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

func removeFromIndex(itemID string, config Config) error {
	index, err := loadIndex(config)
	if err != nil {
		return err
	}

	var remainingItems []DeletedItem
	for _, item := range index.Items {
		if item.ID != itemID {
			remainingItems = append(remainingItems, item)
		}
	}

	index.Items = remainingItems
	return saveIndex(index, config)
}

func logOperation(operation string, item DeletedItem, config Config) {
	logDir := expandPath(config.Logging.Directory)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	logPath := filepath.Join(logDir, "vanish.log")
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

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func showUsage(config Config) {
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

func showThemes() {
	themes := getDefaultThemes()
	fmt.Println("Available Vanish Themes:")
	fmt.Println("=" + strings.Repeat("=", 40))

	for name, theme := range themes {
		fmt.Printf("\nTheme: %s\n", strings.ToUpper(name))
		fmt.Printf("  Primary:    %s\n", theme.UI.Colors.Primary)
		fmt.Printf("  Success:    %s\n", theme.UI.Colors.Success)
		fmt.Printf("  Warning:    %s\n", theme.UI.Colors.Warning)
		fmt.Printf("  Error:      %s\n", theme.UI.Colors.Error)
		fmt.Printf("  Progress:   %s\n", theme.UI.Progress.Style)
		fmt.Printf("  Emoji:      %t\n", theme.UI.Progress.ShowEmoji)
	}

	fmt.Println("\nTo use a theme, set 'theme = \"name\"' in your vanish.toml config file.")
	fmt.Println("You can also override individual colors in the [ui.colors] section.")
}

func showList(config Config) error {
	index, err := loadIndex(config)
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
			formatBytes(item.Size),
			status,
			daysLeft,
			item.OriginalPath,
		)
	}

	return nil
}

func showInfo(pattern string, config Config) error {
	index, err := loadIndex(config)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	var matchingItems []DeletedItem
	for _, item := range index.Items {
		if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(pattern)) {
			matchingItems = append(matchingItems, item)
		}
	}

	if len(matchingItems) == 0 {
		fmt.Printf("No cached items found matching pattern: %s\n", pattern)
		return nil
	}

	fmt.Printf("Found %d matching item(s):\n", len(matchingItems))
	fmt.Println(strings.Repeat("=", 60))

	for _, item := range matchingItems {
		fmt.Printf("\nID: %s\n", item.ID)
		fmt.Printf("Original Path: %s\n", item.OriginalPath)
		fmt.Printf("Cache Path: %s\n", item.CachePath)
		fmt.Printf("Deleted: %s\n", item.DeleteDate.Format("2006-01-02 15:04:05"))
		fmt.Printf("Type: %s\n", func() string {
			if item.IsDirectory {
				return "Directory"
			}
			return "File"
		}())
		fmt.Printf("Size: %s\n", formatBytes(item.Size))
		if item.FileCount > 0 {
			fmt.Printf("Files Inside: %d\n", item.FileCount)
		}

		expiryDate := item.DeleteDate.Add(time.Duration(config.Cache.Days) * 24 * time.Hour)
		daysLeft := int(time.Until(expiryDate).Hours() / 24)

		if daysLeft > 0 {
			fmt.Printf("Expires: %s (%d days left)\n",
				expiryDate.Format("2006-01-02 15:04:05"), daysLeft)
		} else {
			fmt.Printf("Status: EXPIRED (can be purged)\n")
		}
	}

	return nil
}

func showStats(config Config) error {
	index, err := loadIndex(config)
	if err != nil {
		return fmt.Errorf("error loading index: %v", err)
	}

	if len(index.Items) == 0 {
		fmt.Println("Cache is empty.")
		return nil
	}

	var totalSize int64
	var fileCount, dirCount int
	var expiredCount int

	cutoff := time.Now().Add(-time.Duration(config.Cache.Days) * 24 * time.Hour)

	for _, item := range index.Items {
		totalSize += item.Size
		if item.IsDirectory {
			dirCount++
		} else {
			fileCount++
		}

		if item.DeleteDate.Before(cutoff) {
			expiredCount++
		}
	}

	fmt.Printf("Vanish Cache Statistics\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Cache Directory: %s\n", expandPath(config.Cache.Directory))
	fmt.Printf("Total Items: %d\n", len(index.Items))
	fmt.Printf("  Files: %d\n", fileCount)
	fmt.Printf("  Directories: %d\n", dirCount)
	fmt.Printf("Total Size: %s\n", formatBytes(totalSize))
	fmt.Printf("Retention Period: %d days\n", config.Cache.Days)
	fmt.Printf("Expired Items: %d\n", expiredCount)

	if expiredCount > 0 {
		fmt.Printf("\nRun 'vx --purge %d' to clean up expired items.\n", config.Cache.Days)
	}

	return nil
}

// Main function and argument parsing
func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	args := os.Args[1:]

	if len(args) == 0 {
		showUsage(config)
		return
	}

	// Parse command line arguments
	var operation string
	var filenames []string
	var noConfirm bool

	for i, arg := range args {
		switch arg {
		case "-h", "--help":
			showUsage(config)
			return
		case "--themes":
			showThemes()
			return
		case "--path":
			fmt.Println(expandPath(config.Cache.Directory))
			return
		case "--config-path":
			fmt.Println(getConfigPath())
			return
		case "--list":
			if err := showList(config); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		case "--stats":
			if err := showStats(config); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		case "--clear":
			operation = "clear"
			filenames = []string{""}
		case "--noconfirm":
			noConfirm = true
		case "--restore":
			operation = "restore"
			if i+1 < len(args) {
				filenames = args[i+1:]
			} else {
				log.Fatal("Error: --restore requires at least one pattern")
			}
			break
		case "--info":
			if i+1 < len(args) {
				if err := showInfo(args[i+1], config); err != nil {
					log.Fatalf("Error: %v", err)
				}
			} else {
				log.Fatal("Error: --info requires a pattern")
			}
			return
		case "--purge":
			if i+1 < len(args) {
				operation = "purge"
				filenames = []string{args[i+1]}
			} else {
				log.Fatal("Error: --purge requires number of days")
			}
			break
		default:
			// If no operation is set yet, this is a regular delete operation
			if operation == "" {
				operation = "delete"
				filenames = args[i:]
				break
			}
		}

		// Break early for operations that consume remaining args
		if operation == "restore" || operation == "delete" {
			break
		}
	}

	// Default to delete operation if no specific operation was specified
	if operation == "" && len(filenames) == 0 && len(args) > 0 {
		operation = "delete"
		// Filter out --noconfirm from filenames
		for _, arg := range args {
			if arg != "--noconfirm" {
				filenames = append(filenames, arg)
			}
		}
	}

	// Validate arguments
	if operation == "" || len(filenames) == 0 {
		if operation != "clear" {
			showUsage(config)
			os.Exit(1)
		}
	}

	// Initialize the TUI model
	m, err := initialModel(filenames, operation, noConfirm)
	if err != nil {
		log.Fatalf("Error initializing: %v", err)
	}

	// Create and run the Bubble Tea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
