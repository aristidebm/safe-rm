package cmd

import "github.com/spf13/cobra"

var emptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Empty the trash",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
