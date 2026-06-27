package cmd

import (
	"fmt"
	"os"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/log"
	"example.com/safe-rm/internal/tui"

	"github.com/spf13/cobra"
)

var (
	cfg         *config.Config
	debugMode   bool
	recursive   bool
	force       bool
	interactive bool
	verbose     bool
	oneFS       bool
)

var rootCmd = &cobra.Command{
	Use:   "safe-rm [flags] <files...>",
	Short: "A safe rm wrapper with trash support",
	Long: `safe-rm is a rm replacement that moves files to trash
instead of permanently deleting them. It follows the FreeDesktop
Trash specification and supports glob-based policies.`,
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := log.Init(debugMode); err != nil {
			return fmt.Errorf("log init: %w", err)
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		initTheme(cfg)

		if trashDir, err := cfg.ResolvedTrashDir(); err == nil {
			tui.SetTrashPath(trashDir)
		}

		return nil
	},
	RunE: rmRunE,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug logging")

	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "remove directories recursively (-R also accepted)")
	rootCmd.Flags().BoolVarP(&force, "force", "f", false, "ignore nonexistent files, never prompt")
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "prompt before every removal")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "explain what is being done")
	rootCmd.Flags().BoolVar(&oneFS, "one-file-system", false, "accepted, silently ignored")

	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(emptyCmd)
}

func initTheme(cfg *config.Config) {
	if cfg.Theme == nil {
		return
	}

	t := tui.DefaultTheme()
	if v := cfg.Theme.TitleFG; v != "" {
		t.Colors.TitleFG = v
	}
	if v := cfg.Theme.DangerFG; v != "" {
		t.Colors.DangerFG = v
	}
	if v := cfg.Theme.WarningFG; v != "" {
		t.Colors.WarningFG = v
	}
	if v := cfg.Theme.MutedFG; v != "" {
		t.Colors.MutedFG = v
	}
	if v := cfg.Theme.SelectedFG; v != "" {
		t.Colors.SelectedFG = v
	}
	if v := cfg.Theme.UnselectedFG; v != "" {
		t.Colors.UnselectedFG = v
	}
	if v := cfg.Theme.PermanentFG; v != "" {
		t.Colors.PermanentFG = v
	}
	if v := cfg.Theme.PermanentBG; v != "" {
		t.Colors.PermanentBG = v
	}
	if v := cfg.Theme.TrashFG; v != "" {
		t.Colors.TrashFG = v
	}
	if v := cfg.Theme.TrashBG; v != "" {
		t.Colors.TrashBG = v
	}
	if v := cfg.Theme.TrashPathFG; v != "" {
		t.Colors.TrashPathFG = v
	}
	if v := cfg.Theme.BorderColor; v != "" {
		t.Colors.BorderColor = v
	}
	if v := cfg.Theme.KeyHintFG; v != "" {
		t.Colors.KeyHintFG = v
	}
	tui.SetTheme(t)
}
