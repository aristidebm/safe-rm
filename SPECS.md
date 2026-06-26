# safe-rm — AI Agent Implementation Plan

## Overview

You are implementing `safe-rm`, a `rm` wrapper written in Go that moves files to
trash instead of permanently deleting them. It follows the FreeDesktop Trash
specification by default, supports glob-based policies, and provides a bubbletea
TUI for dangerous operations and restoration.

**Stack**:
- Language: Go (latest stable)
- CLI: `cobra`
- TUI: `bubbletea` + `bubbles` + `lipgloss`
- Config: `toml` (`github.com/BurntSushi/toml`)
- Nix: flakes with devShell, package, NixOS module, home-manager module

**Non-negotiables**:
- Follow XDG Base Directory specification for all paths
- Follow FreeDesktop Trash specification when using default trash location
- Mirror the POSIX `rm` interface exactly (flags, exit codes, stderr format)
- All TUI actions must also be available as CLI flags/subcommands

---

## Phase 0 — Repository Scaffold

### 0.1 Directory structure

Create the following directory tree (empty files where noted):

```
safe-rm/
├── flake.nix
├── nix/
│   ├── shell.nix
│   ├── package.nix
│   └── module.nix
├── go.mod
├── go.sum                  (generated, do not create manually)
├── main.go
├── cmd/
│   ├── root.go
│   ├── rm.go
│   ├── restore.go
│   ├── list.go
│   └── empty.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── engine/
│   │   ├── delete.go
│   │   ├── restore.go
│   │   ├── glob.go
│   │   └── meta.go
│   └── tui/
│       ├── confirm.go
│       ├── restore.go
│       └── styles.go
└── README.md
```

### 0.2 go.mod

Module path: `github.com/yourorg/safe-rm`

Required dependencies (run `go get` for each):
```
github.com/spf13/cobra
github.com/BurntSushi/toml
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss
golang.org/x/sys
```

### 0.3 main.go

Minimal entry point:
```go
package main

import "github.com/yourorg/safe-rm/cmd"

func main() {
    cmd.Execute()
}
```

---

## Phase 1 — Nix Files

### 1.1 `nix/package.nix`

A standard `buildGoModule` derivation:
- `pname = "safe-rm"`
- `version = "0.1.0"`
- `src = ./.` (relative to flake)
- `vendorHash` set to `lib.fakeHash` initially — the agent must note that the
  maintainer needs to update this after first `go mod vendor`
- `meta.description = "A safe rm wrapper with trash support"`
- `meta.license = lib.licenses.mit`
- `meta.mainProgram = "safe-rm"`

### 1.2 `nix/shell.nix`

A `mkShell` with:
- `buildInputs`: `go`, `gopls`, `gotools`, `golangci-lint`
- `shellHook`: prints "safe-rm dev shell" and sets `GOPATH` if not set

### 1.3 `nix/module.nix`

A NixOS + home-manager compatible module. Expose the following options under
`programs.safe-rm`:

| Option | Type | Default | Maps to TOML key |
|---|---|---|---|
| `enable` | bool | false | — |
| `trashDir` | `types.nullOr types.str` | null | `trash_dir` |
| `bypassList` | `types.listOf types.str` | `[]` | `bypass_list` |
| `dangerList` | `types.listOf types.str` | `[]` | `danger_list` |

Implementation notes:
- When `enable = true`, add the package to `home.packages` or
  `environment.systemPackages` depending on which module system is active
- Generate `$XDG_CONFIG_HOME/safe-rm/config.toml` from the option values using
  `pkgs.writeText` or `home.file`
- Only write `trash_dir` to the file if `trashDir != null`

### 1.4 `flake.nix`

Outputs:
```nix
{
  packages.x86_64-linux.default   = import ./nix/package.nix { inherit pkgs; };
  packages.aarch64-linux.default  = ...;
  packages.x86_64-darwin.default  = ...;
  packages.aarch64-darwin.default = ...;
  devShells.*.default             = import ./nix/shell.nix { inherit pkgs; };
  nixosModules.default            = import ./nix/module.nix;
  homeManagerModules.default      = import ./nix/module.nix;
}
```

Use `flake-utils` for system iteration to avoid repeating per-system blocks.
Input: `nixpkgs` (nixos-unstable) and `flake-utils`.

---

## Phase 2 — Configuration (`internal/config/config.go`)

### 2.1 TOML schema (Go struct)

