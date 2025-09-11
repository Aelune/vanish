package tui

import (
	"fmt"
	"os"
	// "path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"vanish/internal/config"
	"vanish/internal/filesystem"
	"vanish/internal/logging"
	"vanish/internal/models"
	"vanish/internal/ui"
)

// Model represents the application state
type Model struct {
	filenames      []string
	fileInfos      []models.FileInfo
	currentIndex   int
	state          string
	progress       progress.Model
	progressVal    float64
	confirmed      bool
	errorMsg       string
	config         config.Config
	styles         ui.ThemeStyles
	processedItems []models.DeletedItem
	clearAll       bool
	noConfirm      bool
	totalFiles     int
	processedFiles int
	stats          models.OperationStats
	safeMode       bool // NEW: disable all styling if terminal has issues
}

// Messages (unchanged)
type filesExistMsg struct {
	fileInfos []models.FileInfo
}

type fileMoveMsg struct {
	item models.DeletedItem
	err  error
}

type cleanupMsg struct{}

type clearMsg struct {
	err error
}

type errorMsg string

func Start(m Model) error {
	// Check if we should run in safe mode (no TUI)
	if m.safeMode || !term.IsTerminal(int(os.Stdout.Fd())) {
		return runSafeMode(m)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}

// Run in safe mode without TUI
func runSafeMode(m Model) error {
	if m.clearAll {
		fmt.Println("Clearing all cached files...")
		if err := filesystem.ClearAllCache(m.config); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Println("Cache cleared successfully.")
		return nil
	}

	// Check files
	fmt.Println("Checking files...")
	validFiles := 0
	for i, filename := range m.filenames {
		m.fileInfos[i] = filesystem.CheckFileInfo(filename, m.config)
		if m.fileInfos[i].Exists {
			validFiles++
			fmt.Printf("  Found: %s\n", filename)
		} else {
			fmt.Printf("  Not found: %s\n", filename)
		}
	}

	if validFiles == 0 {
		return fmt.Errorf("no valid files found")
	}

	// Confirm deletion if needed
	if !m.noConfirm {
		fmt.Printf("\nDelete %d file(s)? [y/N]: ", validFiles)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Process files
	fmt.Println("Moving files to cache...")
	for _, info := range m.fileInfos {
		if !info.Exists {
			continue
		}

		fmt.Printf("  Processing: %s\n", info.Path)
		item, err := filesystem.MoveFileToCache(info.Path, m.config)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		if err := filesystem.AddToIndex(item, m.config); err != nil {
			fmt.Printf("  Warning: Failed to add to index: %v\n", err)
		}

		logging.LogOperation("DELETE", item, m.config)
		m.processedItems = append(m.processedItems, item)
	}

	// Cleanup
	fmt.Println("Cleaning up old files...")
	if err := filesystem.CleanupOldFiles(m.config); err != nil {
		fmt.Printf("Warning: Cleanup failed: %v\n", err)
	}

	fmt.Printf("Successfully processed %d item(s).\n", len(m.processedItems))
	return nil
}

// NewModel creates a new TUI model
func NewModel(filenames []string, clearAll bool, noConfirm bool, cfg config.Config) Model {
	// Detect if terminal has issues with complex styling
	safeMode := detectTerminalIssues()

	var prog progress.Model
	var styles ui.ThemeStyles

	if safeMode {
		// Create minimal progress and styles
		prog = progress.New()
		prog.Width = 0
		styles = createSafeStyles()
	} else {
		prog = ui.SetupProgress(cfg)
		styles = ui.CreateThemeStyles(cfg)
	}

	return Model{
		filenames:      filenames,
		fileInfos:      make([]models.FileInfo, len(filenames)),
		state:          "checking",
		progress:       prog,
		config:         cfg,
		styles:         styles,
		clearAll:       clearAll,
		noConfirm:      noConfirm,
		processedItems: make([]models.DeletedItem, 0),
		totalFiles:     len(filenames),
		stats:          models.OperationStats{},
		safeMode:       safeMode,
	}
}

// Detect terminal issues
func detectTerminalIssues() bool {
	// Check if stdout is not a TTY
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}

	// Check for problematic terminals
	termVar := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	// Common problematic terminals
	problemTerms := []string{
		"dumb",
		"unknown",
		"",
	}

	for _, pt := range problemTerms {
		if termVar == pt {
			return true
		}
	}

	// If no color support at all, use safe mode
	if !strings.Contains(termVar, "color") && colorterm == "" {
		return true
	}

	// Additional check: if TERM suggests limited capabilities
	if strings.Contains(termVar, "basic") || strings.Contains(termVar, "minimal") {
		return true
	}

	return false
}

// Create safe styles with no ANSI codes
func createSafeStyles() ui.ThemeStyles {
	base := ui.ThemeStyles{}
	// All styles are essentially no-ops in safe mode
	return base
}

func (m Model) Init() tea.Cmd {
	if m.clearAll {
		m.state = "clearing"
		return tea.Batch(
			clearAllCache(m.config),
		)
	}

	return checkFilesExist(m.filenames, m.config)
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
				return m, moveNextFileToCache(m.fileInfos, m.currentIndex, m.config)
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
				if info.IsDirectory {
					m.stats.TotalDirs++
					m.stats.TotalFiles += info.FileCount
				} else {
					m.stats.TotalFiles++
				}
				m.stats.TotalSize += info.Size
			}
		}

		if validFiles == 0 {
			m.state = "error"
			m.errorMsg = "No valid files or directories found"
			return m, nil
		}

		hasProtected := false
		for _, info := range m.fileInfos {
			if info.IsProtected || info.NeedsConfirm {
				hasProtected = true
				break
			}
		}

		if m.noConfirm && !hasProtected {
			m.confirmed = true
			m.state = "moving"
			m.currentIndex = 0
			return m, moveNextFileToCache(m.fileInfos, m.currentIndex, m.config)
		}

		m.state = "confirming"
		return m, nil

	case fileMoveMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error moving file: %v", msg.err)
			logging.LogError("MOVE_ERROR", m.fileInfos[m.currentIndex].Path, msg.err, m.config)
			return m, nil
		}

		if msg.item.ID != "" {
			m.processedItems = append(m.processedItems, msg.item)
			m.processedFiles++

			if msg.item.IsDirectory {
				m.stats.ProcessedDirs++
			} else {
				m.stats.ProcessedFiles++
			}
			m.stats.ProcessedSize += msg.item.Size
		}

		m.currentIndex++

		nextIndex := findNextValidFile(m.fileInfos, m.currentIndex)
		if nextIndex != -1 {
			m.currentIndex = nextIndex
			return m, moveNextFileToCache(m.fileInfos, m.currentIndex, m.config)
		}

		m.state = "cleanup"
		return m, cleanupOldFiles(m.config)

	case cleanupMsg:
		m.state = "done"
		return m, nil

	case clearMsg:
		if msg.err != nil {
			m.state = "error"
			m.errorMsg = fmt.Sprintf("Error clearing cache: %v", msg.err)
			return m, nil
		}
		m.state = "done"
		return m, nil

	case progress.FrameMsg:
		if !m.safeMode && m.config.UI.Progress.Enabled {
			progressModel, cmd := m.progress.Update(msg)
			m.progress = progressModel.(progress.Model)
			cmds = append(cmds, cmd)
		}

	case errorMsg:
		m.state = "error"
		m.errorMsg = string(msg)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.safeMode {
		return m.viewSafe()
	}
	return m.viewStyled()
}

