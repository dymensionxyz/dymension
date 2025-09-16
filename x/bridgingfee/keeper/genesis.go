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
	// Export fee hooks
	var feeHooks []types.HLFeeHook
	err := k.feeHooks.Walk(ctx, nil, func(key uint64, value types.HLFeeHook) (stop bool, err error) {
		feeHooks = append(feeHooks, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	// Export aggregation hooks
	var aggregationHooks []types.AggregationHook
	err = k.aggregationHooks.Walk(ctx, nil, func(key uint64, value types.AggregationHook) (stop bool, err error) {
		aggregationHooks = append(aggregationHooks, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		FeeHooks:         feeHooks,
		AggregationHooks: aggregationHooks,
	}
}