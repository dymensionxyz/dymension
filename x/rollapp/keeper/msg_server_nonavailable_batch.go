package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) SubmitNonAvailableBatch(goCtx context.Context, msg *types.MsgNonAvailableBatch) (*types.MsgNonAvailableBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// load rollapp object for stateful validations
	_, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}

	err := k.VerifyNonAvailableBatch(ctx, msg)

	if err == nil {
		//FIXME: handle deposit burn on wrong FP
		k.Logger(ctx).Info("unable to verif non-available proof ", "rollappID", msg.RollappId)

	}

	switch err {
	case types.ErrWrongCommitment:
		k.Logger(ctx).Info("wrong commitment proof verified", "rollappID", msg.RollappId)

		//FIXME: handle slashing
	case types.ErrBatchNotAvailable:
		k.Logger(ctx).Info("non-available proof verified", "rollappID", msg.RollappId)

		//FIXME: handle slashing
	case types.ErrInvalidBlobData:
		k.Logger(ctx).Info("invalid blob data verified", "rollappID", msg.RollappId)

		//FIXME: handle slashing
	}

	return &types.MsgNonAvailableBatchResponse{}, nil
}
