package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

// InitGenesis initializes the streamer module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {

	k.SetParams(ctx, genState.Params)
	for _, denom := range genState.Denommetadatas {
		err := k.SetDenomMetadataWithRefKey(ctx, &denom)
		if err != nil {
			panic(err)
		}
	}
	k.SetLastDenomMetadataID(ctx, genState.LastDenommetadataId)
}

// ExportGenesis returns the x/streamer module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		Denommetadatas:      k.GetAllDenomMetadata(ctx),
		LastDenommetadataId: k.GetLastDenomMetadataID(ctx),
	}
}