```go
type Config struct {
    TrashDir   *string  `toml:"trash_dir"`   // nil → use FreeDesktop default
    BypassList []string `toml:"bypass_list"`
    DangerList []string `toml:"danger_list"`
}
```

### 2.2 Path resolution (XDG)

Implement `func Load() (*Config, error)` that:

1. Determines config file path in this priority order:
   - `$SAFE_RM_CONFIG` env var (if set and file exists)
   - `$XDG_CONFIG_HOME/safe-rm/config.toml`
   - `~/.config/safe-rm/config.toml` (XDG fallback)

2. If no config file exists, returns a zero-value `Config{}` with no error
   (all defaults apply).

3. Parses the TOML file into `Config`.

4. Expands `~` and env vars in `TrashDir` if set.

### 2.3 Trash directory resolution

Implement `func (c *Config) ResolvedTrashDir() (string, error)` that:

- If `TrashDir` is set in config → return that path (FreeDesktop layout NOT used)
- Otherwise → return `$XDG_DATA_HOME/Trash` falling back to
  `~/.local/share/Trash`

### 2.4 Index directory

Implement `func IndexDir() (string, error)` that always returns:
- `$XDG_DATA_HOME/safe-rm/` falling back to `~/.local/share/safe-rm/`

This is where `trash.jsonl` lives regardless of trash location.

---

## Phase 3 — Core Engine

### 3.1 `internal/engine/meta.go` — Trash entry

```go
type TrashEntry struct {
    ID           string    `json:"id"`
    OriginalPath string    `json:"original_path"`
    TrashedAt    time.Time `json:"trashed_at"`
    Size         int64     `json:"size"`
    IsDir        bool      `json:"is_dir"`
    Checksum     string    `json:"checksum"` // sha256 hex, empty for dirs
    Permanent    bool      `json:"permanent"` // true = was bypass_list match, should not appear in trash
    FreeDesktop  bool      `json:"free_desktop"` // true = .trashinfo was written
}
```

Implement:
- `func NewEntry(originalPath string, isDir bool) (*TrashEntry, error)`:
  generates a short random ID (6 hex chars), computes checksum for files,
  sets `TrashedAt` to `time.Now().UTC()`
- `func AppendEntry(indexPath string, e *TrashEntry) error`: acquires an
  exclusive `flock` on `trash.jsonl`, appends one JSON line, releases lock
- `func ReadAllEntries(indexPath string) ([]*TrashEntry, error)`: reads and
  parses every line, skips malformed lines with a warning to stderr
- `func WriteAllEntries(indexPath string, entries []*TrashEntry) error`:
  acquires exclusive `flock`, rewrites entire file atomically (write to
  `.tmp`, then `os.Rename`), releases lock

### 3.2 `internal/engine/glob.go` — Glob matching

Implement `func MatchesAny(path string, patterns []string) (bool, error)`:
- Expand `~` in each pattern before matching
- Use `path/filepath.Match` for non-`**` patterns
- Implement `**` support manually: if a pattern contains `**`, split on `**`
  and check prefix/suffix matching against the absolute path
- Return the first match found; return `false, nil` if none match

### 3.3 `internal/engine/delete.go` — Delete logic

#### Policy routing

Implement `func Route(path string, cfg *config.Config) (Policy, error)` where:

```go
type Policy int
const (
    PolicySoftDelete Policy = iota
    PolicyPermanent
    PolicyDanger      // danger only, destination = trash
    PolicyDangerPermanent // danger + bypass = confirm then permanent delete
)
```

Logic (strict order):
1. Check `danger_list` first
2. Check `bypass_list` second
3. Combine results:
   - danger=true, bypass=true  → `PolicyDangerPermanent`
   - danger=true, bypass=false → `PolicyDanger`
   - danger=false, bypass=true → `PolicyPermanent`
   - danger=false, bypass=false → `PolicySoftDelete`

#### Soft delete

Implement `func SoftDelete(path string, cfg *config.Config, freeDesktop bool) error`:

1. Resolve absolute path
2. Compute destination in trash:
   - FreeDesktop mode: `<trashDir>/files/<ID>` for the content,
     `<trashDir>/info/<ID>.trashinfo` for metadata
   - Custom mode: `<trashDir>/files/<ID>`
3. Create parent directories if needed (`os.MkdirAll`)
4. `os.Rename(src, dst)` — if rename fails (cross-device), fallback to
   copy+remove
5. If FreeDesktop mode: write `.trashinfo` file:
   ```ini
   [Trash Info]
   Path=/original/absolute/path
   DeletionDate=2026-06-26T14:32:00
   ```
