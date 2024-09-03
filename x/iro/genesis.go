package iro

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// check if the iro account exists so it could receive funds
	moduleAcc := k.AK.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	k.SetParams(ctx, genState.Params)

	lastPlanId := uint64(0)
	for _, plan := range genState.Plans {
		k.SetPlan(ctx, plan)
		if plan.Id > lastPlanId {
			lastPlanId = plan.Id
		}
	}
	k.SetLastPlanId(ctx, lastPlanId)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.GenesisState{}
	genesis.Params = k.GetParams(ctx)
	genesis.Plans = append(genesis.Plans, k.GetAllPlans(ctx)...)

	return &genesis
}
