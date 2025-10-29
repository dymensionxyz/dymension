package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// UpdateParams implements types.MsgServer.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the sender is the authority
	if req.Authority != m.authority {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can update params")
	}

	err := req.NewParams.ValidateBasic()
	if err != nil {
		return nil, err
	}
	/* -------------------------------------------------------------------------- */
	/*                 stateful validation for standard launch params                 */
	/* -------------------------------------------------------------------------- */
	standardLaunch := req.NewParams.StandardLaunch

	// validate target raise denom is allowed
	targetRaise := standardLaunch.TargetRaise
	targetRaiseMetadata, ok := m.BK.GetDenomMetaData(ctx, targetRaise.Denom)
	if !ok {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "denom %s not registered", targetRaise.Denom)
	}
	targetRaiseExponent := targetRaiseMetadata.DenomUnits[len(targetRaiseMetadata.DenomUnits)-1].Exponent

	// validate target raise denom is allowed
	if !slices.Contains(m.Keeper.gk.GetParams(ctx).AllowedPoolCreationDenoms, targetRaise.Denom) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom not allowed")
	}

	// validate curve is valid
	allocationDec := types.ScaleFromBase(standardLaunch.AllocationAmount, 18)
	evaluationDec := types.ScaleFromBase(targetRaise.Amount, int64(targetRaiseExponent)).MulInt64(2)
	liquidityPart := math.LegacyOneDec()

	// FIXME: review
	calculatedM := types.CalculateM(evaluationDec, allocationDec, standardLaunch.CurveExp, liquidityPart)
	if !calculatedM.IsPositive() {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "calculated M parameter is not positive: %s", calculatedM)
	}

	// Create bonding curve with calculated M and global parameters
	bondingCurve := types.NewBondingCurve(
		calculatedM,
		standardLaunch.CurveExp,
		math.LegacyZeroDec(),
		18,
		uint64(targetRaiseExponent),
	)

	// Validate the bonding curve
	if err := bondingCurve.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid bonding curve: %v", err.Error())
	}

	m.SetParams(ctx, req.NewParams)

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetParams sets the module parameters in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, b)
}

// GetParams returns the module parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ParamsKey)
	k.cdc.MustUnmarshal(b, &params)
	return params
}
