package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) DecreaseBond(goCtx context.Context, msg *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.GetRealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}
	if err := k.tryUnbond(ctx, &seq, msg.GetDecreaseAmount()); err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	// TODO: write object
	return &types.MsgDecreaseBondResponse{}, nil
}
