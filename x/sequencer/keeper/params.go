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

// ValidateParams is stateful validation for params
// it checks if that UnbondingTime greater then x/rollapp's disputePeriod
// and that the correct denom is set
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
