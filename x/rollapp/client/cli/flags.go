package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	// Create rollapp flags

	FlagGenesisTransfersEnabled = "transfers-enabled"
)

// FlagSetCreateRollapp returns flags for creating gauges.
func FlagSetCreateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.Bool(FlagGenesisTransfersEnabled, false, "Enable ibc transfers immediately. Must be false if using genesis transfers (genesis accounts).")
	return fs
}
