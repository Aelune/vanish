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

type infoModel struct {
	config        types.Config
	pattern       string
	index         types.Index
	styles        types.ThemeStyles
	matchingItems []types.DeletedItem
	width         int
	height        int
	currentPage   int
	itemsPerPage  int
}

type infoLoaded struct {
	index types.Index
	err   error
}

func (m *infoModel) Init() tea.Cmd {
	return loadInfoCmd(m.config)
}

func loadInfoCmd(config types.Config) tea.Cmd {
	return func() tea.Msg {
		index, err := helpers.LoadIndex(config)
		return infoLoaded{index: index, err: err}
	}
}

func (m *infoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case infoLoaded:
		if msg.err != nil {
			return m, tea.Quit
		}
		m.index = msg.index
		m.findMatches()
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "n", "right":
			if (m.currentPage+1)*m.itemsPerPage < len(m.matchingItems) {
				m.currentPage++
			}
		case "p", "left":
			if m.currentPage > 0 {
				m.currentPage--
			}
		}
	}
	return m, nil
}

func (m *infoModel) findMatches() {
	m.matchingItems = []types.DeletedItem{}
	for _, item := range m.index.Items {
		if strings.Contains(strings.ToLower(item.OriginalPath), strings.ToLower(m.pattern)) {
			m.matchingItems = append(m.matchingItems, item)
		}
	}
}

func (m *infoModel) View() string {
	if len(m.matchingItems) == 0 {
		return m.renderNotFound()
	}

	var sections []string

	// Title
	title := m.styles.Title.Render(fmt.Sprintf("🔍 Search Results for \"%s\"", m.pattern))
	sections = append(sections, title)

	// Summary
	summary := m.renderSummary()
	sections = append(sections, summary)

	// Items
	items := m.renderItems()
	sections = append(sections, items)

	// Pagination if needed
	if len(m.matchingItems) > m.itemsPerPage {
		pagination := m.renderPagination()
		sections = append(sections, pagination)
	}

	// Help
	help := m.renderHelp()
	sections = append(sections, help)

	return m.styles.Root.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

func (m *infoModel) renderNotFound() string {
	icon := m.styles.IconStyle.Foreground(lipgloss.Color(m.config.UI.Colors.Warning)).Render("🔍")
	notFoundMsg := m.styles.Warning.Render(fmt.Sprintf("No matches found for \"%s\"", m.pattern))

	hint := m.styles.Help.Render("💡 Try using vx --list to see all cached items")

	content := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s %s", icon, notFoundMsg),
		"",
		hint,
	)

	return m.styles.Root.Render(content)
}

func (m *infoModel) renderSummary() string {
	count := len(m.matchingItems)
	var summaryText string

	if count == 1 {
		summaryText = "Found 1 cached item"
	} else {
		summaryText = fmt.Sprintf("Found %d cached items", count)
	}

	icon := m.styles.IconStyle.Render("✨")
	summary := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).Bold(true).Render(summaryText)

	return fmt.Sprintf("%s %s", icon, summary)
}

