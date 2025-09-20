// Package config manages all config related code
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"vanish/internal/types"
)

// GetDefaultThemes returns a map of predefined themes used by the Vanish TUI.
func GetDefaultThemes() map[string]types.Config {
	themes := make(map[string]types.Config)

	// Default theme - Modern blue with good contrast
	defaultTheme := types.Config{}
	defaultTheme.UI.Theme = "default"
	defaultTheme.UI.Colors.Primary = "#2563EB"   // Blue 600 - professional blue
	defaultTheme.UI.Colors.Secondary = "#3B82F6" // Blue 500 - lighter blue
	defaultTheme.UI.Colors.Success = "#10B981"   // Emerald 500 - success green
	defaultTheme.UI.Colors.Warning = "#F59E0B"   // Amber 500 - warning yellow
	defaultTheme.UI.Colors.Error = "#EF4444"     // Red 500 - error red
	defaultTheme.UI.Colors.Text = "#F8FAFC"      // Slate 50 - clean white text
	defaultTheme.UI.Colors.Muted = "#94A3B8"     // Slate 400 - muted text
	defaultTheme.UI.Colors.Border = "#475569"    // Slate 600 - subtle borders
	defaultTheme.UI.Colors.Highlight = "#60A5FA" // Blue 400 - highlight
	defaultTheme.UI.Progress.Style = "gradient"
	defaultTheme.UI.Progress.ShowEmoji = true
	defaultTheme.UI.Progress.Animation = true
	themes["default"] = defaultTheme

	// Dark theme - Deep purples and rich colors
	darkTheme := types.Config{}
	darkTheme.UI.Theme = "dark"
	darkTheme.UI.Colors.Primary = "#7C3AED"   // Violet 600 - deep purple
	darkTheme.UI.Colors.Secondary = "#A855F7" // Purple 500 - lighter purple
	darkTheme.UI.Colors.Success = "#22C55E"   // Green 500 - vibrant green
	darkTheme.UI.Colors.Warning = "#EAB308"   // Yellow 500 - golden yellow
	darkTheme.UI.Colors.Error = "#F97316"     // Orange 500 - warm red-orange
	darkTheme.UI.Colors.Text = "#E2E8F0"      // Slate 200 - soft white
	darkTheme.UI.Colors.Muted = "#64748B"     // Slate 500 - muted gray
	darkTheme.UI.Colors.Border = "#374151"    // Gray 700 - dark borders
	darkTheme.UI.Colors.Highlight = "#C084FC" // Purple 400 - bright highlight
	darkTheme.UI.Progress.Style = "gradient"
	darkTheme.UI.Progress.ShowEmoji = true
	darkTheme.UI.Progress.Animation = true
	themes["dark"] = darkTheme

	// Light theme - Clean and professional
	lightTheme := types.Config{}
	lightTheme.UI.Theme = "light"
	lightTheme.UI.Colors.Primary = "#1D4ED8"   // Blue 700 - strong blue
	lightTheme.UI.Colors.Secondary = "#2563EB" // Blue 600 - medium blue
	lightTheme.UI.Colors.Success = "#15803D"   // Green 700 - forest green
	lightTheme.UI.Colors.Warning = "#CA8A04"   // Yellow 600 - amber
	lightTheme.UI.Colors.Error = "#DC2626"     // Red 600 - strong red
	lightTheme.UI.Colors.Text = "#0F172A"      // Slate 900 - dark text
	lightTheme.UI.Colors.Muted = "#64748B"     // Slate 500 - medium gray
	lightTheme.UI.Colors.Border = "#CBD5E1"    // Slate 300 - light borders
	lightTheme.UI.Colors.Highlight = "#0EA5E9" // Sky 500 - bright blue
	lightTheme.UI.Progress.Style = "solid"
	lightTheme.UI.Progress.ShowEmoji = true
	lightTheme.UI.Progress.Animation = false
	themes["light"] = lightTheme

	// Cyberpunk theme - Neon colors with retro-futuristic feel
	cyberpunkTheme := types.Config{}
	cyberpunkTheme.UI.Theme = "cyberpunk"
	cyberpunkTheme.UI.Colors.Primary = "#00F5FF"   // Electric cyan - neon blue
	cyberpunkTheme.UI.Colors.Secondary = "#FF1493" // Deep pink - hot pink
	cyberpunkTheme.UI.Colors.Success = "#39FF14"   // Electric lime - neon green
	cyberpunkTheme.UI.Colors.Warning = "#FFD700"   // Gold - electric yellow
	cyberpunkTheme.UI.Colors.Error = "#FF073A"     // Neon red - bright red
	cyberpunkTheme.UI.Colors.Text = "#00FFFF"      // Cyan - matrix green text
	cyberpunkTheme.UI.Colors.Muted = "#9D4EDD"     // Electric purple - muted neon
	cyberpunkTheme.UI.Colors.Border = "#FF00FF"    // Magenta - neon border
	cyberpunkTheme.UI.Colors.Highlight = "#FFFF00" // Electric yellow - highlight
	cyberpunkTheme.UI.Progress.Style = "rainbow"
	cyberpunkTheme.UI.Progress.ShowEmoji = false // Keep it tech-focused
	cyberpunkTheme.UI.Progress.Animation = true
	themes["cyberpunk"] = cyberpunkTheme

	// Minimal theme - Monochromatic with subtle accents
	minimalTheme := types.Config{}
	minimalTheme.UI.Theme = "minimal"
	minimalTheme.UI.Colors.Primary = "#374151"   // Gray 700 - dark gray
	minimalTheme.UI.Colors.Secondary = "#6B7280" // Gray 500 - medium gray
	minimalTheme.UI.Colors.Success = "#16A34A"   // Green 600 - clean green
	minimalTheme.UI.Colors.Warning = "#D97706"   // Amber 600 - muted orange
	minimalTheme.UI.Colors.Error = "#DC2626"     // Red 600 - clear red
	minimalTheme.UI.Colors.Text = "#111827"      // Gray 900 - almost black
	minimalTheme.UI.Colors.Muted = "#9CA3AF"     // Gray 400 - light gray
	minimalTheme.UI.Colors.Border = "#D1D5DB"    // Gray 300 - subtle border
	minimalTheme.UI.Colors.Highlight = "#4B5563" // Gray 600 - subtle highlight
	minimalTheme.UI.Progress.Style = "solid"
	minimalTheme.UI.Progress.ShowEmoji = false // Clean, no emojis
	minimalTheme.UI.Progress.Animation = false // Static for minimalism
	themes["minimal"] = minimalTheme

	// Ocean theme - Blue and teal palette inspired by the sea
	oceanTheme := types.Config{}
	oceanTheme.UI.Theme = "ocean"
	oceanTheme.UI.Colors.Primary = "#0891B2"   // Cyan 600 - ocean blue
	oceanTheme.UI.Colors.Secondary = "#06B6D4" // Cyan 500 - lighter ocean
	oceanTheme.UI.Colors.Success = "#059669"   // Emerald 600 - sea green
	oceanTheme.UI.Colors.Warning = "#0284C7"   // Sky 600 - deep blue warning
	oceanTheme.UI.Colors.Error = "#0F766E"     // Teal 700 - dark teal for errors
	oceanTheme.UI.Colors.Text = "#F0FDFF"      // Cyan 50 - sea foam white
	oceanTheme.UI.Colors.Muted = "#67E8F9"     // Cyan 300 - light sea blue
	oceanTheme.UI.Colors.Border = "#155E75"    // Cyan 800 - deep ocean border
	oceanTheme.UI.Colors.Highlight = "#22D3EE" // Cyan 400 - bright ocean highlight
	oceanTheme.UI.Progress.Style = "gradient"
	oceanTheme.UI.Progress.ShowEmoji = true
	oceanTheme.UI.Progress.Animation = true
	themes["ocean"] = oceanTheme

	// Forest theme - Green palette inspired by nature
	forestTheme := types.Config{}
	forestTheme.UI.Theme = "forest"
	forestTheme.UI.Colors.Primary = "#15803D"   // Green 700 - deep forest
	forestTheme.UI.Colors.Secondary = "#16A34A" // Green 600 - medium forest
	forestTheme.UI.Colors.Success = "#22C55E"   // Green 500 - bright success
	forestTheme.UI.Colors.Warning = "#EAB308"   // Yellow 500 - autumn warning
	forestTheme.UI.Colors.Error = "#DC2626"     // Red 600 - danger red
	forestTheme.UI.Colors.Text = "#F0FDF4"      // Green 50 - light forest text
	forestTheme.UI.Colors.Muted = "#86EFAC"     // Green 300 - soft forest muted
	forestTheme.UI.Colors.Border = "#14532D"    // Green 900 - dark forest border
	forestTheme.UI.Colors.Highlight = "#4ADE80" // Green 400 - fresh forest highlight
	forestTheme.UI.Progress.Style = "gradient"
	forestTheme.UI.Progress.ShowEmoji = true
	forestTheme.UI.Progress.Animation = true
	themes["forest"] = forestTheme

	// Sunset theme - Warm colors inspired by sunset
	sunsetTheme := types.Config{}
	sunsetTheme.UI.Theme = "sunset"
	sunsetTheme.UI.Colors.Primary = "#EA580C"   // Orange 600 - sunset orange
	sunsetTheme.UI.Colors.Secondary = "#F97316" // Orange 500 - lighter sunset
	sunsetTheme.UI.Colors.Success = "#84CC16"   // Lime 500 - sunset green
	sunsetTheme.UI.Colors.Warning = "#EAB308"   // Yellow 500 - sunset yellow
	sunsetTheme.UI.Colors.Error = "#DC2626"     // Red 600 - sunset red
	sunsetTheme.UI.Colors.Text = "#FFF7ED"      // Orange 50 - warm white
	sunsetTheme.UI.Colors.Muted = "#FDBA74"     // Orange 300 - warm muted
	sunsetTheme.UI.Colors.Border = "#9A3412"    // Orange 800 - deep sunset border
	sunsetTheme.UI.Colors.Highlight = "#FBBF24" // Amber 400 - golden highlight
	sunsetTheme.UI.Progress.Style = "gradient"
	sunsetTheme.UI.Progress.ShowEmoji = true
	sunsetTheme.UI.Progress.Animation = true
	themes["sunset"] = sunsetTheme

	return themes
}

