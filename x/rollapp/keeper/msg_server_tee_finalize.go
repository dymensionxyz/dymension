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

func (k msgServer) ToggleTEE(goCtx context.Context, msg *types.MsgToggleTEE) (*types.MsgToggleTEEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, gerrc.ErrNotFound
	}

	if msg.Owner != rollapp.Owner {
		return nil, gerrc.ErrPermissionDenied
	}

	rollapp.EnableTee = msg.Enable
	k.SetRollapp(ctx, rollapp)

	return &types.MsgToggleTEEResponse{}, nil
}

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
	///////////
	// TEE feature must be enabled, message from proposer etc
	///////////

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	teeConfig := params.TeeConfig

	if !teeConfig.Enabled {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	if msg.Nonce.HubChainId != ctx.ChainID() {
		return nil, gerrc.ErrInvalidArgument.Wrapf("hub chain id does not match token nonce chain id: nonce: %s, actual; %s", msg.Nonce.HubChainId, ctx.ChainID())
	}

	rollapp := msg.Nonce.RollappId

	rol, found := k.GetRollapp(ctx, rollapp)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", rollapp)
	}
	if !rol.EnableTee {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled for rollapp")
	}

	///////////
	// TEE node must have started from a finalized state
	///////////

	fromGenesis := msg.Nonce.FinalizedHeight == 0
	fullNodeTrustedHeightOk := fromGenesis || k.IsHeightFinalized(ctx, rollapp, msg.Nonce.FinalizedHeight)

	if !fullNodeTrustedHeightOk {
		return nil, gerrc.ErrInvalidArgument.Wrapf("claimed finalized height is not finalized")
	}

	///////////
	// TEE node must genuinely have reached the proposed new latest finalized state
	///////////

	info, err := k.FindStateInfoByHeight(ctx, rollapp, msg.Nonce.CurrHeight)
	if err != nil {
		return nil, gerrc.ErrNotFound.Wrapf("state info for rollapp: %s", rollapp)
	}

	indexToFinalize := info.GetIndex().Index
	if info.GetLatestHeight() != msg.Nonce.CurrHeight {
		// its not the last block so we cant finalize everything in the state info yet
		indexToFinalize--
	}

	// Avoid letting txs through which would do a 'costly'
	// attestation while not making progress
	if k.IsIndexFinalized(ctx, rollapp, indexToFinalize) {
		return nil, gerrc.ErrInvalidArgument.Wrap("index is already finalized")
	}

	if teeConfig.Verify {
		err = k.ValidateAttestation(ctx, msg.Nonce.Hash(), msg.AttestationToken)
		if err != nil {
			return nil, errorsmod.Wrap(err, "validate attestation")
		}
	}

	err = k.FastFinalizeRollappStatesUntilStateIndex(ctx, rollapp, indexToFinalize)
	if err != nil {
		return nil, errorsmod.Wrap(err, "fast finalize states")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollapp),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", indexToFinalize)),
		),
	)

	return &types.MsgFastFinalizeWithTEEResponse{}, nil
}
