Here’s a clean and structured Markdown file you can use as documentation for Vanish’s configuration:

````markdown
# Vanish - Configuration Documentation

This document describes all available configuration options for **Vanish** (`vx`), a safe file/directory removal tool. Use this as a reference to customize the behavior, appearance, and safety features of Vanish.

---

## ⚠️ Important Warning

**Do not modify the cache directory if it already contains files.**
The `directory` setting below determines where Vanish stores deleted files. If you point this to a folder that already contains important data, Vanish will treat it as its cache. Operations such as **restore**, **purge**, and **clear** may not work as expected and could result in **data loss**.

---

## Cache Settings

```toml
[cache]
directory = ".cache/vanish"
days      = 10
````

| Key         | Type   | Default         | Description                                                      |
| ----------- | ------ | --------------- | ---------------------------------------------------------------- |
| `directory` | string | `.cache/vanish` | Relative path to store deleted files (relative to your `$HOME`). |
| `days`      | int    | `10`            | Number of days to keep deleted files before automatic cleanup.   |

---

## Logging

```toml
[logging]
enabled   = true
directory = ".cache/vanish/logs"
```

| Key         | Type   | Default              | Description                                         |
| ----------- | ------ | -------------------- | --------------------------------------------------- |
| `enabled`   | bool   | `true`               | Enable or disable logging.                          |
| `directory` | string | `.cache/vanish/logs` | Directory for log files (relative to cache folder). |

---

## User Interface (UI) Settings

```toml
[ui]
theme      = "default"
no_confirm = false
```

| Key          | Type   | Default     | Description                                                                                                                 |
| ------------ | ------ | ----------- | --------------------------------------------------------------------------------------------------------------------------- |
| `theme`      | string | `"default"` | Theme for the UI. Options: `"default"`, `"dark"`, `"light"`, `"cyberpunk"`, `"minimal"`, `"ocean"`, `"forest"`, `"sunset"`. |
| `no_confirm` | bool   | `false`     | If `true`, skips confirmation prompts. Use with caution!                                                                    |

---

## UI Color Customization

```toml
[ui.colors]
# primary   = "#3B82F6"
# secondary = "#6366F1"
# success   = "#10B981"
# warning   = "#F59E0B"
# error     = "#EF4444"
# text      = "#F9FAFB"
# muted     = "#9CA3AF"
# border    = "#374151"
# highlight = "#FBBF24"
```

You can uncomment and modify the hex codes to customize the UI colors.

| Key         | Description            |
| ----------- | ---------------------- |
| `primary`   | Main accent color      |
| `secondary` | Secondary accent color |
| `success`   | Success messages       |
| `warning`   | Warning messages       |
| `error`     | Error messages         |
| `text`      | Main text color        |
| `muted`     | Muted/help text        |
| `border`    | Border color           |
| `highlight` | Highlighted filename   |

---

## Progress Bar Settings

```toml
[ui.progress]
# style      = "gradient"
# show_emoji = true
# animation  = true
```

| Key          | Type   | Description                                                        |
| ------------ | ------ | ------------------------------------------------------------------ |
| `style`      | string | Progress bar style: `"gradient"`, `"solid"`, `"rainbow"`           |
| `show_emoji` | bool   | If `true`, adds emoji to progress messages                         |
| `animation`  | bool   | Smooth animation. Disable for better performance on slow terminals |

---

## Summary

The configuration file allows full control over:

* **Cache behavior** (where files are stored and retention)
* **Logging** (enable/disable, location)
* **UI theme & colors** (appearance customization)
* **Progress bar** (style, emojis, animation)

Always back up your config file before making changes, especially if your cache already contains files.

---

> **Tip:** Vanish automatically creates default configuration in `$HOME/.config/vanish/vanish.toml` if none exists.

```

---

If you want, I can also **expand this MD** to include all other future sections like **Notifications**, **Advanced Options**, and **Safety Settings**, so it becomes a full `docs/config.md` for your repo.

Do you want me to do that?
```
