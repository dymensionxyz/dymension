package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Foo(context.Context, *types.QueryFooRequest) (*types.QueryFooResponse, error) {
	panic("unimplemented")
}

func (k Keeper) WithdrawalStatus(ctx context.Context, req *types.QueryWithdrawalStatusRequest) (*types.QueryWithdrawalStatusResponse, error) {
	ret := make([]types.WithdrawalStatus, len(req.WithdrawalId))

	for i, id := range req.WithdrawalId {
		if err := id.ValidateBasic(); err != nil {
			return nil, err
		}

		err := k.ValidateWithdrawal(ctx, *id)
		if err != nil {
			return nil, err
		}

		processed, err := k.processedWithdrawals.Has(ctx, collections.Join(id.MustMailboxId().GetInternalId(), id.MustMessageId().Bytes()))
		if err != nil {
			return nil, err
		}
		if processed {
			ret[i] = types.WithdrawalStatus_WITHDRAWAL_STATUS_PROCESSED
		} else {
			ret[i] = types.WithdrawalStatus_WITHDRAWAL_STATUS_UNPROCESSED
		}
	}

	return &types.QueryWithdrawalStatusResponse{Status: ret}, nil
}

func (k Keeper) ValidateWithdrawal(ctx context.Context, id types.WithdrawalID) error {
	dispatched, err := k.hypercoreK.Messages.Has(ctx, collections.Join(id.MustMailboxId().GetInternalId(), id.MustMessageId().Bytes()))
	if err != nil {
		return err
	}
	if !dispatched {
		return gerrc.ErrNotFound.Wrapf("message not dispatched")
	}

	return nil
}
