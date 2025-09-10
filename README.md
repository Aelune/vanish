# Vanish 🗑️

A modern, safe alternative to `rm` for Linux/Unix systems. Vanish moves your files to a recoverable cache instead of permanently deleting them, giving you peace of mind and the ability to recover accidentally deleted files.

## ✨ Features

- **Safe Deletion**: Files are moved to cache, not permanently deleted
- **Recoverable**: Files are kept for a configurable period (default: 10 days)
- **Modern TUI**: Beautiful terminal interface with themes and progress bars
- **No Confirmation Mode**: `--noconfirm` flag for batch operations
- **Smart Protection**: Automatic protection for system paths and important files
- **Configurable Themes**: Multiple built-in themes (default, dark, light, cyberpunk, minimal)
- **Detailed Logging**: Comprehensive operation logging with multiple levels
- **Cross-filesystem**: Works across different filesystems and mount points
- **Large File Handling**: Special handling for large files and directories
- **Batch Operations**: Process multiple files and directories at once

## 🚀 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/vanish.git
cd vanish

# Build and install
make build
make install
```

### Manual Installation

```bash
# Build the binary
go build -o vx main.go

# Move to your PATH
sudo mv vx /usr/local/bin/
```

## 📖 Usage

### Basic Usage

```bash
# Delete a single file
vx file.txt

# Delete multiple files and directories
vx file1.txt file2.txt some_directory/

# Delete without confirmation (for scripts/automation)
vx --noconfirm *.log temp_files/

# Clear all cached files immediately
vx --clear

# Show available themes
vx --themes

# Show help
vx --help
```

### Examples

```bash
# Safe deletion with confirmation
vx important_document.pdf

# Batch deletion for cleanup scripts
vx --noconfirm /tmp/* ~/.cache/thumbnails/

# Delete with pattern matching
vx *.tmp *.bak build/

# Clear the entire cache
vx --clear
```

## ⚙️ Configuration

Vanish uses a TOML configuration file located at `~/.config/vanish/vanish.toml`. The file is automatically created with defaults on first run.

### Configuration Options

```toml
[cache]
# Directory where deleted files are stored
directory = ".cache/vanish"
# Number of days to keep deleted files
days = 10

[logging]
# Enable logging
enabled = true
# Directory for log files
directory = ".cache/vanish/logs"
# Log level: "info", "debug", "error"
level = "info"

[ui]
# Theme: "default", "dark", "light", "cyberpunk", "minimal"
theme = "default"
# Padding around content
padding_x = 2
padding_y = 1
# Show detailed file information
show_details = true
# Use compact display mode
compact = false

[ui.colors]
# Customize colors (hex values)
primary = "#3B82F6"
secondary = "#6366F1"
success = "#10B981"
warning = "#F59E0B"
error = "#EF4444"
text = "#F9FAFB"
muted = "#9CA3AF"
border = "#374151"
highlight = "#FBBF24"

[ui.progress]
# Progress bar style: "gradient", "solid", "rainbow"
style = "gradient"
# Show emoji in messages
show_emoji = true
# Enable animations
animation = true
# Show/hide progress bar completely
enabled = true

[behavior]
# Skip confirmation prompts (same as --noconfirm)
auto_confirm = false
# Show detailed output during operations
verbose_output = false
# Show file count for directories
show_file_count = true
# Always confirm for large files/directories
confirm_on_large = true
# Size limit for "large" files (bytes)
large_size_limit = 104857600  # 100MB
# File count limit for "large" directories
large_count_limit = 1000

[safety]
# Paths that cannot be deleted (always protected)
protected_paths = [
    "/", "/home", "/usr", "/etc", "/var", "/boot", "/sys", "/proc"
]
# File patterns that always require confirmation
require_confirm = [
    "*.env", "*.key", "*.pem", "config.toml", "*.config"
]
# Create additional backup for important files
backup_important = false
```

## 🎨 Themes

Vanish comes with several built-in themes:

- **default**: Blue-focused theme with gradients
- **dark**: Purple and dark theme
- **light**: Clean light theme for light terminals
- **cyberpunk**: Neon colors with high contrast
- **minimal**: Simple black and white theme

View all available themes:
```bash
vx --themes
```

## 🔧 Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for build automation)

### Building

```bash
# Setup development environment
make dev-setup

# Build
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint
```

### Project Structure

```
vanish/
├── main.go                 # Entry point
├── internal/
│   ├── app/               # Application logic
│   ├── config/            # Configuration management
│   ├── filesystem/        # File operations
│   ├── logging/           # Logging system
│   ├── models/            # Data models
│   ├── tui/               # Terminal UI
│   └── ui/                # UI components and themes
├── go.mod                 # Go module definition
├── Makefile              # Build automation
└── README.md             # This file
```

### Building for Multiple Platforms

```bash
# Build for all platforms
make build-all

# Create release packages
make package
```

## 🛡️ Safety Features

### Protected Paths

Vanish automatically protects critical system directories:
- `/`, `/home`, `/usr`, `/etc`, `/var`, `/boot`, `/sys`, `/proc`

### Smart Confirmation

Vanish will always ask for confirmation when:
- Deleting protected paths
- Files match confirmation patterns (`.env`, `.key`, etc.)
- Large files (>100MB) or directories (>1000 files)
- When `confirm_on_large` is enabled

### File Recovery

Files are stored in `~/.cache/vanish/` with timestamps and can be manually recovered if needed. The cache is automatically cleaned after the configured retention period.

## 📊 Logging

Vanish provides comprehensive logging:

- **Text logs**: Human-readable logs in `vanish.log`
- **JSON logs**: Machine-readable logs for analysis (debug level)
- **Operation tracking**: All delete, cleanup, and recovery operations
- **Error logging**: Detailed error information with context

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests for new functionality
5. Run tests: `make test`
6. Commit your changes: `git commit -am 'Add some feature'`
7. Push to the branch: `git push origin feature-name`
8. Submit a pull request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh/) for the excellent TUI libraries
- [BurntSushi/toml](https://github.com/BurntSushi/toml) for TOML configuration support
- The Go community for great tooling and libraries

## 📝 Changelog

### v1.0.0
- Initial release with safe deletion
- Multi-module architecture
- Configurable themes and behavior
- Comprehensive logging
- Protection for system paths
- `--noconfirm` flag support

---

**Note**: Vanish is designed to be a safer alternative to `rm`, but it's not a backup solution. Always maintain proper backups of important data.
