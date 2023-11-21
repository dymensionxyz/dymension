package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/streamer/types"
)

func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bk.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}