func createDefaultConfig(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configContent :=
		`[cache]
# ============================================================================
# ⚠️ WARNING: Do not modify the cache directory if it already have files stored!
#
# The directory setting below determines where Vanish stores deleted files.
# If you point this to a folder that already contains important data, Vanish
# will treat it as its cache and operations (restore, purge, clear) may not
# work as expected, potentially leading to data loss.
# ============================================================================

# Directory where deleted files are stored (relative to HOME directory)
directory = ".cache/vanish"

# Number of days to keep deleted files before automatic cleanup
days = 10

# ------------------------------
# Logging Configuration
# ------------------------------
[logging]
# Enable or disable logging (true/false)
enabled = true

# Directory for log files (relative to the cache directory above)
directory = ".cache/vanish/logs"

# ------------------------------
# User Interface (UI) Settings
# ------------------------------
[ui]
# Theme options: "default", "dark", "light", "cyberpunk", "minimal", "ocean", "forest", "sunset"
theme = "default"

# Skip confirmation prompts (use with caution!)
no_confirm = false

# ------------------------------
# UI Color Customization
# Uncomment and customize hex values if you want a custom look.
# ------------------------------
[ui.colors]
# primary   = "#3B82F6"  # Main accent color
# secondary = "#6366F1"  # Secondary accent
# success   = "#10B981"  # Success messages
# warning   = "#F59E0B"  # Warning messages
# error     = "#EF4444"  # Error messages
# text      = "#F9FAFB"  # Main text color
# muted     = "#9CA3AF"  # Muted/help text
# border    = "#374151"  # Border color
# highlight = "#FBBF24"  # Highlighted filename

# ------------------------------
# Progress Bar Settings
# ------------------------------
[ui.progress]
# style       = "gradient"   # Options: "gradient", "solid", "rainbow"
# show_emoji  = true         # Adds emoji to progress messages
# animation   = true         # Smooth animation (disable for performance)
`

	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// LoadConfig loads the user's configuration from ~/.config/vanish/vanish.toml.
// If the file does not exist, it creates a default config.
// It also applies any matching theme and preserves custom overrides.
func LoadConfig() (types.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return types.Config{}, err
	}

	configPath := filepath.Join(homeDir, ".config", "vanish", "vanish.toml")

	// Default configuration
	config := types.Config{}
	config.Cache.Directory = filepath.Join(homeDir, ".cache", "vanish")
	config.Cache.Days = 10
	config.Logging.Enabled = true
	config.Logging.Directory = filepath.Join(homeDir, ".cache", "vanish", "logs")

	themes := GetDefaultThemes()

	// Try to load config file
	if _, err := os.Stat(configPath); err == nil {
		// Load the entire config from file first
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return config, fmt.Errorf("error parsing config file: %v", err)
		}

		// fmt.Printf("DEBUG: Loaded theme from config: '%s'\n", config.UI.Theme)

		// Determine which theme to use
		themeName := config.UI.Theme
		if themeName == "" {
			themeName = "default"
		}

		// Check if theme exists
		if themeConfig, exists := themes[themeName]; exists {
			// Store any custom colors that might have been set in config
			customColors := config.UI.Colors
			customProgress := config.UI.Progress

			// Apply the base theme
			config.UI = themeConfig.UI

			// Restore any custom colors that were explicitly set and differ from empty
			// We need a better way to detect custom vs default colors
			// For now, we'll use a simple heuristic: if the config file had the [ui.colors] section

			// Check if any custom colors were actually set by seeing if they differ from the theme default
			themeColors := themeConfig.UI.Colors

			if customColors.Primary != "" && customColors.Primary != themeColors.Primary {
				config.UI.Colors.Primary = customColors.Primary
			}
			if customColors.Secondary != "" && customColors.Secondary != themeColors.Secondary {
				config.UI.Colors.Secondary = customColors.Secondary
			}
			if customColors.Success != "" && customColors.Success != themeColors.Success {
				config.UI.Colors.Success = customColors.Success
			}
			if customColors.Warning != "" && customColors.Warning != themeColors.Warning {
				config.UI.Colors.Warning = customColors.Warning
			}
			if customColors.Error != "" && customColors.Error != themeColors.Error {
				config.UI.Colors.Error = customColors.Error
			}
			if customColors.Text != "" && customColors.Text != themeColors.Text {
				config.UI.Colors.Text = customColors.Text
			}
			if customColors.Muted != "" && customColors.Muted != themeColors.Muted {
				config.UI.Colors.Muted = customColors.Muted
			}
			if customColors.Border != "" && customColors.Border != themeColors.Border {
				config.UI.Colors.Border = customColors.Border
			}
			if customColors.Highlight != "" && customColors.Highlight != themeColors.Highlight {
				config.UI.Colors.Highlight = customColors.Highlight
			}

			// Restore custom progress settings
			if customProgress.Style != "" && customProgress.Style != themeConfig.UI.Progress.Style {
				config.UI.Progress.Style = customProgress.Style
			}

			// For booleans, we need to check if they were explicitly set in the config
			// This is tricky with TOML, but we can make reasonable assumptions
			config.UI.Progress.ShowEmoji = customProgress.ShowEmoji
			config.UI.Progress.Animation = customProgress.Animation

			// fmt.Printf("DEBUG: Applied theme '%s' with colors: Primary=%s, Success=%s\n",
			// themeName, config.UI.Colors.Primary, config.UI.Colors.Success)

		} else {
			// Unknown theme, fall back to default
			fmt.Printf("WARNING: Unknown theme '%s', falling back to default\n", themeName)
			defaultTheme := themes["default"]
			config.UI = defaultTheme.UI
		}

	} else {
		// No config file exists, use default theme
		defaultTheme := themes["default"]
		config.UI = defaultTheme.UI

		// Create default config file
		if err := createDefaultConfig(configPath); err != nil {
			log.Printf("Warning: Could not create default config: %v", err)
		}
	}

	return config, nil
}
