package cli

import (
	flag "github.com/spf13/pflag"
)

// Flags for incentives module tx commands.
const (
	FlagStartTime       = "start-time"
	FlagEpochIdentifier = "epoch-identifier"
	FlagEpochs          = "epochs"
	FlagOwner           = "owner"
)

// FlagSetCreateStream returns flags for creating gauges.
func FlagSetCreateStream() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagStartTime, "", "Timestamp to begin distribution")
	fs.String(FlagEpochIdentifier, "day", "Epoch identifier to begin distribution (e.g. 'day', 'week')")
	fs.Uint64(FlagEpochs, 365, "Total epochs to distribute tokens")
	return fs
}
