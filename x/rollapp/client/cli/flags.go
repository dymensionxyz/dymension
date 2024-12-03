package cli

import (
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	flag "github.com/spf13/pflag"
)

const (
	FlagInitSequencer    = "init-sequencer"
	FlagMinSequencerBond = "min-sequencer-bond"
	FlagGenesisChecksum  = "genesis-checksum"
	FlagNativeDenom      = "native-denom"
	FlagInitialSupply    = "initial-supply"
	FlagMetadata         = "metadata"
	FlagBech32Prefix     = "bech32-prefix"
	FlagGenesisAccounts  = "genesis-accounts"
)

// FlagSetUpdateRollapp returns flags for updating rollapps.
func FlagSetUpdateRollapp() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagInitSequencer, "", "The address of the sequencer that will be used to initialize the rollapp")
	fs.String(FlagMinSequencerBond, rollapptypes.DefaultMinSequencerBondGlobalCoin.Amount.String(), "Minimum amount of bond required to be a sequencer in DYM (not adym)")
	fs.String(FlagGenesisChecksum, "", "The checksum of the genesis file of the rollapp")
	fs.String(FlagNativeDenom, "", "The native denomination of the rollapp")
	fs.String(FlagInitialSupply, "", "The initial supply of the rollapp")
	fs.String(FlagMetadata, "", "The metadata of the rollapp")
	fs.String(FlagBech32Prefix, "", "Bech32 prefix of the rollapp")
	fs.String(FlagGenesisAccounts, "", "<address>:<amount>,<address>:<amount>")

	return fs
}
