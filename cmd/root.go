package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "blockscout-vc",
		Short: "Blockscout Virtual Chain toolset",
		Long:  `Blockscout Virtual Chain toolset`,
		// Default behavior is to show help
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	return rootCmd
}
