package cli

import (
	flag "github.com/spf13/pflag"
)

const (
// Create rollapp flags

)

// FlagSetCreateRollapp returns flags for creating gauges.
func FlagSetCreateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	return fs
}
