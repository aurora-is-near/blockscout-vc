package main

import (
	"fmt"
	"os"

	"blockscout-vc/cmd"
)

func main() {
	// Initialize the root command
	c := cmd.RootCmd()
	// Add the sidecar subcommand
	c.AddCommand(cmd.StartSidecarCmd())

	// Execute the command and handle any errors
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing Blockscout CLI '%s'", err)
		os.Exit(1)
	}
}
