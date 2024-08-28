package iro

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	_ = k.AK.GetModuleAccount(ctx, types.ModuleName) // called to ensure the module account is set
	k.SetParams(ctx, genState.Params)

	for _, plan := range genState.Plan {
		k.SetPlan(ctx, plan)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	plans := k.GetAllPlans(ctx)
	for _, plan := range plans {
		genesis.Plan = append(genesis.Plan, plan)
	}

	return genesis
}
