package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// ValidateParams is a stateful validation for params.
// it validates that unbonding time  is greater then x/rollapp's dispute period
// and that the correct denom is set.
// The unbonding time is set by governance hence it's more of a sanity/human error check which
// in theory should never fail as setting such unbonding time has wide protocol security implications beyond the dispute period.
func (k Keeper) ValidateParams(ctx sdk.Context, params types.Params) error {
	// validate min bond denom
	denom, err := sdk.GetBaseDenom()
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInternal, "failed to get base denom: %v", err)
	}
	if params.MinBond.Denom != denom {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min bond denom must be equal to base denom")
	}
	if params.KickThreshold.Denom != denom {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "kick threshold denom must be equal to base denom")
	}
	return nil
}

// SetParams sets the auth module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

// GetParams gets the auth module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}
