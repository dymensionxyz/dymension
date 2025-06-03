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
		mailboxId, err := id.DecodeMailboxId()
		if err != nil {
			return nil, err
		}

		messageId, err := id.DecodeMessageId()
		if err != nil {
			return nil, err
		}

		// Note: this will also find received messages
		// It's up to the client to make sure the mailbox is the right one for the Kas bridge
		dispatched, err := k.hypercoreK.Messages.Has(ctx, collections.Join(mailboxId.GetInternalId(), messageId.Bytes()))
		if err != nil {
			return nil, err
		}

		if !dispatched {
			return nil, gerrc.ErrNotFound.Wrapf("message not dispatched: %s", messageId)
		}

		processed, err := k.processedWithdrawals.Has(ctx, collections.Join(mailboxId.GetInternalId(), messageId.Bytes()))
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
