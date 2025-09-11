package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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
		Level     string `toml:"level"` // "info", "debug", "error"
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
			Enabled    bool   `toml:"enabled"` // Show/hide progress bar
		} `toml:"progress"`
		PaddingX    int  `toml:"padding_x"`
		PaddingY    int  `toml:"padding_y"`
		ShowDetails bool `toml:"show_details"` // Show file details in confirmation
		Compact     bool `toml:"compact"`      // Compact display mode
	} `toml:"ui"`

	Behavior struct {
		AutoConfirm    bool `toml:"auto_confirm"`    // Skip confirmation prompts
		VerboseOutput  bool `toml:"verbose_output"`  // Show detailed output
		ShowFileCount  bool `toml:"show_file_count"` // Show file count for directories
		ConfirmOnLarge bool `toml:"confirm_on_large"` // Always confirm for large files/dirs
		LargeSizeLimit int64 `toml:"large_size_limit"` // Size limit in bytes for "large" files
		LargeCountLimit int  `toml:"large_count_limit"` // File count limit for "large" directories
	} `toml:"behavior"`

	Safety struct {
		ProtectedPaths []string `toml:"protected_paths"` // Paths that cannot be deleted
		RequireConfirm []string `toml:"require_confirm"` // Patterns that always require confirmation
		BackupImportant bool    `toml:"backup_important"` // Create additional backup for important files
	} `toml:"safety"`
}

func GetDefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()

	config := Config{}
	config.Cache.Directory = filepath.Join(homeDir, ".cache", "vanish")
	config.Cache.Days = 10

	config.Logging.Enabled = true
	config.Logging.Directory = filepath.Join(homeDir, ".cache", "vanish", "logs")
	config.Logging.Level = "info"

	// Apply default theme
	defaultTheme := GetDefaultThemes()["default"]
	config.UI = defaultTheme.UI

	config.Behavior.AutoConfirm = false
	config.Behavior.VerboseOutput = false
	config.Behavior.ShowFileCount = true
	config.Behavior.ConfirmOnLarge = true
	config.Behavior.LargeSizeLimit = 100 * 1024 * 1024 // 100MB
	config.Behavior.LargeCountLimit = 1000             // 1000 files

	config.Safety.ProtectedPaths = []string{
		"/", "/home", "/usr", "/etc", "/var", "/boot", "/sys", "/proc",
	}
	config.Safety.RequireConfirm = []string{
		"*.env", "*.key", "*.pem", "config.toml", "*.config",
	}
	config.Safety.BackupImportant = false

	return config
}

func GetDefaultThemes() map[string]Config {
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
	defaultTheme.UI.Progress.Enabled = true
	defaultTheme.UI.PaddingX = 2
	defaultTheme.UI.PaddingY = 1
	defaultTheme.UI.ShowDetails = true
	defaultTheme.UI.Compact = false
	themes["default"] = defaultTheme

	// Dark theme
	darkTheme := defaultTheme
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
	themes["dark"] = darkTheme

	// Light theme
	lightTheme := defaultTheme
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
	lightTheme.UI.Progress.Animation = false
	themes["light"] = lightTheme

	// Cyberpunk theme
	cyberpunkTheme := defaultTheme
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
	themes["cyberpunk"] = cyberpunkTheme

	// Minimal theme
	minimalTheme := defaultTheme
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
	minimalTheme.UI.Progress.Enabled = false
	minimalTheme.UI.Compact = true
	themes["minimal"] = minimalTheme

	return themes
}

func LoadConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(homeDir, ".config", "vanish", "vanish.toml")
	config := GetDefaultConfig()

	// Try to load config file
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return config, fmt.Errorf("error parsing config file: %v", err)
		}

		// If a theme is specified, apply it but preserve any custom color overrides
		if config.UI.Theme != "" && config.UI.Theme != "default" {
			themes := GetDefaultThemes()
			if themeConfig, exists := themes[config.UI.Theme]; exists {
				// Save current custom colors
				customColors := config.UI.Colors
				customProgress := config.UI.Progress
				customUI := config.UI

				// Apply theme
				config.UI = themeConfig.UI

				// Restore any custom settings that were explicitly set
				if customColors.Primary != "" {
					config.UI.Colors.Primary = customColors.Primary
				}
				// Preserve other custom settings
				if customProgress.Enabled != themeConfig.UI.Progress.Enabled {
					config.UI.Progress.Enabled = customProgress.Enabled
				}
				if customUI.ShowDetails != themeConfig.UI.ShowDetails {
					config.UI.ShowDetails = customUI.ShowDetails
				}
				if customUI.Compact != themeConfig.UI.Compact {
					config.UI.Compact = customUI.Compact
				}
			}
		}
	} else {
		// Create default config file
		if err := CreateDefaultConfig(configPath, config); err != nil {
			// Don't fail if we can't create config, just warn
			fmt.Printf("Warning: Could not create default config: %v\n", err)
		}
	}

	return config, nil
}

func CreateDefaultConfig(configPath string, config Config) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configContent := `# Vanish Configuration File

[cache]
# Directory where deleted files are stored (relative to home directory)
directory = ".cache/vanish"
# Number of days to keep deleted files
days = 10

[logging]
# Enable logging
enabled = true
# Directory for log files (relative to cache directory)
directory = ".cache/vanish/logs"
# Log level: "info", "debug", "error"
level = "info"

[ui]
# Theme: "default", "dark", "light", "cyberpunk", "minimal"
theme = "default"
# Padding around content
padding_x = 2
padding_y = 1
# Show detailed file information in confirmation
show_details = true
# Use compact display mode (less spacing)
compact = false

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
enabled = true           # Show/hide progress bar completely

[behavior]
# Behavioral settings
auto_confirm = false     # Skip confirmation prompts (same as --noconfirm)
verbose_output = false   # Show detailed output during operations
show_file_count = true   # Show file count for directories
confirm_on_large = true  # Always confirm for large files/directories
large_size_limit = 104857600  # 100MB - size limit for "large" files
large_count_limit = 1000      # file count limit for "large" directories

[safety]
# Safety settings
protected_paths = [      # Paths that cannot be deleted
    "/", "/home", "/usr", "/etc", "/var", "/boot", "/sys", "/proc"
]
require_confirm = [      # File patterns that always require confirmation
    "*.env", "*.key", "*.pem", "config.toml", "*.config"
]
backup_important = false # Create additional backup for important files
`

	return os.WriteFile(configPath, []byte(configContent), 0644)
}