6. Append `TrashEntry` to `trash.jsonl`

#### Permanent delete

Implement `func PermanentDelete(path string) error`:
- `os.RemoveAll(path)` — no metadata written, no JSONL entry

#### Recursive flag handling

`delete.go` must respect:
- `-r` / `-R` / `--recursive`: allow directories
- Without recursive flag + directory target → print error to stderr, exit 1
  (matches POSIX `rm` behaviour)
- `-f` / `--force`: suppress errors on non-existent files, never prompt

### 3.4 `internal/engine/restore.go` — Restore logic

Implement `func Restore(entry *TrashEntry, cfg *config.Config, onConflict ConflictStrategy) error`:

```go
type ConflictStrategy int
const (
    ConflictRename    ConflictStrategy = iota // default
    ConflictOverwrite
    ConflictSkip
)
```

Steps:
1. Resolve trash content path (`<trashDir>/files/<entry.ID>`)
2. Check if `entry.OriginalPath` already exists:
   - `ConflictRename`: append `.1`, `.2`, etc. until free
   - `ConflictOverwrite`: proceed, `os.Rename` will overwrite
   - `ConflictSkip`: print message, return nil
3. `os.MkdirAll` parent of original path
4. `os.Rename(trashPath, entry.OriginalPath)` with cross-device fallback
5. If FreeDesktop: delete `<trashDir>/info/<entry.ID>.trashinfo`
6. Remove entry from `trash.jsonl` (read → filter → rewrite)

---

## Phase 4 — CLI Commands (`cmd/`)

### 4.1 `cmd/root.go`

- Create cobra root command named `safe-rm`
- Load config in `PersistentPreRunE`, store in a package-level var accessible
  to all subcommands
- Register all subcommands here

### 4.2 `cmd/rm.go` — Default command (the main one)

This is the command invoked when the user runs `safe-rm <files...>`.
It must be the root command's `RunE`, not a subcommand, so that
`safe-rm file.txt` works exactly like `rm file.txt`.

**Flags** (mirror POSIX `rm` exactly):

| Flag | Description |
|---|---|
| `-r`, `-R`, `--recursive` | Remove directories recursively |
| `-f`, `--force` | Ignore nonexistent files, never prompt |
| `-i` | Always prompt via TUI regardless of policy |
| `-v`, `--verbose` | Print each file as it is processed |
| `--one-file-system` | Accepted, silently ignored |

**Execution logic**:

```
For each path in args:
  1. Resolve to absolute path
  2. Stat the path (handle -f: skip if not exists)
  3. If directory and not -r: stderr + exit 1
  4. Call engine.Route(path, cfg) → policy
  5. If -i flag: override policy to PolicyDanger (force TUI)
  6. Collect all paths and their policies into a []DeleteTarget

After collecting all targets:
  7. Separate into:
     - directTargets: PolicySoftDelete + PolicyPermanent (no TUI needed)
     - dangerTargets: PolicyDanger + PolicyDangerPermanent (need TUI)

  8. If dangerTargets is non-empty:
     → Launch confirmation TUI with dangerTargets
     → TUI returns a []ConfirmedTarget (user selections)
     → For each confirmed: execute appropriate delete

  9. For directTargets: execute without TUI
```

**Exit codes**: match POSIX `rm` (0 = success, 1 = any error)

### 4.3 `cmd/restore.go`

Two modes:

**With argument** (`safe-rm restore <id>`):
- Flags: `--on-conflict [rename|overwrite|skip]` (default: rename)
- Look up entry by ID in `trash.jsonl`
- Call `engine.Restore(entry, cfg, strategy)`

**Without argument** (`safe-rm restore`):
- Launch restore TUI browser
- TUI returns selected entries + chosen conflict strategy per entry
- Execute restores

### 4.4 `cmd/list.go`

`safe-rm list`

Flags:
- `--json`: output raw JSONL instead of formatted table
- `--sort [date|size|name]` (default: date, newest first)

Default output (table):
```
ID       SIZE     TRASHED AT            PATH
a3f9c1   4.0 KB   2026-06-26 14:32:00   ~/code/myproject/
b7e2a4   812 B    2026-06-26 15:01:12   ~/docs/notes.txt
```

### 4.5 `cmd/empty.go`

`safe-rm empty`

Flags:
- `--older-than <duration>` (e.g. `30d`, `2w`, `1h`): only delete entries
  older than duration; parse `d`, `w`, `h`, `m` suffixes