// Safe view with no styling
func (m Model) viewSafe() string {
	switch m.state {
	case "checking":
		return "Checking files and directories...\n"

	case "confirming":
		var content strings.Builder
		content.WriteString("Are you sure you want to delete the following items?\n\n")

		validCount := 0
		for _, info := range m.fileInfos {
			if info.Exists {
				validCount++
				if info.IsDirectory {
					content.WriteString(fmt.Sprintf("  DIR  %s", info.Path))
					if info.FileCount > 0 {
						content.WriteString(fmt.Sprintf(" (%d items)", info.FileCount))
					}
				} else {
					content.WriteString(fmt.Sprintf("  FILE %s", info.Path))
				}
				content.WriteString("\n")
			} else {
				content.WriteString(fmt.Sprintf("  ERR  %s (does not exist)\n", info.Path))
			}
		}

		if validCount > 0 {
			content.WriteString(fmt.Sprintf("\nTotal items: %d | Recoverable for %d days\n", validCount, m.config.Cache.Days))
			content.WriteString("\nPress 'y' to confirm, 'n' to cancel, or 'q' to quit\n")
		} else {
			content.WriteString("\nPress 'n' to cancel or 'q' to quit\n")
		}

		return content.String()

	case "moving":
		if m.currentIndex < len(m.fileInfos) {
			return fmt.Sprintf("Moving '%s' to cache... (%d/%d)\n",
				m.fileInfos[m.currentIndex].Path, m.processedFiles+1, m.totalFiles)
		}
		return "Moving files to cache...\n"

	case "cleanup":
		return "Cleaning up old cached files...\n"

	case "clearing":
		return "Clearing all cached files...\n"

	case "done":
		if m.clearAll {
			return "All cached files cleared!\nPress Enter or 'q' to exit\n"
		}
		content := fmt.Sprintf("Successfully processed %d item(s)!\n", len(m.processedItems))
		if len(m.processedItems) > 0 {
			deleteAfter := m.processedItems[0].DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
			content += fmt.Sprintf("Permanent deletion after: %s\n", deleteAfter.Format("2006-01-02 15:04"))
		}
		content += "Press Enter or 'q' to exit\n"
		return content

	case "error":
		return fmt.Sprintf("Error: %s\nPress Enter or 'q' to exit\n", m.errorMsg)
	}

	return ""
}

