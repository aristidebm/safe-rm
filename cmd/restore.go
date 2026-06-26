package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/engine"
	"example.com/safe-rm/internal/log"
	"example.com/safe-rm/internal/tui"

	"github.com/spf13/cobra"
)

var conflictStr string

var restoreCmd = &cobra.Command{
	Use:   "restore [id]",
	Short: "Restore trashed files",
	Long: `Restore a trashed file by ID, or launch the restore browser TUI.

Flags:
  --on-conflict [rename|overwrite|skip]  behavior when target exists (default: rename)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		strategy := parseConflictStrategy(conflictStr)

		if len(args) > 0 {
			id := args[0]
			return restoreByID(id, strategy)
		}

		return runRestoreTUI(strategy)
	},
}

func init() {
	restoreCmd.Flags().StringVar(&conflictStr, "on-conflict", "rename", "behavior when target exists: rename, overwrite, or skip")
}

func parseConflictStrategy(s string) engine.ConflictStrategy {
	switch strings.ToLower(s) {
	case "overwrite":
		return engine.ConflictOverwrite
	case "skip":
		return engine.ConflictSkip
	default:
		return engine.ConflictRename
	}
}

func runRestoreTUI(strategy engine.ConflictStrategy) error {
	indexDir, err := config.IndexDir()
	if err != nil {
		return err
	}

	entries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "safe-rm: trash is empty\n")
		return nil
	}

	result := tui.RunRestore(entries)
	if result.Aborted {
		return nil
	}

	for _, e := range result.ToRestore {
		if err := engine.Restore(e, cfg, result.Conflict); err != nil {
			log.Errorf("failed to restore %s: %v", e.ID, err)
			fmt.Fprintf(os.Stderr, "safe-rm: failed to restore %s: %v\n", e.ID, err)
		} else {
			fmt.Fprintf(os.Stderr, "safe-rm: restored %s\n", e.OriginalPath)
		}
	}

	for _, e := range result.ToDelete {
		if err := deleteFromTrash(e); err != nil {
			log.Errorf("failed to delete %s from trash: %v", e.ID, err)
			fmt.Fprintf(os.Stderr, "safe-rm: failed to delete %s from trash: %v\n", e.ID, err)
		} else {
			fmt.Fprintf(os.Stderr, "safe-rm: deleted %s from trash\n", e.OriginalPath)
		}
	}

	return nil
}

func restoreByID(id string, strategy engine.ConflictStrategy) error {
	indexDir, err := config.IndexDir()
	if err != nil {
		return err
	}

	entries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.ID == id {
			if err := engine.Restore(e, cfg, strategy); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "safe-rm: restored %s\n", e.OriginalPath)
			return nil
		}
	}

	return fmt.Errorf("no trashed file with ID %q", id)
}

func deleteFromTrash(e *engine.TrashEntry) error {
	trashDir, err := cfg.ResolvedTrashDir()
	if err != nil {
		return err
	}

	indexDir, err := config.IndexDir()
	if err != nil {
		return err
	}

	contentPath := filepath.Join(trashDir, "files", e.ID)
	if err := os.RemoveAll(contentPath); err != nil {
		return err
	}

	if e.FreeDesktop {
		infoPath := filepath.Join(trashDir, "info", e.ID+".trashinfo")
		os.Remove(infoPath)
	}

	allEntries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if err != nil {
		return err
	}

	var remaining []*engine.TrashEntry
	for _, entry := range allEntries {
		if entry.ID != e.ID {
			remaining = append(remaining, entry)
		}
	}

	return engine.WriteAllEntries(filepath.Join(indexDir, "trash.jsonl"), remaining)
}
