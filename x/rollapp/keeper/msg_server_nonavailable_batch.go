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
	nip, err := msg.DecodeNonInclusionProof()
	if err != nil {
		//TODO: handle deposit burn on wrong non-inclusion proof
		k.Logger(ctx).Info("unable to verif non-inclusion proof ", "rollappID", msg.RollappId)
		return nil, err
	}
	err = k.VerifyNonAvailableBatch(ctx, msg, &nip)
	if err != nil {
		//TODO: handle deposit burn on wrong non-inclusion proof
		k.Logger(ctx).Info("unable to verif non-inclusion proof ", "rollappID", msg.RollappId)
		return nil, err
	}

	return &types.MsgNonAvailableBatchResponse{}, nil
}
