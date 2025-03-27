package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) escrowFromModule(ctx sdk.Context, srcModule string, c sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, srcModule, types.ModuleName, c)
}

func (k Keeper) escrowFromUser(ctx sdk.Context, srcAcc sdk.AccAddress, c sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, srcAcc, types.ModuleName, c)
}

func (k Keeper) refundFromModule(ctx sdk.Context, dstAddr sdk.AccAddress, c sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, dstAddr, c)
}
