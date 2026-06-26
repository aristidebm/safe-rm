package cmd

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List trashed files",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
