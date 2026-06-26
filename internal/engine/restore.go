package engine

import "example.com/safe-rm/internal/config"

type ConflictStrategy int

const (
	ConflictRename ConflictStrategy = iota
	ConflictOverwrite
	ConflictSkip
)

func Restore(entry *TrashEntry, cfg *config.Config, onConflict ConflictStrategy) error {
	return nil
}
