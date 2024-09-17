package keeper

import (
	"context"

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

	err := req.NewParams.Validate()
	if err != nil {
		return nil, err
	}

	m.Keeper.SetParams(ctx, req.NewParams)

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
	if b == nil {
		panic("params should have been set")
	}

	k.cdc.MustUnmarshal(b, &params)
	return params
}
