package cmd

import "github.com/spf13/cobra"

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove files (default command)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
