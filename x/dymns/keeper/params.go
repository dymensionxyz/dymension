package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) dymnstypes.Params {
	return dymnstypes.NewParams(
		k.PriceParams(ctx),
		k.ChainsParams(ctx),
		k.MiscParams(ctx),
		k.PreservedRegistrationParams(ctx),
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

// PreservedRegistrationParams returns the preserved registration params
func (k Keeper) PreservedRegistrationParams(ctx sdk.Context) (res dymnstypes.PreservedRegistrationParams) {
	k.paramStore.Get(ctx, dymnstypes.KeyPreservedRegistrationParams, &res)
	return
}

// CheckChainIsCoinType60ByChainId checks if the chain-id is a CoinType60 chain-id, defined in the params.
func (k Keeper) CheckChainIsCoinType60ByChainId(ctx sdk.Context, chainId string) bool {
	if k.IsRollAppId(ctx, chainId) {
		// all RollApps on Dymension use secp256k1
		return true
	}

	for _, coinType60ChainId := range k.ChainsParams(ctx).CoinType60ChainIds {
		if coinType60ChainId == chainId {
			return true
		}
	}

	return false
}

// CanUseAliasForNewRegistration checks if the alias can be used for a new alias registration.
//
// It returns False when
//   - The format is invalid.
//   - The alias is exists in the params, mapped by a chain-id or registered to a Roll-App.
//   - The alias is equals to a known chain-id, from the params.
func (k Keeper) CanUseAliasForNewRegistration(ctx sdk.Context, aliasCandidate string) (can bool, err error) {
	// TODO DymNS: use this in RollApp registration

	if !dymnsutils.IsValidAlias(aliasCandidate) {
		err = gerrc.ErrInvalidArgument.Wrap("alias candidate")
		return
	}

	params := k.GetParams(ctx)

	for _, aliasesOfChainId := range params.Chains.AliasesOfChainIds {
		if aliasCandidate == aliasesOfChainId.ChainId {
			can = false
			return
		}

		for _, alias := range aliasesOfChainId.Aliases {
			if aliasCandidate == alias {
				can = false
				return
			}
		}
	}

	for _, chainId := range params.Chains.CoinType60ChainIds {
		if aliasCandidate == chainId {
			can = false
			return
		}
	}

	if isRollAppId := k.IsRollAppId(ctx, aliasCandidate); isRollAppId {
		can = false
		return
	}

	_, found := k.GetRollAppIdByAlias(ctx, aliasCandidate)
	if found {
		can = false
		return
	}

	can = true
	return
}
