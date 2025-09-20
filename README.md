# Vanish (vx) 🗑️

[![Release](https://img.shields.io/github/v/release/Aelune/vanish?style=flat-square)](https://github.com/Aelune/vanish/releases)
[![License](https://img.shields.io/github/license/Aelune/vanish?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aelune/vanish?style=flat-square)](https://goreportcard.com/report/github.com/Aelune/vanish)

> A modern, safe file deletion tool with recovery capabilities and beautiful TUI interface.

Vanish provides a secure alternative to permanent file deletion by moving files to a managed cache with full recovery options. Never lose important files again with intelligent backup, pattern-based restoration, and comprehensive file management.

## ✨ Features

- 🛡️ **Safe Deletion**: Files are moved to cache, not permanently deleted
- 🔄 **Pattern-based Recovery**: Restore files using flexible pattern matching
- 📊 **Rich Statistics**: Detailed insights into cache usage and file metrics
- 🎨 **Beautiful TUI**: Modern terminal interface with 8 built-in themes
- ⚡ **Fast Operations**: Optimized for handling large directories and multiple files
- 🔧 **Highly Configurable**: Extensive customization options via TOML config
<!-- - 🔔 **Smart Notifications**: Desktop notifications for operations (Linux/macOS/Windows) -->
- 📝 **Comprehensive Logging**: Track all operations with detailed audit trails
- 🧹 **Automated Cleanup**: Configurable retention policies and purging
<!-- - 🐚 **Shell Completion**: Full completion support for Bash, Zsh, Fish, PowerShell -->

## 🚀 Installation

### Download Binary
```bash
# Linux/macOS
curl -L https://github.com/Aelune/vanish/releases/latest/download/vx-linux-amd64 -o vx
chmod +x vx
sudo mv vx /usr/local/bin/

# Or use install script
curl -fsSL https://raw.githubusercontent.com/Aelune/vanish/main/install.sh | bash
```

### Build from Source
```bash
git clone https://github.com/Aelune/vanish.git
cd vanish
go build -o vx .
sudo mv vx /usr/local/bin/
```

<!-- ### Package Managers
```bash
# Homebrew (macOS/Linux)
brew install Aelune/tap/vanish

# Arch Linux (AUR)
yay -S vanish-bin

# Snap
sudo snap install vanish
``` -->

## 📖 Quick Start

### Basic Usage
```bash
# Delete files/directories safely
vx file.txt folder/ *.log

# List cached files
vx --list

# Restore specific files
vx --restore .txt project-n

# View detailed information
vx --info "important-file"

# Clear all cached files
vx --clear
```

### Advanced Operations
```bash
# Purge files older than 30 days
vx --purge 30

# Check cache statistics
vx --stats

# Restore with no confirmation
vx --restore --noconfirm "*.backup"

# View available themes
vx --themes
```

## 🎨 Themes & Customization

Vanish includes 8 beautiful built-in themes:

| Theme | Description |
|-------|-------------|
| `default` | Clean, professional look |
| `dark` | High contrast dark mode |
| `light` | Bright, minimal interface |
| `cyberpunk` | Neon colors with retro-futuristic feel |
| `minimal` | Ultra-clean, distraction-free |
| `ocean` | Calming blues and teals |
| `forest` | Natural greens and earth tones |
| `sunset` | Warm oranges and purples |

```bash
# Preview all themes interactively
vx --themes

# Set theme via config or environment
export VX_THEME=cyberpunk
```

## ⚙️ Configuration

Vanish uses a TOML based configuration for easy understanding look at [Config Documentation](https://dwukn.vercel.app/)

<!-- ## 🔧 Shell Completion

Enable tab completion for enhanced productivity:

```bash
# Bash
vx --completion bash | sudo tee /etc/bash_completion.d/vx

# Zsh
vx --completion zsh > ~/.oh-my-zsh/completions/_vx

# Fish
vx --completion fish > ~/.config/fish/completions/vx.fish

# PowerShell
vx --completion powershell >> $PROFILE
``` -->

## 📋 Command Reference

### File Operations
| Command | Description |
|---------|-------------|
| `vx <files...>` | Move files/directories to cache |
| `vx -r <pattern>` `vx --restore <pattern>` | Restore file based on patter so it can restore multiple files better use `vx -i` or `vx -l` and find exact fine to restore
| `vx -l` `vx --list` | Show all cached files |
| `vx -i <patern>` `vx --info <pattern>` | Detailed info about cached items |
| `vx -c` `vx --clear` | Empty entire cache |
| `vx -pr <days>` `vx --purge <days>` | Remove files older than N days |
| `vx -s` `vx --stats` | Cache usage statistics |
| `vx -p` `vx --path` | Show cache directory path |
| `vx -t` `vx --themes` | Interactive theme selector |
| `vx -cp` `vx --config-path` | Show config file location |
| `-f` `--noconfirm` | Skip all confirmation prompts |
| `-h` `--help` | Show help information |
| `-v` `--version` | Display version information |

## 📊 Pattern Matching

Vanish supports powerful pattern matching for restoration:

```bash
# Exact filename
vx --restore "document.pdf"

# Wildcard patterns
vx --restore "*.txt" "backup-*" "project-2024-*"

# Multiple patterns
vx --restore "*.log" "config.*" "test-*"
```

## 🛡️ Safety Features

- **Atomic Operations**: All moves are atomic to prevent data corruption
- **Path Validation**: Comprehensive checks prevent cache conflicts
- **Collision Detection**: Automatic handling of naming conflicts during restore
- **Permission Preservation**: File permissions and ownership maintained
- **Transaction Logging**: Complete audit trail of all operations
- **Recovery Verification**: Integrity checks during restoration

## 🚨 Important Notes

### ⚠️ Cache Directory Warning
**Never manually modify the cache directory structure.** If you need to change the cache location, use the configuration file and run `vx --clear` to empty the old location first.

### 🔒 Security Considerations
- Cache files maintain original permissions
- Vanish respects file system ACLs and extended attributes
- Symbolic links are preserved but not followed during deletion
- Hidden files require explicit specification (no accidental deletion)

## 🤝 Contributing

We welcome contributions!
<!--  Please see our [Contributing Guide](CONTRIBUTING.md) for details. -->

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/Notification`)
3. Commit your changes (`git commit -m 'Added Notification Feature'`)
4. Push to the branch (`git push origin feature/Notification`)
5. Open a Pull Request

## 🐛 Bug Reports & Feature Requests

- **Bug Reports**: [GitHub Issues](https://github.com/Aelune/vanish/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/Aelune/vanish/discussions)
- **Security Issues**: Please email security@yourdomain.com

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Inspired by modern CLI tools and safety-first design principles

---

<div align="center">

**[Homepage](https://dwukn.vercel.app/)** •
**[Documentation](https://dwukn.vercel.app/)** •
**[Releases](https://github.com/Aelune/vanish/releases)** •
<!-- **[Discussions](https://github.com/Aelune/vanish/discussions)** -->

Made with ❤️ by [Aelune](https://github.com/Aelune)

</div>
