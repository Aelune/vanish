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
	processedItems []DeletedItem
	clearAll       bool
	totalFiles     int
	processedFiles int
}

var (
	questionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")) // bright blue

	deleterName = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Foreground(lipgloss.Color("#FBBF24")) // amber/yellow highlight for filename

	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CBD5E1")). // light slate gray
		Background(lipgloss.Color("#1E293B")). // dark slate blue background
		Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F87171")). // pastel red
		Bold(true).
		Background(lipgloss.Color("#7F1D1D")). // dark red background
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FCA5A5")) // soft red border

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#34D399")). // mint green
		Bold(true).
		Background(lipgloss.Color("#064E3B")). // dark green background
		Padding(0, 1).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#6EE7B7")) // light mint border

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")). // muted gray-blue
		Italic(true).
		Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")). // amber
		Bold(true)
)

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

	// Try to load config file
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return config, fmt.Errorf("error parsing config file: %v", err)
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
`

	return os.WriteFile(configPath, []byte(configContent), 0644)
}

func initialModel(filenames []string, clearAll bool) (Model, error) {
	config, err := loadConfig()
	if err != nil {
		return Model{}, err
	}

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50

	return Model{
		filenames:      filenames,
		fileInfos:      make([]FileInfo, len(filenames)),
		state:          "checking",
		progress:       prog,
		config:         config,
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
	var s strings.Builder

	switch m.state {
	case "checking":
		s.WriteString("🔍 Checking files and directories...\n\n")
		s.WriteString(m.progress.View())

	case "confirming":
		s.WriteString(questionStyle.Render("Are you sure you want to delete the following items?"))
		s.WriteString("\n\n")

		validCount := 0
		invalidCount := 0
		totalFileCount := 0

		for _, info := range m.fileInfos {
			if info.Exists {
				validCount++
				if info.IsDirectory {
					s.WriteString(fmt.Sprintf("📁 %s", deleterName.Render(info.Path)))
					if info.FileCount > 0 {
						s.WriteString(fmt.Sprintf(" (%d items)", info.FileCount))
						totalFileCount += info.FileCount
					} else {
						s.WriteString(" (empty)")
					}
				} else {
					s.WriteString(fmt.Sprintf("📄 %s", deleterName.Render(info.Path)))
					totalFileCount++
				}
				s.WriteString("\n")
			} else {
				invalidCount++
				s.WriteString(fmt.Sprintf("❌ %s", warningStyle.Render(info.Path)))
				s.WriteString(" (does not exist)\n")
			}
		}

		if invalidCount > 0 {
			s.WriteString(fmt.Sprintf("\n%s %d file(s) will be skipped (do not exist)\n",
				warningStyle.Render("Warning:"), invalidCount))
		}

		if validCount > 0 {
			s.WriteString(fmt.Sprintf("\nTotal items to delete: %d\n", validCount))
			if totalFileCount > validCount {
				s.WriteString(fmt.Sprintf("Total files/subdirectories affected: %d\n", totalFileCount))
			}
			s.WriteString(fmt.Sprintf("Content will be moved to cache and can be recovered for %d days.\n\n", m.config.Cache.Days))
			s.WriteString(helpStyle.Render("Press 'y' to confirm, 'n' to cancel, or 'q' to quit"))
		} else {
			s.WriteString("\n")
			s.WriteString(helpStyle.Render("Press 'n' to cancel or 'q' to quit"))
		}

	case "moving":
		if m.currentIndex < len(m.fileInfos) {
			currentFile := m.fileInfos[m.currentIndex]
			if currentFile.IsDirectory {
				s.WriteString(normalStyle.Render(fmt.Sprintf("📦 Moving directory '%s' to safe cache... (%d/%d)",
					currentFile.Path, m.processedFiles+1, m.totalFiles)))
			} else {
				s.WriteString(normalStyle.Render(fmt.Sprintf("📦 Moving file '%s' to safe cache... (%d/%d)",
					currentFile.Path, m.processedFiles+1, m.totalFiles)))
			}
		} else {
			s.WriteString(normalStyle.Render("📦 Moving files to safe cache..."))
		}
		s.WriteString("\n\n")
		s.WriteString(m.progress.View())

	case "cleanup":
		s.WriteString("🧹 Cleaning up old cached files...\n\n")
		s.WriteString(m.progress.View())

	case "clearing":
		s.WriteString("🗑️  Clearing all cached files...\n\n")
		s.WriteString(m.progress.View())

	case "done":
		if m.clearAll {
			s.WriteString(successStyle.Render("✅ All cached files cleared!"))
		} else {
			s.WriteString(successStyle.Render(fmt.Sprintf("✅ Successfully processed %d item(s)!", len(m.processedItems))))
		}
		s.WriteString("\n\n")

		if !m.clearAll && len(m.processedItems) > 0 {
			for i, item := range m.processedItems {
				if i < 5 { // Show details for first 5 items to avoid cluttering
					s.WriteString(fmt.Sprintf("• %s -> %s\n", item.OriginalPath, filepath.Base(item.CachePath)))
				}
			}

			if len(m.processedItems) > 5 {
				s.WriteString(fmt.Sprintf("... and %d more item(s)\n", len(m.processedItems)-5))
			}

			if len(m.processedItems) > 0 {
				deleteAfter := m.processedItems[0].DeleteDate.Add(time.Duration(m.config.Cache.Days) * 24 * time.Hour)
				s.WriteString(fmt.Sprintf("\nWill be permanently deleted after: %s\n\n", deleteAfter.Format("2006-01-02 15:04:05")))
			}
		}

		s.WriteString(m.progress.View())
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter or 'q' to exit"))

	case "error":
		s.WriteString(errorStyle.Render("❌ Error"))
		s.WriteString("\n\n")
		s.WriteString(m.errorMsg)
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter or 'q' to exit"))
	}

	return s.String()
}

// Messages
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
