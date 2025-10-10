package command

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"vanish/internal/helpers"
	"vanish/internal/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	itemsPerPage = 10
)

type listModel struct {
	items       []types.DeletedItem
	config      types.Config
	cursor      int
	currentPage int
	totalPages  int
	styles      types.ThemeStyles
	err         error
}

type loadIndexMsg struct {
	items []types.DeletedItem
	err   error
}

func loadIndexCmd(config types.Config) tea.Cmd {
	return func() tea.Msg {
		index, err := helpers.LoadIndex(config)
		if err != nil {
			return loadIndexMsg{err: err}
		}

		// Sort by delete date (newest first)
		sort.Slice(index.Items, func(i, j int) bool {
			return index.Items[i].DeleteDate.After(index.Items[j].DeleteDate)
		})

		return loadIndexMsg{items: index.Items}
	}
}

func initialModel(config types.Config) listModel {
	return listModel{
		config:      config,
		currentPage: 0,
		styles:      helpers.CreateThemeStyles(config),
	}
}

func (m listModel) Init() tea.Cmd {
	return loadIndexCmd(m.config)
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Move to previous page if cursor goes above current page
				if m.cursor < m.currentPage*itemsPerPage {
					m.currentPage--
				}
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				// Move to next page if cursor goes below current page
				if m.cursor >= (m.currentPage+1)*itemsPerPage {
					m.currentPage++
				}
			}

		case "left", "h":
			if m.currentPage > 0 {
				m.currentPage--
				m.cursor = m.currentPage * itemsPerPage
			}

		case "right", "l":
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
				m.cursor = m.currentPage * itemsPerPage
			}

		case "home", "g":
			m.cursor = 0
			m.currentPage = 0

		case "end", "G":
			m.cursor = len(m.items) - 1
			m.currentPage = m.totalPages - 1

		case "pgup":
			if m.currentPage > 0 {
				m.currentPage--
				m.cursor = m.currentPage * itemsPerPage
			}

		case "pgdown":
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
				m.cursor = m.currentPage * itemsPerPage
			}
		}

	case tea.WindowSizeMsg:
		// No need to adjust viewport size anymore

	case loadIndexMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.items = msg.items
		// Calculate total pages
		if len(m.items) > 0 {
			m.totalPages = (len(m.items) + itemsPerPage - 1) / itemsPerPage
		} else {
			m.totalPages = 0
		}
	}

	return m, nil
}

func (m listModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading index: %v\n", m.err)
	}

	var b strings.Builder

	// Title
	title := fmt.Sprintf("Cached Files (%d items)", len(m.items))
	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n")

	// Empty state
	if len(m.items) == 0 {
		b.WriteString(m.styles.Info.Render("No cached files found."))
		b.WriteString("\n\n")
		b.WriteString(m.styles.Help.Render("Press q to quit"))
		return b.String()
	}

	// Page info
	pageInfo := fmt.Sprintf("Page %d of %d", m.currentPage+1, m.totalPages)
	b.WriteString(m.styles.Help.Render(pageInfo))
	b.WriteString("\n")

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.config.UI.Colors.Primary)).
		Background(lipgloss.Color(m.config.UI.Colors.Border)).
		Padding(0, 1)
	header := fmt.Sprintf("%-4s | %-16s | %-8s | %-8s | %-10s | %s",
		"Type", "Deleted", "Size", "Status", "Days Left", "Original Path")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Calculate visible range for current page
	visibleStart := m.currentPage * itemsPerPage
	visibleEnd := minInt(len(m.items), visibleStart+itemsPerPage)

	// Items
	for i := visibleStart; i < visibleEnd; i++ {
		item := m.items[i]
		line := m.formatItem(item, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Navigation info
	b.WriteString("\n")
	if m.totalPages > 1 {
		navInfo := fmt.Sprintf("← Page %d/%d →", m.currentPage+1, m.totalPages)
		b.WriteString(m.styles.Help.Render(navInfo))
		b.WriteString("\n")
	}

	// Help text
	help := "↑/k up • ↓/j down • ←/h prev page • →/l next page • g home • G end • q quit"
	b.WriteString(m.styles.Help.Render(help))

	return b.String()
}

func (m listModel) formatItem(item types.DeletedItem, isSelected bool) string {
	fileType := "FILE"
	if item.IsDirectory {
		fileType = "DIR"
	}

	expiryDate := item.DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
	daysLeft := int(time.Until(expiryDate).Hours() / 24)

	status := "OK"
	var statusColor lipgloss.Color
	if daysLeft <= 0 {
		status = "EXPIRED"
		statusColor = lipgloss.Color(m.config.UI.Colors.Error)
	} else if daysLeft <= 2 {
		status = "EXPIRING"
		statusColor = lipgloss.Color(m.config.UI.Colors.Warning)
	} else {
		status = "OK"
		statusColor = lipgloss.Color(m.config.UI.Colors.Success)
	}

	// Simple status style with just color, no borders or padding
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)

	// Format the line
	line := fmt.Sprintf("%-4s | %-16s | %-8s | %-8s | %-10s | %s",
		fileType,
		item.DeleteDate.Format("2006-01-02 15:04"),
		helpers.FormatBytes(item.Size),
		statusStyle.Render(status),
		fmt.Sprintf("%d days", daysLeft),
		item.OriginalPath,
	)

	// Apply selection style if this is the cursor position
	if isSelected {
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.config.UI.Colors.Text)).
			Background(lipgloss.Color(m.config.UI.Colors.Highlight)).
			Bold(true)
		return selectedStyle.Render(line)
	}

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.config.UI.Colors.Text))
	return normalStyle.Render(line)
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ShowList displays an interactive TUI list of cached files and directories
func ShowList(config types.Config) error {
	p := tea.NewProgram(initialModel(config))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %v", err)
	}
	return nil
}
