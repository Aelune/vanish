// Package tui manages all the tui related code
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"vanish/internal/helpers"
	"vanish/internal/types"
	"vanish/internal/config"
)

// Model defines the state and data used by the TUI.
type Model struct {
	Filenames      []string
	FileInfos      []types.FileInfo
	CurrentIndex   int
	State          string
	Progress       progress.Model
	ProgressVal    float64
	Confirmed      bool
	ErrorMsg       string
	Config         types.Config
	Styles         types.ThemeStyles
	ProcessedItems []types.DeletedItem
	ClearAll       bool
	TotalFiles     int
	ProcessedFiles int
	NoConfirm      bool
	Operation      string // "delete", "restore", "clear", "purge"
	RestoreItems   []types.DeletedItem
}


func InitialModel(filenames []string, operation string, noConfirm bool) (*Model, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	prog := helpers.SetUpProgress(cfg)
	styles := helpers.CreateThemeStyles(cfg)

	// Check if no_confirm is set in config and not overridden by flag
	if cfg.Cache.NoConfirm && !noConfirm {
		noConfirm = true
	}

	return &Model{
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

// Init initializes the TUI model and triggers the initial
// command based on the selected operation.
func (m *Model) Init() tea.Cmd {
	switch m.Operation {
	case "clear":
		m.State = "clearing"
		return tea.Batch(
			m.Progress.SetPercent(0.1),
			helpers.ClearAllCache(m.Config),
		)
	case "purge":
		m.State = "purging"
		return tea.Batch(
			m.Progress.SetPercent(0.1),
			helpers.PurgeOldFiles(m.Config, m.Filenames[0]),
		)
	case "restore":
		m.State = "checking"
		return tea.Batch(
			helpers.CheckRestoreItems(m.Filenames, m.Config),
			m.Progress.SetPercent(0.1),
		)
	default: // delete
		return tea.Batch(
			helpers.CheckFilesExist(m.Filenames),
			m.Progress.SetPercent(0.1),
		)
	}
}

// Update handles incoming messages and updates the TUI model state accordingly.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y", "Y":
			if m.State == "confirming" {
				m.Confirmed = true
				if m.Operation == "restore" {
					m.State = "restoring"
				} else {
					m.State = "moving"
				}
				m.CurrentIndex = 0
				return m, tea.Batch(
					m.Progress.SetPercent(0.3),
					processNextItem(m),
				)
			}
		case "n", "N":
			if m.State == "confirming" {
				return m, tea.Quit
			}
		case "enter":
			if m.State == "done" || m.State == "error" {
				return m, tea.Quit
			}
		}

	case types.FilesExistMsg:
		m.FileInfos = msg.FileInfos
		validFiles := 0
		for _, info := range m.FileInfos {
			if info.Exists {
				validFiles++
			}
		}

		if validFiles == 0 {
			m.State = "error"
			m.ErrorMsg = "No valid files or directories found"
			return m, nil
		}

		if m.NoConfirm {
			m.Confirmed = true
			m.State = "moving"
			m.CurrentIndex = 0
			return m, tea.Batch(
				m.Progress.SetPercent(0.3),
				processNextItem(m),
			)
		}
		m.State = "confirming"
		return m, m.Progress.SetPercent(0.2)

	case types.RestoreItemsMsg:
		m.RestoreItems = msg.Items
		if len(m.RestoreItems) == 0 {
			m.State = "error"
			m.ErrorMsg = "No matching items found in cache for restoration"
			return m, nil
		}

		if m.NoConfirm {
			m.Confirmed = true
			m.State = "restoring"
			m.CurrentIndex = 0
			return m, tea.Batch(
				m.Progress.SetPercent(0.3),
				processNextItem(m),
			)
		}
		m.State = "confirming"
		return m, m.Progress.SetPercent(0.2)

	case types.FileMoveMsg:
		if msg.Err != nil {
			m.State = "error"
			m.ErrorMsg = fmt.Sprintf("Error processing item: %v", msg.Err)
			return m, nil
		}

		if msg.Item.ID != "" {
			m.ProcessedItems = append(m.ProcessedItems, msg.Item)
			m.ProcessedFiles++
		}

		// Find the next valid file index, starting from current + 1
		nextIndex := helpers.FindNextValidFile(m.FileInfos, m.CurrentIndex+1)

		// Update progress based on processed files vs total valid files
		validFileCount := helpers.CountValidFiles(m.FileInfos)
		progressPercent := 0.3 + (float64(m.ProcessedFiles)/float64(validFileCount))*0.4

		// Check if we have more valid items to process
		if nextIndex != -1 {
			m.CurrentIndex = nextIndex
			return m, tea.Batch(
				m.Progress.SetPercent(progressPercent),
				processNextItem(m),
			)
		}

		// All items processed, move to cleanup
		m.State = "cleanup"
		return m, tea.Batch(
			m.Progress.SetPercent(0.7),
			cleanupOldFiles(m.Config),
		)

	case types.RestoreMsg:
		if msg.Err != nil {
			m.State = "error"
			m.ErrorMsg = fmt.Sprintf("Error restoring item: %v", msg.Err)
			return m, nil
		}

		if msg.Item.ID != "" {
			m.ProcessedItems = append(m.ProcessedItems, msg.Item)
			m.ProcessedFiles++
		}

		m.CurrentIndex++

		// Update progress
		progressPercent := 0.3 + (float64(m.CurrentIndex)/float64(len(m.RestoreItems)))*0.4

		// Check if we have more items to restore
		if m.CurrentIndex < len(m.RestoreItems) {
			return m, tea.Batch(
				m.Progress.SetPercent(progressPercent),
				processNextItem(m),
			)
		}

		// All items restored
		m.State = "done"
		return m, m.Progress.SetPercent(1.0)

	case types.CleanupMsg:
		m.State = "done"
		return m, m.Progress.SetPercent(1.0)

	case types.ClearMsg:
		if msg.Err != nil {
			m.State = "error"
			m.ErrorMsg = fmt.Sprintf("Error clearing cache: %v", msg.Err)
			return m, nil
		}
		m.State = "done"
		return m, m.Progress.SetPercent(1.0)

	case types.PurgeMsg:
		if msg.Err != nil {
			m.State = "error"
			m.ErrorMsg = fmt.Sprintf("Error purging cache: %v", msg.Err)
			return m, nil
		}
		m.ProcessedFiles = msg.PurgedCount
		m.State = "done"
		return m, m.Progress.SetPercent(1.0)

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case types.ErrorMsg:
		m.State = "error"
		m.ErrorMsg = string(msg)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI based on the model's state.
func (m *Model) View() string {
	var content strings.Builder

	// Get terminal dimensions for responsive design
	// termWidth, termHeight := lipgloss.Size(m.styles.Root.String())
	// if termWidth == 0 {
	// 	termWidth = 80 // fallback width
	// }
	// if termHeight == 0 {
	// 	termHeight = 24 // fallback height
	// }
	termWidth, _ := helpers.GetTerminalSize()
	// Calculate content width (leaving some margin)
	contentWidth := termWidth - 8 // 4 chars margin on each side

	switch m.State {
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

	return m.Styles.Root.Render(content.String())
}

func processNextItem(m *Model) tea.Cmd {
	if m.Operation == "restore" {
		if m.CurrentIndex >= len(m.RestoreItems) {
			return nil
		}
		return restoreFromCache(m.RestoreItems[m.CurrentIndex], m.Config)
	}
	// Make sure we have a valid index
	if m.CurrentIndex < 0 || m.CurrentIndex >= len(m.FileInfos) {
		return nil
	}
	// Make sure the file at current index exists
	if !m.FileInfos[m.CurrentIndex].Exists {
		return nil
	}
	return moveFileToCache(m.FileInfos[m.CurrentIndex].Path, m.Config)
}

func restoreFromCache(item types.DeletedItem, config types.Config) tea.Cmd {
	return func() tea.Msg {
		// Check if cache file exists
		if _, err := os.Stat(item.CachePath); os.IsNotExist(err) {
			return types.RestoreMsg{Err: fmt.Errorf("cached file not found: %s", item.CachePath)}
		}

		// Create directory for original path if needed
		originalDir := filepath.Dir(item.OriginalPath)
		if err := os.MkdirAll(originalDir, 0755); err != nil {
			return types.RestoreMsg{Err: fmt.Errorf("failed to create directory %s: %v", originalDir, err)}
		}

		// Check if original path already exists
		if _, err := os.Stat(item.OriginalPath); !os.IsNotExist(err) {
			return types.RestoreMsg{Err: fmt.Errorf("destination already exists: %s", item.OriginalPath)}
		}

		// Move file back
		if item.IsDirectory {
			if err := helpers.MoveDirectory(item.CachePath, item.OriginalPath); err != nil {
				return types.RestoreMsg{Err: err}
			}
		} else {
			if err := helpers.MoveFile(item.CachePath, item.OriginalPath); err != nil {
				return types.RestoreMsg{Err: err}
			}
		}

		// Remove from index
		if err := helpers.RemoveFromIndex(item.ID, config); err != nil {
			// Log error but don't fail the restore
			if config.Logging.Enabled {
				logDir := helpers.ExpandPath(config.Logging.Directory)
				err := os.MkdirAll(logDir, 0755)
				if err != nil {
					return nil
				}
				logPath := filepath.Join(logDir, "vanish.log")
				logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					defer logFile.Close()
					_, err := logFile.WriteString(fmt.Sprintf("%s ERROR Failed to remove from index: %s\n",
						time.Now().Format("2006-01-02 15:04:05"), item.ID))
					if err != nil {
						return nil
					}
				}
			}
		}

		// Log the restore operation
		if config.Logging.Enabled {
			helpers.LogOperation("RESTORE", item, config)
		}
		// if config.Notifications.NotifySuccess {
		// 	helpers.SendNotification("Vanish", fmt.Sprintf("Restored %s", filepath.Base(item.OriginalPath)), config)
		// }

		return types.RestoreMsg{Item: item, Err: nil}
	}
}

func moveFileToCache(filename string, config types.Config) tea.Cmd {
	return func() tea.Msg {

		// Ensure cache directory exists
		cacheDir := helpers.ExpandPath(config.Cache.Directory)
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return types.FileMoveMsg{Err: err}
		}

		// Get file info
		stat, err := os.Stat(filename)
		if err != nil {
			return types.FileMoveMsg{Err: err}
		}

		// Get absolute path
		absPath, err := filepath.Abs(filename)
		if err != nil {
			return types.FileMoveMsg{Err: err}
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
		// Handle directories
		if isDir {
			fileCount, _ = helpers.CountFilesInDirectory(filename)
			size, _ = helpers.GetDirectorySize(filename)

			if err := helpers.MoveDirectory(filename, cachePath); err != nil {
				return types.FileMoveMsg{Err: err}
			}
		} else {
			// Handle files
			if err := helpers.MoveFile(filename, cachePath); err != nil {
				return types.FileMoveMsg{Err: err}
			}
		}

		// Create enhanced deleted item
		item := types.DeletedItem{
			ID:           id,
			OriginalPath: absPath,
			DeleteDate:   now,
			CachePath:    cachePath,
			IsDirectory:  isDir,
			FileCount:    fileCount,
			Size:         size,
		}

		// Update index
		if err := helpers.AddToIndex(item, config); err != nil {
			return types.FileMoveMsg{Err: err}
		}

		// Log the operation
		if config.Logging.Enabled {
			helpers.LogOperation("DELETE", item, config)
		}

		// Send notification if enabled
		// if config.Notifications.NotifySuccess {
		// 	helpers.SendNotification("Vanish", fmt.Sprintf("Moved %s to cache", filepath.Base(filename)), config)
		// }

		return types.FileMoveMsg{Item: item, Err: nil}
	}
}

func cleanupOldFiles(config types.Config) tea.Cmd {
	return func() tea.Msg {
		cutoffDays := time.Duration(config.Cache.Days) * 24 * time.Hour
		cutoff := time.Now().Add(-cutoffDays)

		index, err := helpers.LoadIndex(config)
		if err != nil {
			return types.ErrorMsg(fmt.Sprintf("Error loading index: %v", err))
		}

		var remainingItems []types.DeletedItem
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
					helpers.LogOperation("CLEANUP", item, config)
				}
			} else {
				remainingItems = append(remainingItems, item)
			}
		}

		// Update index
		index.Items = remainingItems
		if err := helpers.SaveIndex(index, config); err != nil {
			return types.ErrorMsg(fmt.Sprintf("Error updating index: %v", err))
		}

		return types.CleanupMsg{}
	}
}

