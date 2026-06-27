package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/engine"
	"example.com/safe-rm/internal/log"

	"github.com/spf13/cobra"
)

var olderThan string
var emptyForce bool
var emptyDryRun bool

var emptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Empty the trash",
	RunE: func(cmd *cobra.Command, args []string) error {
		indexDir, err := config.IndexDir()
		if err != nil {
			return err
		}

		trashDir, err := cfg.ResolvedTrashDir()
		if err != nil {
			return err
		}

		entries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
		if err != nil {
			return err
		}

		age := olderThan
		if age == "" {
			age = cfg.DefaultMaxAge()
		}

		var cutoff time.Time
		if age != "" {
			d, err := parseDuration(age)
			if err != nil {
				return fmt.Errorf("invalid --older-than %q: %w", age, err)
			}
			cutoff = time.Now().UTC().Add(-d)
		}

		var toRemove []*engine.TrashEntry
		var remaining []*engine.TrashEntry

		for _, e := range entries {
			if !cutoff.IsZero() && e.TrashedAt.After(cutoff) {
				remaining = append(remaining, e)
				continue
			}
			toRemove = append(toRemove, e)
		}

		if len(toRemove) == 0 {
			if cutoff.IsZero() {
				fmt.Fprintf(os.Stderr, "safe-rm: trash is already empty\n")
			} else {
				fmt.Fprintf(os.Stderr, "safe-rm: no entries older than %s\n", age)
			}
			return nil
		}

		var totalSize int64
		for _, e := range toRemove {
			totalSize += e.Size
		}

		if cutoff.IsZero() {
			fmt.Fprintf(os.Stderr, "safe-rm: about to remove %d entries (%s)\n", len(toRemove), formatSize(totalSize))
		} else {
			fmt.Fprintf(os.Stderr, "safe-rm: %d entries (%s) older than %s\n", len(toRemove), formatSize(totalSize), age)
		}

		if emptyDryRun {
			fmt.Fprintf(os.Stderr, "safe-rm: dry-run, nothing removed\n")
			return nil
		}

		if !emptyForce {
			fmt.Fprintf(os.Stderr, "Proceed? [y/N] ")

			var response string
			fmt.Scanln(&response)
			response = strings.TrimSpace(strings.ToLower(response))
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

		fmt.Fprintf(os.Stderr, "safe-rm: removed %d entries from trash\n", len(toRemove))
		return nil
	},
}

func init() {
	emptyCmd.Flags().StringVar(&olderThan, "older-than", "", "only delete entries older than duration (e.g. 30d, 2w, 1h); defaults to config max_age")
	emptyCmd.Flags().BoolVarP(&emptyForce, "force", "f", false, "skip confirmation prompt")
	emptyCmd.Flags().BoolVar(&emptyDryRun, "dry-run", false, "show what would be removed without removing")
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty duration")
	}

	last := s[len(s)-1]
	var multiplier time.Duration

	switch last {
	case 'd':
		multiplier = 24 * time.Hour
	case 'w':
		multiplier = 7 * 24 * time.Hour
	case 'h':
		multiplier = time.Hour
	case 'm':
		multiplier = time.Minute
	default:
		return time.ParseDuration(s)
	}

	n, err := strconv.ParseInt(s[:len(s)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse duration %q", s)
	}

	return time.Duration(n) * multiplier, nil
}
