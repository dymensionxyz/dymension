package cli

import (
	flag "github.com/spf13/pflag"
)

// FlagSetCreateRollapp returns flags for creating rollapps.
func FlagSetCreateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	return fs
}

// FlagSetUpdateRollapp returns flags for updating rollapps.
func FlagSetUpdateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String("init-sequencer", "", "The address of the sequencer that will be used to initialize the rollapp")
	fs.String("genesis-checksum", "", "The checksum of the genesis file of the rollapp")
	fs.String("alias", "", "The alias of the rollapp")
	fs.String("metadata", "", "The metadata of the rollapp")

	return fs
}
