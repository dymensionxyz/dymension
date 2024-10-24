package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, elem := range genState.SequencerList {
		k.SetSequencer(ctx, elem)
		if elem.NoticeInProgress(ctx.BlockTime()) {
			k.AddToNoticeQueue(ctx, elem)
		}
	}

	for _, elem := range genState.GenesisProposers {
		k.SetProposer(ctx, elem.RollappId, elem.Address)
	}
}

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
