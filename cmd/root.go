package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "safe-rm",
	Short: "A safe rm wrapper with trash support",
	Long: `safe-rm is a rm replacement that moves files to trash
instead of permanently deleting them. It follows the FreeDesktop
Trash specification and supports glob-based policies.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
