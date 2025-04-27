package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) dymnstypes.Params {
	store := ctx.KVStore(k.storeKey)
	var params dymnstypes.Params
	k.cdc.MustUnmarshal(store.Get(dymnstypes.KeyParams), &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params dymnstypes.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(dymnstypes.KeyParams, bz)
	return nil
}

// PriceParams returns the price params
func (k Keeper) PriceParams(ctx sdk.Context) (res dymnstypes.PriceParams) {
	return k.GetParams(ctx).Price
}

// ChainsParams returns the chains params
func (k Keeper) ChainsParams(ctx sdk.Context) (res dymnstypes.ChainsParams) {
	return k.GetParams(ctx).Chains
}

// MiscParams returns the miscellaneous params
func (k Keeper) MiscParams(ctx sdk.Context) (res dymnstypes.MiscParams) {
	return k.GetParams(ctx).Misc
}

// CanUseAliasForNewRegistration checks if the alias can be used for a new alias registration.
//
// It returns False when
//   - The format is invalid.
//   - The alias is exists in the params, mapped by a chain-id or registered to a Roll-App.
//   - The alias is equals to a known chain-id, from the params.
func (k Keeper) CanUseAliasForNewRegistration(ctx sdk.Context, aliasCandidate string) (can bool) {
	if !dymnsutils.IsValidAlias(aliasCandidate) {
		return false
	}

	if k.IsAliasPresentsInParamsAsAliasOrChainId(ctx, aliasCandidate) {
		// Please read the `processCompleteSellOrderWithAssetTypeAlias` method (msg_server_complete_sell_order.go) for more information.
		return false
	}

	if isRollAppId := k.IsRollAppId(ctx, aliasCandidate); isRollAppId {
		return false
	}

	_, foundRollAppIdFromAlias := k.GetRollAppIdByAlias(ctx, aliasCandidate)
	return !foundRollAppIdFromAlias
}