- `--force` / `-f`: skip confirmation prompt
- Without `--force`: print count + total size, ask y/N on stdout before proceeding

Steps:
1. Read all entries from `trash.jsonl`
2. Apply `--older-than` filter if set
3. For each entry: `os.RemoveAll` the trash content path,
   delete `.trashinfo` if FreeDesktop
4. Rewrite `trash.jsonl` without removed entries

---

## Phase 5 — TUI (`internal/tui/`)

### 5.1 `internal/tui/styles.go`

Define all lipgloss styles here. No styles should be defined inline in other
files. Required styles:

```go
var (
    StyleTitle        lipgloss.Style  // bold, accent color
    StyleDanger       lipgloss.Style  // bold red, used for danger header
    StyleWarning      lipgloss.Style  // yellow
    StyleMuted        lipgloss.Style  // dim/gray
    StyleSelected     lipgloss.Style  // green checkbox prefix
    StyleUnselected   lipgloss.Style  // dim checkbox prefix
    StylePermanent    lipgloss.Style  // red badge "PERMANENT"
    StyleTrash        lipgloss.Style  // blue badge "TRASH"
    StyleBorder       lipgloss.Style  // rounded border container
    StyleKeyHint      lipgloss.Style  // bottom key hint bar
)
```

Use `lipgloss.AdaptiveColor` for all colors so the TUI works on both light
and dark terminals.

### 5.2 `internal/tui/confirm.go` — Deletion confirmation TUI

**Input**: `[]ConfirmItem`:
```go
type ConfirmItem struct {
    Path        string
    IsDir       bool
    Policy      engine.Policy  // PolicyDanger or PolicyDangerPermanent
    Expanded    bool           // for directories: whether children are shown
    Children    []string       // populated lazily on expand
    Selected    bool
}
```

**Layout**:
```
┌─────────────────────────────────────────────────────┐
│  ⚠  DANGER — Review files before deleting           │
├─────────────────────────────────────────────────────┤
│  [x] ~/code/myproject/          📁  TRASH           │
│  [ ]   ├── main.go              📄  (expanded child) │
│  [ ]   └── go.mod               📄                  │
│  [x] ~/.ssh/Password.kdbx       🔒  PERMANENT        │
│  [x] ~/.env                     📄  PERMANENT        │
├─────────────────────────────────────────────────────┤
│  3 selected · 1 permanent · 2 to trash              │
│  space: toggle  enter: expand dir  a: all  n: none  │
│  q/esc: abort   x: confirm deletion                 │
└─────────────────────────────────────────────────────┘
```

**Key bindings**:

| Key | Action |
|---|---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `space` | Toggle selected on cursor item |
| `enter` | Expand/collapse directory (lazy: read children on first expand) |
| `a` | Select all |
| `n` | Deselect all |
| `x` | Confirm and delete selected items |
| `q` / `esc` | Abort — delete nothing |

**Return value**: `([]ConfirmItem, bool)` — the items with updated `Selected`
fields, and a bool indicating whether the user confirmed (false = aborted).

**Directory expansion**: on first `enter` on a directory item, call
`os.ReadDir` and populate `Children`. Insert child rows into the list
immediately below the parent with a tree prefix (`├──`, `└──`). Children
are not individually selectable for policy purposes but are shown for
transparency. Toggling the parent toggles all children visually.

### 5.3 `internal/tui/restore.go` — Restore browser TUI

**Input**: `[]*engine.TrashEntry`

**Layout**:
```
┌─────────────────────────────────────────────────────┐
│  🗑  Trash — select files to restore                │
├─────────────────────────────────────────────────────┤
│  Filter: _                                          │
├─────────────────────────────────────────────────────┤
│  [ ] ~/code/myproject/     📁   4.0 KB   2026-06-26 │
│  [x] ~/docs/notes.txt      📄   812 B    2026-06-26 │
├─────────────────────────────────────────────────────┤
│  1 selected                                         │
│  /: filter  space: toggle  a: all  n: none          │
│  r: restore selected  d: delete selected  q: abort  │
└─────────────────────────────────────────────────────┘
```

**Key bindings**:

| Key | Action |
|---|---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `space` | Toggle selected |
| `/` | Focus filter input (fuzzy match on path) |
| `esc` | Blur filter input / abort |
| `a` | Select all (filtered) |
| `n` | Deselect all |
| `r` | Restore selected — opens conflict strategy picker if any conflict |
| `d` | Permanently delete selected from trash (with inline confirmation prompt) |
| `q` | Quit without action |

