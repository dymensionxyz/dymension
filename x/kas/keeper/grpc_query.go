package keeper

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Outpoint(goCtx context.Context, req *types.QueryOutpointRequest) (*types.QueryOutpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.Ready(ctx) {
		return nil, gerrc.ErrFailedPrecondition.Wrap("queries disabled")
	}

	return &types.QueryOutpointResponse{
		Outpoint: uptr.To(k.MustOutpoint(ctx)),
	}, nil
}

func (k Keeper) WithdrawalStatus(goCtx context.Context, req *types.QueryWithdrawalStatusRequest) (*types.QueryWithdrawalStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.Ready(ctx) {
		return nil, gerrc.ErrFailedPrecondition.Wrap("queries disabled")
	}

	ret := make([]types.WithdrawalStatus, len(req.WithdrawalId))

	for i, id := range req.WithdrawalId {
		if err := id.ValidateBasic(); err != nil {
			return nil, err
		}

		err := k.ValidateWithdrawal(ctx, *id)
		if err != nil {
			return nil, err
		}

		processed, err := k.processedWithdrawals.Has(ctx, collections.Join(k.MustMailbox(ctx), id.MustMessageId().Bytes()))
		if err != nil {
			return nil, err
		}
		if processed {
			ret[i] = types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED
		} else {
			ret[i] = types.WithdrawalStatus_WITHDRAWAL_STATUS_UNPROCESSED
		}
	}

	return &types.QueryWithdrawalStatusResponse{
		Status:   ret,
		Outpoint: uptr.To(k.MustOutpoint(ctx)),
	}, nil
}

func (k Keeper) ValidateWithdrawal(ctx sdk.Context, id types.WithdrawalID) error {
	dispatched, err := k.hypercoreK.Messages.Has(ctx, collections.Join(k.MustMailbox(ctx), id.MustMessageId().Bytes()))
	if err != nil {
		return err
	}
	if !dispatched {
		return gerrc.ErrNotFound.Wrapf("message not dispatched")
	}

	return nil
}
