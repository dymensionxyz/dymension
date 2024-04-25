package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	// Create rollapp flags
	FlagGenesisAccountsPath = "genesis-accounts-path"
)

// FlagSetCreateRollapp returns flags for creating gauges.
func FlagSetCreateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagGenesisAccountsPath, "", "path to a json file containing genesis accounts")
	return fs
}
