package keeper

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
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
	ix := msg.StateIndex

	_, found := k.GetRollapp(ctx, rollapp)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", rollapp)
	}

	proposer := k.SequencerK.GetProposer(ctx, rollapp)

	if proposer.Address != msg.Creator {
		return nil, gerrc.ErrPermissionDenied.Wrapf("only active sequencer can submit TEE attestation: expected %s, got %s",
			proposer.Address, msg.Creator)
	}

	if k.IsFinalizedIndex(ctx, rollapp, ix) {
		return nil, gerrc.ErrOutOfRange.Wrapf("state index is already finalized")
	}

	info, found := k.GetStateInfo(ctx, rollapp, ix)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("state info for rollapp: %s", rollapp)
	}

	if info.GetLatestHeight() != uint64(msg.Nonce.Height) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("height index mismatch")
	}

	bd, _ := info.LastBlockDescriptor()

	if !bytes.Equal(bd.StateRoot, msg.Nonce.StateRoot) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("state root mismatch")
	}

	err := k.validateAttestation(ctx, teeConfig.GcpRootCertPem, msg.Nonce.Hash(), msg.AttestationToken)
	if err != nil {
		return nil, errorsmod.Wrap(err, "validate attestation")
	}

	err = k.FastFinalizeRollappStatesUntilStateIndex(ctx, rollapp, ix)
	if err != nil {
		return nil, errorsmod.Wrap(err, "fast finalize states")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollapp),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", ix)),
		),
	)

	return &types.MsgFastFinalizeWithTEEResponse{}, nil
}