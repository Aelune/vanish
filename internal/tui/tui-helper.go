package tui

import (
	// "github.com/charmbracelet/bubbles/progress"
	// "github.com/charmbracelet/bubbletea"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"path/filepath"
	"strings"
	"time"
	// "os"

	"vanish/internal/types"
)

// ErrorMsg represents an error message in the TUI
type ErrorMsg string

func (m *Model) getFileIcon(isDirectory bool) string {
	if m.Config.UI.Progress.ShowEmoji {
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

func (m *Model) getFileTypeString(isDirectory bool) string {
	if isDirectory {
		return "directory"
	}
	return "file"
}

func (m *Model) renderDeleteConfirmation(content *strings.Builder, contentWidth int) {
	content.WriteString(m.Styles.Question.Render("Are you sure you want to delete the following items?"))
	content.WriteString("\n")

	validCount, invalidCount, totalFileCount := m.analyzeFileInfos()
	listContent := m.buildFileInfosList(validCount, invalidCount, &totalFileCount)

	content.WriteString(m.Styles.List.Render(listContent))

	if invalidCount > 0 {
		m.renderInvalidFilesWarning(content, invalidCount)
	}

	if validCount > 0 {
		m.renderDeleteSummary(content, validCount, totalFileCount, contentWidth)
	}
}

func (m *Model) analyzeFileInfos() (validCount, invalidCount, totalFileCount int) {
	for _, info := range m.FileInfos {
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

func (m *Model) buildFileInfosList(_, _ int, totalFileCount *int) string {
	var listContent strings.Builder

	for _, info := range m.FileInfos {
		if info.Exists {
			m.appendValidFileInfo(&listContent, info, totalFileCount)
		} else {
			m.appendInvalidFileInfo(&listContent, info)
		}
	}

	return listContent.String()
}

func (m *Model) appendValidFileInfo(listContent *strings.Builder, info types.FileInfo, _ *int) {
	icon := m.getFileIcon(info.IsDirectory)
	listContent.WriteString(icon)
	listContent.WriteString(m.Styles.Filename.Render(info.Path))

	if info.IsDirectory {
		inlineInfoStyle := m.Styles.Info.Border(lipgloss.Border{}).Padding(0)
		if info.FileCount > 0 {
			listContent.WriteString(inlineInfoStyle.Render(fmt.Sprintf(" (%d items)", info.FileCount)))
		} else {
			listContent.WriteString(inlineInfoStyle.Render(" (empty)"))
		}
	}

	listContent.WriteString("\n")
}

func (m *Model) appendInvalidFileInfo(listContent *strings.Builder, info types.FileInfo) {
	icon := "ERR "
	if m.Config.UI.Progress.ShowEmoji {
		icon = "❌ "
	}

	listContent.WriteString(icon)
	listContent.WriteString(m.Styles.StatusBad.Render(info.Path))
	listContent.WriteString(m.Styles.Warning.Render(" (does not exist)"))
}

func (m *Model) renderInvalidFilesWarning(content *strings.Builder, invalidCount int) {
	content.WriteString("\n")
	warningText := fmt.Sprintf("⚠ Warning: %d file(s) will be skipped", invalidCount)
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.Config.UI.Colors.Warning)).
		Bold(true)
	content.WriteString(warningStyle.Render(warningText))
}

func (m *Model) renderDeleteSummary(content *strings.Builder, validCount, totalFileCount, contentWidth int) {
	content.WriteString("\n")

	infoText := fmt.Sprintf("Total items to delete: %d", validCount)
	if totalFileCount > validCount {
		infoText += fmt.Sprintf(" | Files affected: %d", totalFileCount)
	}
	infoText += fmt.Sprintf(" | Recoverable for %d days", m.Config.Cache.Days)

	infoStyle := m.Styles.Info.MaxWidth(contentWidth).Align(lipgloss.Left)
	content.WriteString(infoStyle.Render(infoText))
}

func (m *Model) renderMovingState(content *strings.Builder, contentWidth int) {
	statusText := m.buildProgressStatusText("Moving", "📦")
	m.renderProgressState(content, statusText, contentWidth)
}

func (m *Model) renderRestoringState(content *strings.Builder, contentWidth int) {
	statusText := m.buildRestoreStatusText()
	m.renderProgressState(content, statusText, contentWidth)
}

func (m *Model) buildProgressStatusText(action, emoji string) string {
	if m.CurrentIndex < len(m.FileInfos) {
		currentFile := m.FileInfos[m.CurrentIndex]
		fileType := m.getFileTypeString(currentFile.IsDirectory)

		emojiPrefix := ""
		if m.Config.UI.Progress.ShowEmoji {
			emojiPrefix = emoji + " "
		}

		return fmt.Sprintf("%s%s %s '%s' to safe cache... (%d/%d)",
			emojiPrefix, action, fileType, currentFile.Path, m.ProcessedFiles+1, m.TotalFiles)
	}

	fallback := fmt.Sprintf("%s files to safe cache...", action)
	if m.Config.UI.Progress.ShowEmoji {
		fallback = emoji + " " + fallback
	}
	return fallback
}

func (m *Model) buildRestoreStatusText() string {
	if m.CurrentIndex < len(m.RestoreItems) {
		currentItem := m.RestoreItems[m.CurrentIndex]
		fileType := m.getFileTypeString(currentItem.IsDirectory)

		emojiPrefix := ""
		if m.Config.UI.Progress.ShowEmoji {
			emojiPrefix = "♻️ "
		}

		return fmt.Sprintf("%sRestoring %s '%s'... (%d/%d)",
			emojiPrefix, fileType, currentItem.OriginalPath, m.ProcessedFiles+1, len(m.RestoreItems))
	}

	fallback := "Restoring files from cache..."
	if m.Config.UI.Progress.ShowEmoji {
		fallback = "♻️ " + fallback
	}
	return fallback
}

func (m *Model) renderProgressState(content *strings.Builder, statusText string, contentWidth int) {
	statusStyle := m.Styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MaxWidth(contentWidth)
	content.WriteString(statusStyle.Render(statusText))
	content.WriteString("\n")
	content.WriteString(m.Styles.Progress.Render(m.Progress.View()))
}

func (m *Model) renderCleanupState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🧹", "Cleaning up old cached files...")
}

func (m *Model) renderClearingState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🗑️", "Clearing all cached files...")
}

