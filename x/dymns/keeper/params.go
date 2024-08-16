package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) dymnstypes.Params {
	return dymnstypes.NewParams(
		k.PriceParams(ctx),
		k.ChainsParams(ctx),
		k.MiscParams(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params dymnstypes.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	k.paramStore.SetParamSet(ctx, &params)
	return nil
}

// PriceParams returns the price params
func (k Keeper) PriceParams(ctx sdk.Context) (res dymnstypes.PriceParams) {
	k.paramStore.Get(ctx, dymnstypes.KeyPriceParams, &res)
	return
}

// ChainsParams returns the chains params
func (k Keeper) ChainsParams(ctx sdk.Context) (res dymnstypes.ChainsParams) {
	k.paramStore.Get(ctx, dymnstypes.KeyChainsParams, &res)
	return
}

// MiscParams returns the miscellaneous params
func (k Keeper) MiscParams(ctx sdk.Context) (res dymnstypes.MiscParams) {
	k.paramStore.Get(ctx, dymnstypes.KeyMiscParams, &res)
	return
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
		// Please read the `processActiveAliasSellOrders` method (hooks.go) for more information.
		return false
	}

	if isRollAppId := k.IsRollAppId(ctx, aliasCandidate); isRollAppId {
		return false
	}

	_, foundRollAppIdFromAlias := k.GetRollAppIdByAlias(ctx, aliasCandidate)
	if foundRollAppIdFromAlias {
		return false
	}

	return true
}