func (m *Model) renderCheckingState(content *strings.Builder) {
	if m.Config.UI.Progress.ShowEmoji {
		content.WriteString("🔍 ")
	}

	message := "Checking files and directories...\n"
	if m.Operation == "restore" {
		message = "Checking items for restoration...\n"
	}

	content.WriteString(message)
	content.WriteString(m.Styles.Progress.Render(m.Progress.View()))
}

func (m *Model) renderConfirmingState(content *strings.Builder, contentWidth int) {
	if m.Operation == "restore" {
		m.renderRestoreConfirmation(content)
	} else {
		m.renderDeleteConfirmation(content, contentWidth)
	}

	content.WriteString("\n")
	content.WriteString(m.Styles.Help.Render("Press 'y' to confirm, 'n' to cancel, or 'q' to quit"))
}

func (m *Model) renderRestoreConfirmation(content *strings.Builder) {
	content.WriteString(m.Styles.Question.Render("Are you sure you want to restore the following items?"))
	content.WriteString("\n")

	listContent := m.buildRestoreItemsList()
	content.WriteString(m.Styles.List.Render(listContent))

	infoStyle := m.Styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MarginTop(1)
	content.WriteString(infoStyle.Render(fmt.Sprintf("Total items to restore: %d", len(m.RestoreItems))))
}

func (m *Model) buildRestoreItemsList() string {
	var listContent strings.Builder

	for _, item := range m.RestoreItems {
		icon := m.getFileIcon(item.IsDirectory)
		listContent.WriteString(icon)
		listContent.WriteString(m.Styles.Filename.Render(item.OriginalPath))
		listContent.WriteString(m.Styles.Info.Render(fmt.Sprintf(" (deleted: %s)", item.DeleteDate.Format("2006-01-02 15:04"))))
		listContent.WriteString("\n")
	}

	return listContent.String()
}