func (m *Model) renderPurgingState(content *strings.Builder) {
	m.renderSimpleProgressState(content, "🔥", "Purging old cached files...")
}

func (m *Model) renderSimpleProgressState(content *strings.Builder, emoji, message string) {
	if m.Config.UI.Progress.ShowEmoji {
		content.WriteString(emoji + " ")
	}
	content.WriteString(message + "\n")
	content.WriteString(m.Styles.Progress.Render(m.Progress.View()))
}

func (m *Model) renderDoneState(content *strings.Builder, contentWidth int) {
	successMsg := m.buildSuccessMessage()
	content.WriteString(m.Styles.Success.Render(successMsg))
	content.WriteString("\n")

	if m.shouldShowItemDetails() {
		m.renderItemDetails(content, contentWidth)
	}

	content.WriteString(m.Styles.Progress.Render(m.Progress.View()))
	content.WriteString("\n")
	content.WriteString(m.Styles.Help.Render("Press Enter or 'q' to exit"))
}

func (m *Model) buildSuccessMessage() string {
	var successMsg string
	emoji := ""

	if m.Config.UI.Progress.ShowEmoji {
		emoji = "✅ "
	}

	switch m.Operation {
	case "clear":
		if m.Config.UI.Progress.ShowEmoji {
			successMsg = "✅ All cached files cleared!"
		} else {
			successMsg = "SUCCESS: All cached files cleared!"
		}
	case "purge":
		if m.Config.UI.Progress.ShowEmoji {
			successMsg = fmt.Sprintf("✅ Purged %d old cached files!", m.ProcessedFiles)
		} else {
			successMsg = fmt.Sprintf("SUCCESS: Purged %d old cached files!", m.ProcessedFiles)
		}
	case "restore":
		successMsg = fmt.Sprintf("%sSuccessfully restored %d item(s)!", emoji, len(m.ProcessedItems))
	default:
		successMsg = fmt.Sprintf("%sSuccessfully processed %d item(s)!", emoji, len(m.ProcessedItems))
	}

	return successMsg
}

func (m *Model) shouldShowItemDetails() bool {
	return (m.Operation == "delete" || m.Operation == "restore") && len(m.ProcessedItems) > 0
}

func (m *Model) renderItemDetails(content *strings.Builder, contentWidth int) {
	detailsBuilder := m.buildItemDetailsText()
	content.WriteString(m.Styles.List.Render(detailsBuilder))

	if m.Operation == "delete" && len(m.ProcessedItems) > 0 {
		m.renderDeletionInfo(content, contentWidth)
	}
}

func (m *Model) buildItemDetailsText() string {
	var detailsBuilder strings.Builder

	maxItems := 5
	for i, item := range m.ProcessedItems {
		if i >= maxItems {
			break
		}

		if m.Operation == "restore" {
			detailsBuilder.WriteString(fmt.Sprintf("• %s ← %s\n",
				m.Styles.Filename.Render(item.OriginalPath), "cache"))
		} else {
			detailsBuilder.WriteString(fmt.Sprintf("• %s → %s\n",
				m.Styles.Filename.Render(item.OriginalPath), filepath.Base(item.CachePath)))
		}
	}

	if len(m.ProcessedItems) > maxItems {
		infoStyle := m.Styles.Info.Border(lipgloss.Border{}).Padding(0)
		detailsBuilder.WriteString(infoStyle.Render(fmt.Sprintf("... and %d more item(s)", len(m.ProcessedItems)-maxItems)))
		detailsBuilder.WriteString("\n")
	}

	return detailsBuilder.String()
}

func (m *Model) renderDeletionInfo(content *strings.Builder, contentWidth int) {
	deleteAfter := m.ProcessedItems[0].DeleteDate.Add(time.Duration(m.Config.Cache.Days) * 24 * time.Hour)
	infoStyle := m.Styles.Info.
		Border(lipgloss.Border{}).
		Padding(0).
		MaxWidth(contentWidth)

	content.WriteString("\n")
	content.WriteString(infoStyle.Render(fmt.Sprintf("Will be permanently deleted after: %s", deleteAfter.Format("2006-01-02 15:04:05"))))
	content.WriteString("\n")
}

func (m *Model) renderErrorState(content *strings.Builder) {
	errorMsg := "ERROR"
	if m.Config.UI.Progress.ShowEmoji {
		errorMsg = "❌ Error"
	}

	content.WriteString(m.Styles.Error.Render(errorMsg))
	content.WriteString("\n")
	content.WriteString(m.ErrorMsg)
	content.WriteString("\n")
	content.WriteString(m.Styles.Help.Render("Press Enter or 'q' to exit"))
}
