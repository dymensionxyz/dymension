package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

// InitGenesis initializes the capability module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Set fee hooks
	for _, hook := range genState.FeeHooks {
		if err := k.feeHooks.Set(ctx, hook.Id.GetInternalId(), hook); err != nil {
			panic(err)
		}
	}

	// Set aggregation hooks
	for _, hook := range genState.AggregationHooks {
		if err := k.aggregationHooks.Set(ctx, hook.Id.GetInternalId(), hook); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	iter, err := k.feeHooks.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	feeHooks, err := iter.Values()
	if err != nil {
		panic(err)
	}

	iter1, err := k.aggregationHooks.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	aggregationHooks, err := iter1.Values()
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		FeeHooks:         feeHooks,
		AggregationHooks: aggregationHooks,
	}
}
