package keeper

import (
	"fmt"

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
	for _, stream := range genState.Streams {
		stream := stream
		err := k.SetStreamWithRefKey(ctx, &stream)
		if err != nil {
			panic(err)
		}
	}
	k.SetLastStreamID(ctx, genState.LastStreamId)
}

// ExportGenesis returns the x/streamer module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		Streams:      k.GetNotFinishedStreams(ctx),
		LastStreamId: k.GetLastStreamID(ctx),
	}
}
