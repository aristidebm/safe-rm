package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/engine"
	"example.com/safe-rm/internal/log"

	"github.com/spf13/cobra"
)

var pruneOlderThan string
var pruneDryRun bool
var pruneForce bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune old entries from trash",
	Long: `Remove trash entries older than a specified duration.

If --older-than is not provided, the configured max_age from config.toml is used.
The duration format is: 30d, 2w, 24h, 90m, etc.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		indexDir, err := config.IndexDir()
		if err != nil {
			return err
		}

		trashDir, err := cfg.ResolvedTrashDir()
		if err != nil {
			return err
		}

		age := pruneOlderThan
		if age == "" {
			age = cfg.DefaultMaxAge()
		}
		if age == "" {
			return fmt.Errorf("no duration specified; use --older-than or set max_age in config.toml")
		}

		d, err := parseDuration(age)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", age, err)
		}

		cutoff := time.Now().UTC().Add(-d)

		entries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
		if err != nil {
			return err
		}

		var toRemove []*engine.TrashEntry
		var remaining []*engine.TrashEntry

		for _, e := range entries {
			if e.TrashedAt.Before(cutoff) {
				toRemove = append(toRemove, e)
			} else {
				remaining = append(remaining, e)
			}
		}

		if len(toRemove) == 0 {
			fmt.Fprintf(os.Stderr, "safe-rm: no entries older than %s\n", age)
			return nil
		}

		var totalSize int64
		for _, e := range toRemove {
			totalSize += e.Size
		}

		fmt.Fprintf(os.Stderr, "safe-rm: %d entries (%s) older than %s\n", len(toRemove), formatSize(totalSize), age)

		if pruneDryRun {
			fmt.Fprintf(os.Stderr, "safe-rm: dry-run, nothing removed\n")
			return nil
		}

		if !pruneForce {
			fmt.Fprintf(os.Stderr, "Proceed? [y/N] ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "yes" {
				fmt.Fprintf(os.Stderr, "safe-rm: cancelled\n")
				return nil
			}
		}

		for _, e := range toRemove {
			contentPath := filepath.Join(trashDir, "files", e.ID)
			if err := os.RemoveAll(contentPath); err != nil {
				log.Errorf("failed to remove trash content %s: %v", contentPath, err)
			}

			if e.FreeDesktop {
				infoPath := filepath.Join(trashDir, "info", e.ID+".trashinfo")
				os.Remove(infoPath)
			}
		}

		if err := engine.WriteAllEntries(filepath.Join(indexDir, "trash.jsonl"), remaining); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "safe-rm: pruned %d entries older than %s\n", len(toRemove), age)
		return nil
	},
}

func init() {
	pruneCmd.Flags().StringVar(&pruneOlderThan, "older-than", "", "prune entries older than duration (e.g. 30d, 2w, 1h); defaults to config max_age")
	pruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "show what would be removed without removing")
	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "skip confirmation prompt")
}
