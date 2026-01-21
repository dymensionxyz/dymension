package main

import (
	"fmt"
	"os"

	// SDK & CLI Framework
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	
	// Dymension Core
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/cmd/dymd/cmd"
)

/**
 * Main function: The bootstrap entry point for the dymd node.
 * It initializes the root command and executes the server command line.
 */
func main() {
	// Initialize the root command structure which includes all subcommands (init, start, tx, query)
	rootCmd := cmd.NewRootCmd()

	// Execute the command logic. 
	// DefaultNodeHome provides the standard directory (e.g., ~/.dymension)
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		// Optimization: Ensure errors are written to Stderr and exit with code 1 for CI/CD compatibility
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
