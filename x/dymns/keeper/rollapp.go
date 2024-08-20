package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsRollAppId checks if the chain-id is a RollApp-Id.
func (k Keeper) IsRollAppId(ctx sdk.Context, chainId string) bool {
	_, found := k.rollappKeeper.GetRollapp(ctx, chainId)
	return found
}

// IsRollAppCreator returns true if the input bech32 address is the creator of the RollApp.
func (k Keeper) IsRollAppCreator(ctx sdk.Context, rollAppId, account string) bool {
	rollApp, found := k.rollappKeeper.GetRollapp(ctx, rollAppId)
	return found && rollApp.Owner == account
}

// GetRollAppBech32Prefix returns the Bech32 prefix of the RollApp by the chain-id.
func (k Keeper) GetRollAppBech32Prefix(ctx sdk.Context, chainId string) (bech32Prefix string, found bool) {
	rollApp, found := k.rollappKeeper.GetRollapp(ctx, chainId)
	if found && len(rollApp.Bech32Prefix) > 0 {
		return rollApp.Bech32Prefix, true
	}

	return "", false
}
