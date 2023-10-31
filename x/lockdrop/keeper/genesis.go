package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	recipientAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if recipientAcc == nil {
		panic(fmt.Sprintf("module account %s does not exist", types.ModuleName))
	}

	k.SetParams(ctx, genState.Params)
	if genState.DistrInfo == nil {
		k.SetDistrInfo(ctx, types.DistrInfo{
			TotalWeight: sdk.NewInt(0),
			Records:     nil,
		})
	} else {
		k.SetDistrInfo(ctx, *genState.DistrInfo)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	distrInfo := k.GetDistrInfo(ctx)

	return &types.GenesisState{
		Params:    k.GetParams(ctx),
		DistrInfo: &distrInfo,
	}
}
