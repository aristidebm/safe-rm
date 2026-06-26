package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/engine"

	"github.com/spf13/cobra"
)

var jsonOutput bool
var sortBy string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List trashed files",
	RunE: func(cmd *cobra.Command, args []string) error {
		indexDir, err := config.IndexDir()
		if err != nil {
			return err
		}

		entries, err := engine.ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
		if err != nil {
			return err
		}

		entries = filterPermanent(entries)
		sortEntries(entries, sortBy)

		if jsonOutput {
			return listJSON(entries)
		}

		return listTable(entries)
	},
}

func init() {
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "output raw JSONL")
	listCmd.Flags().StringVar(&sortBy, "sort", "date", "sort by: date, size, or name")
}

func filterPermanent(entries []*engine.TrashEntry) []*engine.TrashEntry {
	var out []*engine.TrashEntry
	for _, e := range entries {
		if !e.Permanent {
			out = append(out, e)
		}
	}
	return out
}

func sortEntries(entries []*engine.TrashEntry, by string) {
	switch by {
	case "size":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Size > entries[j].Size
		})
	case "name":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].OriginalPath < entries[j].OriginalPath
		})
	default:
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].TrashedAt.After(entries[j].TrashedAt)
		})
	}
}

func listJSON(entries []*engine.TrashEntry) error {
	enc := json.NewEncoder(os.Stdout)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

func listTable(entries []*engine.TrashEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSIZE\tTRASHED AT\tPATH")

	for _, e := range entries {
		sizeStr := formatSize(e.Size)
		timeStr := e.TrashedAt.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.ID, sizeStr, timeStr, e.OriginalPath)
	}

	return w.Flush()
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
