package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagInitSequencer   = "init-sequencer"
	FlagGenesisChecksum = "genesis-checksum"
	FlagNativeDenom     = "native-denom"
	FlagMetadata        = "metadata"
	FlagBech32Prefix    = "bech32-prefix"
)

// FlagSetUpdateRollapp returns flags for updating rollapps.
func FlagSetUpdateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagInitSequencer, "", "The address of the sequencer that will be used to initialize the rollapp")
	fs.String(FlagGenesisChecksum, "", "The checksum of the genesis file of the rollapp")
	fs.String(FlagNativeDenom, "", "The native denomination of the rollapp")
	fs.String(FlagMetadata, "", "The metadata of the rollapp")
	fs.String(FlagBech32Prefix, "", "Bech-32 prefix of the rollapp")

	return fs
}