func (m *infoModel) renderItems() string {
	start := m.currentPage * m.itemsPerPage
	end := start + m.itemsPerPage
	if end > len(m.matchingItems) {
		end = len(m.matchingItems)
	}

	var items []string
	for i := start; i < end; i++ {
		item := m.matchingItems[i]
		itemView := m.renderSingleItem(item)
		items = append(items, itemView)

		// Add separator between items (but not after the last one)
		if i < end-1 {
			items = append(items, "")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m *infoModel) renderSingleItem(item types.DeletedItem) string {
	var rows []string

	// Item header with index
	//headerText := fmt.Sprintf("📦 Item %d", index)
	//rows = append(rows, m.styles.Header.Render(headerText))
	//rows = append(rows, "")

	// ID
	idLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("ID:")
	idValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(item.ID)
	rows = append(rows, fmt.Sprintf("  %s %s", idLabel, idValue))

	// Original Path
	pathLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Original Path:")
	pathValue := m.styles.Filename.Render(item.OriginalPath)
	rows = append(rows, fmt.Sprintf("  %s %s", pathLabel, pathValue))

	// Cache Path
	cacheLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Cache Path:")
	cacheValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Italic(true).Render(item.CachePath)
	rows = append(rows, fmt.Sprintf("  %s %s", cacheLabel, cacheValue))

	rows = append(rows, "")

	// Type and Size
	var typeIcon, typeText string
	if item.IsDirectory {
		typeIcon = "📁"
		typeText = "Directory"
	} else {
		typeIcon = "📄"
		typeText = "File"
	}

	typeLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Type:")
	typeValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render(fmt.Sprintf("%s %s", typeIcon, typeText))
	rows = append(rows, fmt.Sprintf("  %s %s", typeLabel, typeValue))

	sizeLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Size:")
	sizeValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Success)).Bold(true).Render(helpers.FormatBytes(item.Size))
	rows = append(rows, fmt.Sprintf("  %s %s", sizeLabel, sizeValue))

	// File count for directories
	if item.FileCount > 0 {
		filesLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Files Inside:")
		filesValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Secondary)).Render(fmt.Sprintf("%d", item.FileCount))
		rows = append(rows, fmt.Sprintf("  %s %s", filesLabel, filesValue))
	}

	rows = append(rows, "")

	// Timing information
	deletedLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Deleted:")
	deletedValue := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Text)).Render(item.DeleteDate.Format("2006-01-02 15:04:05"))
	deletedAgo := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Italic(true).Render(fmt.Sprintf("(%s ago)", formatDuration(time.Since(item.DeleteDate))))
	rows = append(rows, fmt.Sprintf("  %s %s %s", deletedLabel, deletedValue, deletedAgo))

	// Expiry status
	expiryDate := item.DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
	daysLeft := int(time.Until(expiryDate).Hours() / 24)

	if daysLeft > 0 {
		expiryIcon := "⏰"
		expiryLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Expires:")
		expiryValue := m.styles.StatusGood.Render(expiryDate.Format("2006-01-02 15:04:05"))
		expiryLeft := m.styles.StatusGood.Render(fmt.Sprintf("(%d days left)", daysLeft))
		rows = append(rows, fmt.Sprintf("  %s %s %s %s", expiryIcon, expiryLabel, expiryValue, expiryLeft))
	} else {
		expiryIcon := "❌"
		statusLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Status:")
		statusText := m.styles.StatusBad.Render("EXPIRED")
		expiryHint := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Italic(true).Render("(can be purged)")
		rows = append(rows, fmt.Sprintf("  %s %s %s %s", expiryIcon, statusLabel, statusText, expiryHint))
	}

	rows = append(rows, "")

	// Restore command
	restoreIcon := m.styles.IconStyle.Render("🔄")
	restoreLabel := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Muted)).Render("Restore:")
	restoreCmd := m.styles.Filename.Render(fmt.Sprintf("vx --restore %s", m.pattern))
	rows = append(rows, fmt.Sprintf("  %s %s %s", restoreIcon, restoreLabel, restoreCmd))

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return content
}

func (m *infoModel) renderPagination() string {
	totalPages := (len(m.matchingItems) + m.itemsPerPage - 1) / m.itemsPerPage
	currentPage := m.currentPage + 1

	paginationText := fmt.Sprintf("Page %d of %d", currentPage, totalPages)
	paginationStyle := m.styles.Info.Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).Bold(true)

	navigation := m.styles.Help.Render("[←/p: Previous] [→/n: Next]")

	return lipgloss.JoinVertical(lipgloss.Center,
		"",
		paginationStyle.Render(paginationText),
		navigation,
	)
}

func (m *infoModel) renderHelp() string {
	helpText := m.styles.Help.Render("Press q or ESC to exit")
	return lipgloss.NewStyle().MarginTop(1).Render(helpText)
}

// ShowInfo searches for cached items matching the given pattern and displays
// detailed metadata for each item using a beautiful Bubble Tea TUI.
func ShowInfo(pattern string, config types.Config) error {
	styles := helpers.CreateThemeStyles(config)

	m := &infoModel{
		config:       config,
		pattern:      pattern,
		styles:       styles,
		itemsPerPage: 3, // Show 3 items per page
	}

	p := tea.NewProgram(m)
	_, err := p.Run()

	if err != nil {
		return fmt.Errorf("error running info display: %v", err)
	}

	return nil
}
