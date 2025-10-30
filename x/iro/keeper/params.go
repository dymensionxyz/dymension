package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"
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
	allowedDenoms := m.Keeper.gk.GetParams(ctx).AllowedPoolCreationDenoms
	if !slices.Contains(allowedDenoms, standardLaunch.TargetRaise.Denom) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom not allowed")
	}

	// validate the curve is valid, for all allowed liquidity denoms
	for _, denom := range allowedDenoms {
		bondingCurve, _, err := m.GetCurveByLiquidityDenom(ctx, denom, standardLaunch)
		if err != nil {
			return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "failed to get standard launch curve and graduation point: %v", err.Error())
		}
		if err := bondingCurve.ValidateBasic(); err != nil {
			return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid bonding curve: %v", err.Error())
		}
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
