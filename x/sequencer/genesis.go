package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// InitGenesis initializes the sequencer module's state from a provided genesis
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Set all the sequencer
	for _, elem := range genState.SequencerList {
		k.SetSequencer(ctx, elem)

		// Set the unbonding queue for the sequencer
		if elem.Status == types.Unbonding {
			k.AddSequencerToUnbondingQueue(ctx, elem)
		} else if elem.IsNoticePeriodInProgress() {
			k.AddSequencerToNoticePeriodQueue(ctx, elem)
		}
	}

	for _, elem := range genState.GenesisProposers {
		k.SetProposer(ctx, elem.RollappId, elem.Address)
	}

	for _, bondReduction := range genState.BondReductions {
		k.SetDecreasingBondQueue(ctx, bondReduction)
	}
}

// ExportGenesis returns the sequencer module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.GenesisState{}
	genesis.Params = k.GetParams(ctx)
	genesis.SequencerList = k.GetAllSequencers(ctx)
	proposers := k.GetAllProposers(ctx)
	for _, proposer := range proposers {
		genesis.GenesisProposers = append(genesis.GenesisProposers, types.GenesisProposer{
			RollappId: proposer.RollappId,
			Address:   proposer.Address,
		})
	}

	return &genesis
}
