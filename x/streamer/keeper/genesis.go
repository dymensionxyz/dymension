package keeper

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// InitGenesis initializes the streamer module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	recipientAcc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module account %s does not exist", types.ModuleName))
	}

	k.SetParams(ctx, genState.Params)

	slices.SortFunc(genState.Streams, CmpStreams)

	for _, stream := range genState.Streams {
		err := k.SetStreamWithRefKey(ctx, &stream)
		if err != nil {
			panic(err)
		}
	}

	k.SetLastStreamID(ctx, genState.LastStreamId)

	// Create epoch pointers for all epoch infos
	for _, epoch := range k.ek.AllEpochInfos(ctx) {
		err := k.SaveEpochPointer(ctx, types.NewEpochPointer(epoch.Identifier))
		if err != nil {
			panic(err)
		}
	}

	// Fill epoch pointers specified in the genesis
	for _, pointer := range genState.EpochPointers {
		err := k.SaveEpochPointer(ctx, pointer)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/streamer module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	pointers, err := k.GetAllEpochPointers(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		Params:        k.GetParams(ctx),
		Streams:       k.GetNotFinishedStreams(ctx),
		LastStreamId:  k.GetLastStreamID(ctx),
		EpochPointers: pointers,
	}
}
