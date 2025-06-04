package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/kas/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
)

type msgServer struct {
	*Keeper
}

func (k *Keeper) IndicateProgress(goCtx context.Context, req *types.MsgIndicateProgress) (*types.MsgIndicateProgressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.Ready(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "transactions disabled")
	}

	////////////
	//// Verify

	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}

	threshold, vals := k.MustValidators(ctx)
	metadata := req.MustGetMetadata()
	payload := req.Payload
	digest := payload.MustGetSignBytes()

	ok, err := hypercoretypes.VerifyMultisig(vals, threshold, metadata.Signatures, digest)
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "verify multisig")
	}
	if !ok {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrUnauthenticated, err), "verify multisig")
	}

	////////////
	//// Update

	// CAS
	currentOutpoint, err := k.outpoint.Get(ctx)
	if !payload.OldOutpoint.Equal(&currentOutpoint) {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrFailedPrecondition, err), "old outpoint")
	}

	err = k.outpoint.Set(ctx, *payload.NewOutpoint)
	if err != nil {
		return nil, err
	}

	for _, withdrawal := range payload.ProcessedWithdrawals {
		err = k.ValidateWithdrawal(ctx, *withdrawal)
		if err != nil {
			// should never happen, it means validators are buggy or protocol is broken
			return nil, errorsmod.Wrap(gerrc.ErrFault, "withdrawal not dispatched")
		}
		err = k.processedWithdrawals.Set(ctx, collections.Join(k.MustMailbox(ctx), withdrawal.MustMessageId().Bytes()))
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgIndicateProgressResponse{}, nil
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
