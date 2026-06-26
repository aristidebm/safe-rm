package cmd

import "github.com/spf13/cobra"

var restoreCmd = &cobra.Command{
	Use:   "restore [id]",
	Short: "Restore trashed files",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