// Original styled view (simplified to avoid the complex styling that's causing issues)
func (m Model) viewStyled() string {
	// For now, fallback to safe view to avoid styling issues
	// Once we confirm safe mode works, we can gradually re-enable styling
	return m.viewSafe()
}

// Commands (unchanged from original)
func checkFilesExist(filenames []string, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		fileInfos := make([]models.FileInfo, len(filenames))
		for i, filename := range filenames {
			fileInfos[i] = filesystem.CheckFileInfo(filename, cfg)
		}
		return filesExistMsg{fileInfos: fileInfos}
	}
}

func findNextValidFile(fileInfos []models.FileInfo, startIndex int) int {
	for i := startIndex; i < len(fileInfos); i++ {
		if fileInfos[i].Exists {
			return i
		}
	}
	return -1
}

func moveNextFileToCache(fileInfos []models.FileInfo, index int, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if index >= len(fileInfos) || !fileInfos[index].Exists {
			return fileMoveMsg{item: models.DeletedItem{}, err: nil}
		}
		filename := fileInfos[index].Path
		return moveFileToCache(filename, cfg)()
	}
}

func moveFileToCache(filename string, cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		item, err := filesystem.MoveFileToCache(filename, cfg)
		if err != nil {
			return fileMoveMsg{err: err}
		}

		if err := filesystem.AddToIndex(item, cfg); err != nil {
			return fileMoveMsg{err: err}
		}

		logging.LogOperation("DELETE", item, cfg)
		return fileMoveMsg{item: item, err: nil}
	}
}

func cleanupOldFiles(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := filesystem.CleanupOldFiles(cfg); err != nil {
			return errorMsg(fmt.Sprintf("Error during cleanup: %v", err))
		}
		return cleanupMsg{}
	}
}

func clearAllCache(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := filesystem.ClearAllCache(cfg); err != nil {
			return clearMsg{err: err}
		}
		return clearMsg{err: nil}
	}
}

func formatSize(bytes int64) string {
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
