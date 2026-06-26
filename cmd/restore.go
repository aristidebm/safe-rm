package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/engine"

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

		// TODO: Phase 5 — launch restore TUI
		return fmt.Errorf("restore browser TUI not yet available; use safe-rm restore <id>")
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
