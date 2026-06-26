package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/log"
)

type ConflictStrategy int

const (
	ConflictRename ConflictStrategy = iota
	ConflictOverwrite
	ConflictSkip
)

func Restore(entry *TrashEntry, cfg *config.Config, onConflict ConflictStrategy) error {
	trashDir, err := cfg.ResolvedTrashDir()
	if err != nil {
		return err
	}

	indexDir, err := config.IndexDir()
	if err != nil {
		return err
	}

	trashPath := filepath.Join(trashDir, "files", entry.ID)

	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return fmt.Errorf("trash content not found: %s", trashPath)
	}

	origPath := entry.OriginalPath

	if _, err := os.Stat(origPath); err == nil {
		switch onConflict {
		case ConflictSkip:
			log.Infof("skip restore %s -> %s (target exists)", trashPath, origPath)
			return nil
		case ConflictRename:
			origPath = resolveConflictPath(origPath)
		case ConflictOverwrite:
		}
	}

	parent := filepath.Dir(origPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	if err := moveOrCopy(trashPath, origPath); err != nil {
		return fmt.Errorf("restore move: %w", err)
	}

	if entry.FreeDesktop {
		infoPath := filepath.Join(trashDir, "info", entry.ID+".trashinfo")
		os.Remove(infoPath)
	}

	indexPath := filepath.Join(indexDir, "trash.jsonl")
	entries, err := ReadAllEntries(indexPath)
	if err != nil {
		log.Errorf("failed to read index after restore: %v", err)
		return nil
	}

	filtered := make([]*TrashEntry, 0, len(entries))
	for _, e := range entries {
		if e.ID != entry.ID {
			filtered = append(filtered, e)
		}
	}

	if err := WriteAllEntries(indexPath, filtered); err != nil {
		log.Errorf("failed to rewrite index after restore: %v", err)
		return nil
	}

	log.Infof("restored %s -> %s", trashPath, origPath)
	return nil
}

func resolveConflictPath(path string) string {
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s.%d", path, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
