package tui

import "example.com/safe-rm/internal/engine"

type RestoreResult struct {
	ToRestore []*engine.TrashEntry
	ToDelete  []*engine.TrashEntry
	Conflict  engine.ConflictStrategy
	Aborted   bool
}

func RunRestore(entries []*engine.TrashEntry) *RestoreResult {
	return &RestoreResult{Aborted: true}
}
