package eibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	// Validate all other genesis fields are empty
	genesisFields := []int{
		len(genState.DemandOrders),
	}
	for _, fieldLength := range genesisFields {
		if fieldLength != 0 {
			panic("Only params can be initialized at genesis")
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	// Add the demand orders
	allDemandOrders, err := k.ListAllDemandOrders(ctx)
	if err != nil {
		panic(err)
	}
	genesis.DemandOrders = make([]types.DemandOrder, len(allDemandOrders))
	for i, order := range allDemandOrders {
		genesis.DemandOrders[i] = *order
	}

	return genesis
}