**Conflict strategy picker**: when `r` is pressed and a conflict exists,
show an inline prompt below the list:
```
  Conflict: ~/docs/notes.txt already exists
  [R]ename   [O]verwrite   [S]kip
```
Wait for `r`, `o`, or `s` keypress. Apply chosen strategy to all conflicts
in the current restore batch.

**Filter**: typing after pressing `/` filters the list in real time using
simple `strings.Contains` on the original path (case-insensitive). Press
`esc` or `enter` to exit filter mode; the filtered view persists.

**Return value**:
```go
type RestoreResult struct {
    ToRestore []*engine.TrashEntry
    ToDelete  []*engine.TrashEntry
    Conflict  ConflictStrategy
    Aborted   bool
}
```

---

## Phase 6 — Error Handling & Edge Cases

Implement these throughout all phases:

| Case | Behaviour |
|---|---|
| Source file disappeared between stat and delete | Log warning, continue if `-f`, exit 1 otherwise |
| Cross-device move (rename fails with EXDEV) | Copy recursively then `os.RemoveAll` source |
| `trash.jsonl` malformed line | Skip line, print warning to stderr, continue |
| Trash content missing for a JSONL entry | Show `[missing]` in list/restore TUI, allow removing the orphaned JSONL entry with `d` |
| `trash.jsonl` write lock timeout (>2s) | Return error: "trash index is locked by another process" |
| Config file exists but is invalid TOML | Fatal error with file path and parse error |
| Path argument is `.` or `/` | Always reject with explicit error message, even with `-rf` |
| Empty args | Print usage, exit 0 (match `rm` behaviour) |

---

## Phase 7 — Testing

### Unit tests (table-driven, `_test.go` files alongside source):

- `internal/engine/glob_test.go`: test `MatchesAny` against patterns with
  `*`, `**`, `?`, `~`, absolute paths, relative paths
- `internal/engine/meta_test.go`: test append + read + rewrite roundtrip,
  test concurrent append with goroutines (verify no corruption)
- `internal/config/config_test.go`: test XDG fallback logic, `~` expansion,
  `$SAFE_RM_CONFIG` override

### Integration tests:

- `TestSoftDelete`: create temp file → `SoftDelete` → verify file moved to
  trash, JSONL entry written, `.trashinfo` written if FreeDesktop
- `TestRestore`: soft-delete → restore → verify file back at original path,
  JSONL entry removed
- `TestPolicyRouting`: verify all four policy combinations are correctly
  resolved for paths matching danger/bypass lists
- `TestCrossDeviceMove`: skip if only one device available; verify copy+remove
  fallback works

---

## Phase 8 — README.md

Write a README covering:
1. Installation (Nix flake one-liner, go install)
2. Configuration (full TOML example with comments)
3. Usage examples for all commands and flags
4. FreeDesktop compatibility note
5. Differences from `rm`
6. How the `danger_list` and `bypass_list` interact (the priority explanation
   with the `Password.kdbx` example)

---

## Implementation Order

Execute phases strictly in this order. Do not start a phase until the previous
compiles and its tests pass:

```
Phase 0  →  Phase 1  →  Phase 2  →  Phase 3  →  Phase 4  →  Phase 5  →  Phase 6  →  Phase 7  →  Phase 8
scaffold     nix          config       engine       cli          tui          edge         tests        docs
```

Within Phase 3, implement in this sub-order:
`meta.go` → `glob.go` → `delete.go` → `restore.go`

Within Phase 5, implement in this sub-order:
`styles.go` → `confirm.go` → `restore.go`

---

## Key Invariants — Never Violate These

1. `danger_list` is checked **before** `bypass_list`. A path in both always
   gets the confirmation TUI first, then permanent delete on confirm.
2. Files matching `bypass_list` **never** enter the trash folder or appear
   in `trash.jsonl`.
3. The JSONL index (`trash.jsonl`) lives in `$XDG_DATA_HOME/safe-rm/`
   always, regardless of where the trash content lives.
4. When `trash_dir` is not set in config, use `$XDG_DATA_HOME/Trash` and
   write `.trashinfo` files (FreeDesktop compliant).
5. When `trash_dir` is set in config, do **not** write `.trashinfo` files.
6. `safe-rm . ` and `safe-rm /` must always be rejected, no flags override this.
7. All TUI actions are also available as CLI flags/subcommands — the TUI is
   never the only path to an action.
8. Write to `trash.jsonl` only while holding an exclusive `flock`.
