```toml
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
```
