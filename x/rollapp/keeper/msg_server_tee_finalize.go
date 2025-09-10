package keeper

import (
	"context"
	_ "embed"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
	///////////
	// TEE feature must be enabled, message from proposer etc
	///////////

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	teeConfig := params.TeeConfig

	if !teeConfig.Enabled {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled")
	}

	rollapp := msg.Nonce.RollappId

	_, found := k.GetRollapp(ctx, rollapp)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", rollapp)
	}

	proposer := k.SequencerK.GetProposer(ctx, rollapp)

	if proposer.Address != msg.Creator {
		return nil, gerrc.ErrPermissionDenied.Wrapf("only active sequencer can submit TEE attestation: expected %s, got %s",
			proposer.Address, msg.Creator)
	}

	///////////
	// TEE node must have started from a finalized state
	///////////

	if !k.IsHeightFinalized(ctx, rollapp, msg.Nonce.FinalizedHeight) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("claimed finalized height is not finalized")
	}

	///////////
	// TEE node must genuinely have reached the proposed new latest finalized state
	///////////

	if k.IsFinalizedIndex(ctx, rollapp, msg.StateIndex) {
		return nil, gerrc.ErrOutOfRange.Wrapf("state index is already finalized")
	}

	info, found := k.GetStateInfo(ctx, rollapp, msg.StateIndex)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("state info for rollapp: %s", rollapp)
	}

	if info.GetLatestHeight() != msg.Nonce.CurrHeight {
		return nil, gerrc.ErrInvalidArgument.Wrapf("height index mismatch")
	}

	err := k.ValidateAttestation(ctx, msg.Nonce.Hash(), msg.AttestationToken)
	if err != nil {
		return nil, errorsmod.Wrap(err, "validate attestation")
	}

	err = k.FastFinalizeRollappStatesUntilStateIndex(ctx, rollapp, msg.StateIndex)
	if err != nil {
		return nil, errorsmod.Wrap(err, "fast finalize states")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollapp),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", msg.StateIndex)),
		),
	)

	return &types.MsgFastFinalizeWithTEEResponse{}, nil
}
