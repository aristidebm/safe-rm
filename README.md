# safe-rm

A `rm` wrapper written in Go that moves files to trash instead of permanently
deleting them. Follows the [FreeDesktop Trash specification][fdo-trash] by
default, supports glob-based policies, and provides a TUI for dangerous
operations and restoration.

[fdo-trash]: https://specifications.freedesktop.org/trash-spec/trashspec-latest.html

## Installation

### Nix flake

```sh
nix profile install github:yourorg/safe-rm
# or, as a dependency in your flake:
#   inputs.safe-rm.url = "github:yourorg/safe-rm";
```

### Go install

```sh
go install example.com/safe-rm@latest
```

## Configuration

Create `~/.config/safe-rm/config.toml` (or set `$SAFE_RM_CONFIG`):

```toml
# Custom trash directory (default: $XDG_DATA_HOME/Trash)
# When set, .trashinfo files are NOT written (non-FreeDesktop mode)
trash_dir = "/path/to/custom/trash"

# Default age threshold for `safe-rm empty --older-than` (optional)
max_age = "30d"

# Patterns that bypass the trash entirely (permanent delete)
bypass_list = [
  "*.tmp",
  "*.log",
  "*.cache",
  "~/Downloads/*",
]

# Patterns that require TUI confirmation before deletion
danger_list = [
  "*.secret",
  "*.key",
  "Password.kdbx",
  "~/.ssh/*",
]

# Theme customization — override any TUI color (optional)
[theme]
title_fg = "#e0e0ff"
danger_fg = "#ff4444"
warning_fg = "#ffcc00"
muted_fg = "#666666"
selected_fg = "#44ff44"
unselected_fg = "#555555"
permanent_fg = "#ff4444"
permanent_bg = "#440000"
trash_fg = "#44aaff"
trash_bg = "#002244"
trash_path_fg = "#666666"
border_color = "#555555"
keyhint_fg = "#999999"
```

### Policy interaction

`danger_list` is checked **first**, then `bypass_list`. The four possible
outcomes:

| danger_list | bypass_list | Policy |
|---|---|---|
| No match | No match | Soft delete (move to trash) |
| No match | Match      | Permanent delete (no trash) |
| Match | No match | TUI confirmation → soft delete |
| Match | Match      | TUI confirmation → permanent delete |

Example: `~/.ssh/Password.kdbx` in both lists → user sees a TUI prompt, and
only after confirming is the file permanently deleted.

## Usage

### Deletion (default command)

```sh
safe-rm file.txt              # move to trash
safe-rm -r directory/         # recursively trash a directory
safe-rm -f file.txt           # ignore non-existent, never prompt
safe-rm -i file.txt           # force TUI confirmation
safe-rm -v file.txt           # verbose output
safe-rm --debug file.txt      # enable debug logging to ~/.local/state/safe-rm/safe-rm.log
```

### Listing trashed files

```sh
safe-rm list                  # table view (default: sort by date, newest first)
safe-rm list --json           # JSONL output
safe-rm list --sort size      # sort by size
safe-rm list --sort name      # sort by path
```

### Restoring files

```sh
safe-rm restore <id>                      # restore by ID
safe-rm restore <id> --on-conflict skip   # skip if target exists
safe-rm restore                           # launch TUI browser
```

### Emptying the trash

```sh
safe-rm empty                             # prompt before removing all
safe-rm empty -f                          # skip confirmation
safe-rm empty --older-than 30d            # only entries older than 30 days
safe-rm empty --older-than 2w             # 2 weeks
safe-rm empty --dry-run                   # show what would be removed
```

### TUI key bindings

**Confirmation TUI** (shown for `danger_list` matches):

| Key | Action |
|---|---|
| `↑`/`k` `↓`/`j` | Navigate |
| `space` | Toggle selection |
| `enter` | Expand/collapse directory |
| `a` | Select all |
| `n` | Deselect all |
| `x` | Confirm deletion |
| `q`/`esc` | Abort |

**Restore TUI** (`safe-rm restore`):

| Key | Action |
|---|---|
| `↑`/`k` `↓`/`j` | Navigate |
| `space` | Toggle selection |
| `/` | Filter (type to search, `esc` to exit) |
| `a` | Select all (filtered) |
| `n` | Deselect all |
| `r` | Restore selected |
| `d` | Permanently delete selected from trash |
| `q`/`esc` | Quit |

## FreeDesktop compatibility

When `trash_dir` is **not** set in config:
- Trash content goes to `$XDG_DATA_HOME/Trash/files/`
- Metadata (`.trashinfo`) goes to `$XDG_DATA_HOME/Trash/info/`
- Compatible with other FreeDesktop trash implementations

When `trash_dir` **is** set:
- Content goes to `<trash_dir>/files/`
- No `.trashinfo` files are written
- The JSONL index at `$XDG_DATA_HOME/safe-rm/trash.jsonl` tracks all entries

## Differences from `rm`

- **Safety**: Files go to trash by default, not permanently deleted
- **Policies**: Glob-based rules control which files are trashed vs. deleted
- **TUI**: Dangerous operations require interactive confirmation
- **Restore**: Trashed files can be restored by ID
- **Flags**: Mirrors POSIX `rm` flags exactly (`-r`, `-f`, `-i`, `-v`)
- **Exit codes**: 0 on success, 1 on any error
- **`.` and `/`**: Always rejected with an explicit error, even with `-rf`
