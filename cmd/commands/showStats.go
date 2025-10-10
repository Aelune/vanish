package command

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"vanish/internal/helpers"
	"vanish/internal/types"
)

type statsModel struct {
	config        types.Config
	index         types.Index
	styles        types.ThemeStyles
	totalSize     int64
	fileCount     int
	dirCount      int
	expiredCount  int
	retentionDays int
	cacheDir      string
	width         int
	height        int
	// Additional stats
	largestItem     string
	largestItemSize int64
	oldestItem      string
	oldestItemDate  time.Time
	newestItem      string
	newestItemDate  time.Time
	avgFileSize     int64
}

type statsLoaded struct {
	index types.Index
	err   error
}

func (m *statsModel) Init() tea.Cmd {
	return loadStatsCmd(m.config)
}

func loadStatsCmd(config types.Config) tea.Cmd {
	return func() tea.Msg {
		index, err := helpers.LoadIndex(config)
		return statsLoaded{index: index, err: err}
	}
}

func (m *statsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case statsLoaded:
		if msg.err != nil {
			return m, tea.Quit
		}
		m.index = msg.index
		m.calculateStats()
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *statsModel) calculateStats() {
	if len(m.index.Items) == 0 {
		return
	}

	cutoff := time.Now().Add(-time.Duration(m.retentionDays) * 24 * time.Hour)
	m.oldestItemDate = time.Now()
	m.newestItemDate = time.Time{}

	for _, item := range m.index.Items {
		m.totalSize += item.Size

		if item.IsDirectory {
			m.dirCount++
		} else {
			m.fileCount++
		}

		if item.DeleteDate.Before(cutoff) {
			m.expiredCount++
		}

		// Track largest item
		if item.Size > m.largestItemSize {
			m.largestItemSize = item.Size
			m.largestItem = item.OriginalPath
		}

		// Track oldest item
		if item.DeleteDate.Before(m.oldestItemDate) {
			m.oldestItemDate = item.DeleteDate
			m.oldestItem = item.OriginalPath
		}

		// Track newest item
		if item.DeleteDate.After(m.newestItemDate) {
			m.newestItemDate = item.DeleteDate
			m.newestItem = item.OriginalPath
		}
	}

	// Calculate average file size
	if m.fileCount > 0 {
		m.avgFileSize = m.totalSize / int64(m.fileCount)
	}
}

