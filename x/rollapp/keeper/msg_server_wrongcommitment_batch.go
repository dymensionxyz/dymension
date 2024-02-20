package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) SubmitWrongCommitmentBatch(goCtx context.Context, msg *types.MsgWrongCommitmentBatch) (*types.MsgWrongCommitmentBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// load rollapp object for stateful validations
	/*_, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}*/
	ip, err := msg.DecodeInclusionProof()
	if err != nil {
		return nil, err
	}

	err = k.VerifyWrongCommitmentBatch(ctx, msg, &ip)

	if err == nil {
		//FIXME: handle deposit burn on wrong FP
		k.Logger(ctx).Info("unable to verif wrong-commitment proof ", "rollappID", msg.RollappId)

	}
	//FIXME: handle slashing

	//FIXME: handle deposit burn on wrong proof

	return &types.MsgWrongCommitmentBatchResponse{}, nil
}
