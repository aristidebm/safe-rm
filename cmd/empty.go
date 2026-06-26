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

		var cutoff time.Time
		if olderThan != "" {
			d, err := parseDuration(olderThan)
			if err != nil {
				return fmt.Errorf("invalid --older-than: %w", err)
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
			fmt.Fprintf(os.Stderr, "safe-rm: trash is already empty\n")
			return nil
		}

		if !emptyForce {
			var totalSize int64
			for _, e := range toRemove {
				totalSize += e.Size
			}
			fmt.Fprintf(os.Stderr, "safe-rm: about to remove %d entries (%s)\n", len(toRemove), formatSize(totalSize))
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
	emptyCmd.Flags().StringVar(&olderThan, "older-than", "", "only delete entries older than duration (e.g. 30d, 2w, 1h)")
	emptyCmd.Flags().BoolVarP(&emptyForce, "force", "f", false, "skip confirmation prompt")
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
