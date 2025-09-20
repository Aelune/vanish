package command

import (
	"fmt"
	"sort"
	"strings"

	"vanish/internal/config"
	"vanish/internal/helpers"
	"vanish/internal/types"
)

// MainThemeDisplayer implements the ThemeDisplayer interface,
// providing logic to retrieve and render theme data.
type MainThemeDisplayer struct{}

// ThemeDisplayer defines the interface for retrieving theme-related data,
// rendering previews, and accessing the configuration path.
type ThemeDisplayer interface {
	GetCurrentTheme() string
	GetAvailableThemes() []string
	RenderThemePreview(themeName string) string
	GetConfigPath() string
}

// ShowThemesWithTuiPreview displays a TUI-based theme preview,
// including the current theme and available alternatives.
func ShowThemesWithTuiPreview(displayer ThemeDisplayer) {
	currentTheme := displayer.GetCurrentTheme()
	availableThemes := displayer.GetAvailableThemes()

	width, _ := helpers.GetTerminalSize()
	adjustedWidth := width - 2
	if adjustedWidth < 20 {
		adjustedWidth = 20
	}
	title := "Vanish Theme Showcase - Interactive Previews"
	padding := (adjustedWidth - len(title)) / 2
	fmt.Println(strings.Repeat(" ", padding) + title)
	fmt.Println("=" + strings.Repeat("=", adjustedWidth))

	// Show current theme
	fmt.Printf("\nCURRENT THEME: %s\n", strings.ToUpper(currentTheme))
	fmt.Println(strings.Repeat("-", adjustedWidth/2))
	fmt.Print(displayer.RenderThemePreview(currentTheme))

	// Show other themes
	sort.Strings(availableThemes)
	for _, name := range availableThemes {
		if name != currentTheme {
			fmt.Printf("\n%s:\n", strings.ToUpper(name))
			fmt.Println(strings.Repeat("-", 40))
			fmt.Print(displayer.RenderThemePreview(name))
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("To use a theme, set 'theme = \"<name>\"' in your vanish.toml cfg file.")
	fmt.Println("Config location:", displayer.GetConfigPath())
}

// GetCurrentTheme returns the name of the currently configured theme,
// falling back to "default" if none is found.
func (m *MainThemeDisplayer) GetCurrentTheme() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "default"
	}

	currentTheme := cfg.UI.Theme
	if currentTheme == "" {
		currentTheme = "default"
	}

	defaultThemes := config.GetDefaultThemes()
	if _, exists := defaultThemes[currentTheme]; !exists {
		return currentTheme
	}

	return currentTheme
}

// GetAvailableThemes returns a list of names of all available built-in themes.
func (m *MainThemeDisplayer) GetAvailableThemes() []string {
	themes := config.GetDefaultThemes()
	var names []string
	for name := range themes {
		names = append(names, name)
	}
	return names
}

// RenderThemePreview returns a string representation of the theme preview
// for the given theme name.
func (m *MainThemeDisplayer) RenderThemePreview(themeName string) string {
	defaultThemes := config.GetDefaultThemes()

	var cfg types.Config
	if theme, exists := defaultThemes[themeName]; exists {
		cfg = theme
	} else {
		loadedConfig, err := config.LoadConfig()
		if err != nil {
			cfg = defaultThemes["default"]
		} else {
			cfg = loadedConfig
		}
	}

	return helpers.RenderThemeAsString(cfg)
}

// GetConfigPath returns the file path to the vanish.toml configuration file.
func (m *MainThemeDisplayer) GetConfigPath() string {
	return helpers.GetConfigPath()
}

// func colorsEqual(a, b struct {
// 	Primary     string `toml:"primary"`
// 	Secondary   string `toml:"secondary"`
// 	Success     string `toml:"success"`
// 	Warning     string `toml:"warning"`
// 	Error       string `toml:"error"`
// 	Text        string `toml:"text"`
// 	Muted       string `toml:"muted"`
// 	Border      string `toml:"border"`
// 	Highlight   string `toml:"highlight"`
// }) bool {
// 	return a.Primary == b.Primary &&
// 		a.Secondary == b.Secondary &&
// 		a.Success == b.Success &&
// 		a.Warning == b.Warning &&
// 		a.Error == b.Error &&
// 		a.Text == b.Text &&
// 		a.Muted == b.Muted &&
// 		a.Border == b.Border &&
// 		a.Highlight == b.Highlight
// }
