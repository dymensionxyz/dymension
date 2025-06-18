package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, elem := range genState.SequencerList {
		k.SetSequencer(ctx, elem)
		if err := k.SetSequencerByDymintAddr(ctx, elem.MustProposerAddr(), elem.Address); err != nil {
			panic(err)
		}
	}

	for _, s := range genState.NoticeQueue {
		seq := k.GetSequencer(ctx, s)
		k.AddToNoticeQueue(ctx, seq)
	}

	for _, elem := range genState.GenesisProposers {
		k.SetProposer(ctx, elem.RollappId, elem.Address)
	}
	for _, elem := range genState.GenesisSuccessors {
		k.SetSuccessor(ctx, elem.RollappId, elem.Address)
	}
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.GenesisState{}
	genesis.Params = k.GetParams(ctx)
	genesis.SequencerList = k.AllSequencers(ctx)

	proposers := k.AllProposers(ctx)
	for _, proposer := range proposers {
		genesis.GenesisProposers = append(genesis.GenesisProposers, types.GenesisProposer{
			RollappId: proposer.RollappId,
			Address:   proposer.Address,
		})
	}

	elems := k.AllSuccessors(ctx)
	for _, elem := range elems {
		genesis.GenesisSuccessors = append(genesis.GenesisSuccessors, types.GenesisProposer{
			RollappId: elem.RollappId,
			Address:   elem.Address,
		})
	}

	notice, err := k.NoticeQueue(ctx, nil)
	if err != nil {
		panic(err)
	}
	for _, seq := range notice {
		genesis.NoticeQueue = append(genesis.NoticeQueue, seq.Address)
	}

	return &genesis
}
