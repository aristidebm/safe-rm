# safe-rm Implementation TODO

## Phase 0 — Scaffold
- [x] Create directory tree
- [x] Create .envrc
- [x] Create go.mod with dependencies
- [x] Create main.go entry point
- [x] Create internal/log/log.go (file logging)
- [x] Create stub files for all packages
- [ ] Verify compile and commit

## Phase 1 — Nix Files
- [ ] nix/package.nix
- [ ] nix/shell.nix
- [ ] nix/module.nix
- [ ] flake.nix
- [ ] Verify and commit

## Phase 2 — Configuration
- [ ] internal/config/config.go (TOML struct, Load, ResolvedTrashDir, IndexDir)
- [ ] internal/config/config_test.go
- [ ] Tests pass and commit

## Phase 3 — Core Engine

### meta.go — Trash entry + JSONL index
- [ ] TrashEntry struct + NewEntry
- [ ] AppendEntry, ReadAllEntries, WriteAllEntries with flock
- [ ] Tests and commit

### glob.go — Glob matching
- [ ] MatchesAny with `**` support
- [ ] Tests and commit

### delete.go — Delete logic
- [ ] Policy routing (Route)
- [ ] SoftDelete (FreeDesktop + custom)
- [ ] PermanentDelete
- [ ] Recursive/force flag handling
- [ ] Tests and commit

### restore.go — Restore logic
- [ ] Restore with conflict strategies
- [ ] Tests and commit

## Phase 4 — CLI Commands
- [ ] cmd/root.go (config load, --debug flag)
- [ ] cmd/rm.go (main delete command)
- [ ] cmd/restore.go
- [ ] cmd/list.go
- [ ] cmd/empty.go
- [ ] Build and commit

## Phase 5 — TUI
- [ ] internal/tui/styles.go
- [ ] internal/tui/confirm.go
- [ ] internal/tui/restore.go
- [ ] Build and commit

## Phase 6 — Edge Case Handling
- [ ] Cross-device move fallback
- [ ] `.` and `/` rejection
- [ ] Empty args handling
- [ ] Lock timeout, malformed JSONL, missing content
- [ ] Tests and commit

## Phase 7 — Full Test Coverage
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Final test commit

## Phase 8 — README
- [ ] Write README.md
- [ ] Commit
