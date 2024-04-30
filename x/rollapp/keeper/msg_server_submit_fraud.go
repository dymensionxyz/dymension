package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// SubmitFraud takes a fraud proof and checks that it is itself valid, then verifies it against the current state of the chain
// to check if a fraud actually occurred. If so, the fraudulent actor is punished.
func (k msgServer) SubmitFraud(goCtx context.Context, msg *types.MsgSubmitFraud) (*types.MsgSubmitFraudResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// load rollapp object for stateful validations
	_, isFound := k.GetRollapp(ctx, msg.RollappID)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}
	// FIXME: validate rollapp type/SW version is verifiable

	fp, err := msg.NativeFraudProof()
	if err != nil {
		return nil, fmt.Errorf("native fraud proof: %w", err)
	}

	err = k.ValidateAndRunFraudProof(ctx, msg.RollappID, fp)
	if err != nil {
		return nil, fmt.Errorf("validate and run fraud proof: %w", err)
	}

	k.Logger(ctx).Info("validated and ran fraud proof", "rollapp id", msg.RollappID)

	// FIXME: handle slashing

	// FIXME: handle deposit burn on wrong FP

	return &types.MsgSubmitFraudResponse{}, nil
}
