package types

import simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

type SimSequencer struct {
	// Account is the account of the sequencer account.
	Account simtypes.Account
	// creator is the bech32-encoded address of the account sent the transaction (sequencer creator)
	Creator string
	// RollappIndex is the index of the rollapp in GlobalRollappList
	RollappIndex int
}