func (m *statsModel) View() string {
	if len(m.index.Items) == 0 {
		emptyMsg := m.styles.Warning.Render("📦 Cache is empty")
		help := m.styles.Help.Render("No items currently in the vanish cache")
		return m.styles.Root.Render(lipgloss.JoinVertical(lipgloss.Left, emptyMsg, help))
	}

	var sections []string

	// Title
	title := m.styles.Title.Render("📊 Vanish Cache Statistics")
	sections = append(sections, title)

	// Main stats box
	statsContent := m.buildStatsContent()
	statsBox := m.styles.Border.Render(statsContent)
	sections = append(sections, statsBox)

	// Retention info
	retentionInfo := m.buildRetentionInfo()
	sections = append(sections, retentionInfo)

	// Footer with help text
	if m.expiredCount > 0 {
		footer := m.buildFooter()
		sections = append(sections, footer)
	}

	return m.styles.Root.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m *statsModel) buildStatsContent() string {
	var rows []string

	// Cache directory
	cacheDirLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Cache Location:")
	cacheDirValue := m.styles.Filename.Render(m.cacheDir)
	rows = append(rows, fmt.Sprintf("%s %s", cacheDirLabel, cacheDirValue))

	rows = append(rows, "") // Spacer

	// Total items with icon
	totalIcon := m.styles.IconStyle.Render("📦")
	totalLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render("Total Items:")
	totalValue := m.styles.Header.Render(fmt.Sprintf("%d", len(m.index.Items)))
	rows = append(rows, fmt.Sprintf("%s %s %s", totalIcon, totalLabel, totalValue))

	// Files
	fileIcon := m.styles.IconStyle.Render("📄")
	fileLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("  Files:")
	fileValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(fmt.Sprintf("%d", m.fileCount))
	rows = append(rows, fmt.Sprintf("%s %s %s", fileIcon, fileLabel, fileValue))

	// Directories
	dirIcon := m.styles.IconStyle.Render("📁")
	dirLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("  Directories:")
	dirValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(fmt.Sprintf("%d", m.dirCount))
	rows = append(rows, fmt.Sprintf("%s %s %s", dirIcon, dirLabel, dirValue))

	rows = append(rows, "") // Spacer

	// Total size (no border)
	sizeIcon := m.styles.IconStyle.Render("💾")
	sizeLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render("Total Size:")
	sizeValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Success)).Bold(true).Render(helpers.FormatBytes(m.totalSize))
	rows = append(rows, fmt.Sprintf("%s %s %s", sizeIcon, sizeLabel, sizeValue))

	// Average file size
	if m.fileCount > 0 {
		avgIcon := m.styles.IconStyle.Render("📊")
		avgLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("  Avg File Size:")
		avgValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(helpers.FormatBytes(m.avgFileSize))
		rows = append(rows, fmt.Sprintf("%s %s %s", avgIcon, avgLabel, avgValue))
	}

	rows = append(rows, "") // Spacer

	// Largest item
	largestIcon := m.styles.IconStyle.Render("🔝")
	largestLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render("Largest Item:")
	largestSize := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).Render(helpers.FormatBytes(m.largestItemSize))
	rows = append(rows, fmt.Sprintf("%s %s %s", largestIcon, largestLabel, largestSize))

	// Truncate path if too long
	maxPathLen := 45
	displayPath := m.largestItem
	if len(displayPath) > maxPathLen {
		displayPath = "..." + displayPath[len(displayPath)-maxPathLen+3:]
	}
	pathStyle := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Italic(true)
	rows = append(rows, "  "+pathStyle.Render(displayPath))

	rows = append(rows, "") // Spacer

	// Time-based stats
	timeIcon := m.styles.IconStyle.Render("🕐")
	timeLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render("Time Stats:")
	rows = append(rows, fmt.Sprintf("%s %s", timeIcon, timeLabel))

	// Oldest item
	oldestLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("  Oldest:")
	oldestAge := time.Since(m.oldestItemDate)
	oldestValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(formatDuration(oldestAge) + " ago")
	rows = append(rows, fmt.Sprintf("%s %s", oldestLabel, oldestValue))

	// Newest item
	newestLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("  Newest:")
	newestAge := time.Since(m.newestItemDate)
	newestValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(formatDuration(newestAge) + " ago")
	rows = append(rows, fmt.Sprintf("%s %s", newestLabel, newestValue))

	rows = append(rows, "") // Spacer

	// Retention period
	retentionIcon := m.styles.IconStyle.Render("⏰")
	retentionLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render("Retention Period:")
	retentionValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).
		Bold(true).
		Render(fmt.Sprintf("%d days", m.retentionDays))
	rows = append(rows, fmt.Sprintf("%s %s %s", retentionIcon, retentionLabel, retentionValue))

	// Expired items
	var expiredLine string
	if m.expiredCount > 0 {
		expiredIcon := m.styles.IconStyle.Foreground(lipgloss.Color(m.config.UI.Colors.Warning)).Render("⚠️")
		expiredLabel := m.styles.Warning.Render("Expired Items:")
		expiredValue := m.styles.Warning.Bold(true).Render(fmt.Sprintf("%d", m.expiredCount))
		expiredLine = fmt.Sprintf("%s %s %s", expiredIcon, expiredLabel, expiredValue)
	} else {
		expiredIcon := m.styles.IconStyle.Foreground(lipgloss.Color(m.config.UI.Colors.Success)).Render("✓")
		expiredLabel := m.styles.StatusGood.Render("Expired Items:")
		expiredValue := m.styles.StatusGood.Render("0")
		expiredLine = fmt.Sprintf("%s %s %s", expiredIcon, expiredLabel, expiredValue)
	}
	rows = append(rows, expiredLine)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *statsModel) buildRetentionInfo() string {
	percentage := float64(m.expiredCount) / float64(len(m.index.Items)) * 100

	var statusMsg string
	var barStyle lipgloss.Style

	if m.expiredCount == 0 {
		statusMsg = m.styles.StatusGood.Render("✓ Cache is clean - no expired items")
		barStyle = m.styles.StatusGood
	} else if percentage < 25 {
		statusMsg = m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).
			Render(fmt.Sprintf("ℹ️  %.1f%% of items have expired", percentage))
		barStyle = m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary))
	} else if percentage < 50 {
		statusMsg = m.styles.Warning.Render(fmt.Sprintf("⚠️  %.1f%% of items have expired", percentage))
		barStyle = m.styles.Warning
	} else {
		statusMsg = m.styles.Error.Render(fmt.Sprintf("❗ %.1f%% of items have expired", percentage))
		barStyle = m.styles.Error
	}

	// Simple progress bar
	barWidth := 40
	filled := int(float64(barWidth) * percentage / 100)
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	styledBar := barStyle.Render(bar)

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		statusMsg,
		m.styles.Progress.Render(styledBar),
	)
}

func (m *statsModel) buildFooter() string {
	command := m.styles.Filename.Render(fmt.Sprintf("vx --purge %d", m.retentionDays))
	helpText := m.styles.Help.Render(fmt.Sprintf("Run %s to clean up expired items", command))

	actionBox := m.styles.Info.
		Foreground(lipgloss.Color(m.config.UI.Colors.Warning)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.config.UI.Colors.Warning)).
		Padding(0, 1).
		Render("💡 " + helpText)

	return lipgloss.NewStyle().MarginTop(1).Render(actionBox)
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// ShowStats displays cache statistics using a beautiful Bubble Tea TUI
func ShowStats(config types.Config) error {
	styles := helpers.CreateThemeStyles(config)

	m := &statsModel{
		config:        config,
		styles:        styles,
		retentionDays: config.Cache.Days,
		cacheDir:      helpers.ExpandPath(config.Cache.Directory),
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()

	if err != nil {
		return fmt.Errorf("error running stats display: %v", err)
	}

	// Check if there was an error loading the index
	if fm, ok := finalModel.(*statsModel); ok {
		if len(fm.index.Items) == 0 && fm.totalSize == 0 {
			// Index might have failed to load, but we still showed the view
			return nil
		}
	}

	return nil
}
