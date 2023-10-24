package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/streamer/types"
)

// InitGenesis initializes the incentives module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
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

// ExportGenesis returns the x/incentives module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		Streams:      k.GetNotFinishedStreams(ctx),
		LastStreamId: k.GetLastStreamID(ctx),
	}
}
