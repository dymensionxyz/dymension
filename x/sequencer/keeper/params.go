package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	HubExpectedTimePerBlock = 6 * time.Second
)

// ValidateParams is a stateful validation for params.
// it validates that unbonding time  is greater then x/rollapp's dispute period
// and that the correct denom is set.
// The unbonding time is set by governance hence it's more of a sanity/human error check which 
// in theory should never fail as setting such unbonding time has wide protocol security implications beyond the dispute period. 
func (k Keeper) ValidateParams(ctx sdk.Context, params types.Params) error {
	// validate unbonding time > dispute period
	rollappParams := k.rollappKeeper.GetParams(ctx)
	// Get the time duration of the dispute period
	disputeDuration := time.Duration(rollappParams.DisputePeriodInBlocks) * HubExpectedTimePerBlock // dispute period duration
	if params.UnbondingTime < disputeDuration {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "unbonding time must be greater than dispute period")
	}

	// validate min bond denom
	denom, err := sdk.GetBaseDenom()
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInternal, "failed to get base denom: %v", err)
	}
	if params.MinBond.Denom != denom {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min bond denom must be equal to base denom")
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

func (k Keeper) MinBond(ctx sdk.Context) (res sdk.Coin) {
	return k.GetParams(ctx).MinBond
}

func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	return k.GetParams(ctx).UnbondingTime
}

func (k Keeper) NoticePeriod(ctx sdk.Context) (res time.Duration) {
	return k.GetParams(ctx).NoticePeriod
}

func (k Keeper) LivenessSlashMultiplier(ctx sdk.Context) (res sdk.Dec) {
	return k.GetParams(ctx).LivenessSlashMultiplier
}
