package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (k msgServer) TransferOwnership(goCtx context.Context, msg *dymnstypes.MsgTransferOwnership) (*dymnstypes.MsgTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateTransferOwnership(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err := k.transferOwnership(ctx, *dymName, msg.NewOwner); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgTransferOwnershipResponse{}, nil
}

func (k msgServer) validateTransferOwnership(ctx sdk.Context, msg *dymnstypes.MsgTransferOwnership) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	if dymName.IsExpiredAtContext(ctx) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("Dym-Name is already expired")
	}

	so := k.GetSellOrder(ctx, msg.Name)
	if so != nil {
		// by ignoring SO, can fall into case that SO not completed/lost funds of bidder,...

		return nil, sdkerrors.ErrInvalidRequest.Wrap("can not transfer ownership while there is an active Sell Order")
	}

	return dymName, nil
}

func (k msgServer) transferOwnership(ctx sdk.Context, dymName dymnstypes.DymName, newOwner string) error {
	if err := k.PruneDymName(ctx, dymName.Name); err != nil {
		return err
	}

	newDymNameRecord := dymnstypes.DymName{
		Name:       dymName.Name,
		Owner:      newOwner,         // transfer ownership
		Controller: newOwner,         // transfer controller
		ExpireAt:   dymName.ExpireAt, // keep the same expiration date
		Configs:    nil,              // clear configs
		Contact:    "",               // clear contact
	}

	if err := k.SetDymName(ctx, newDymNameRecord); err != nil {
		return err
	}

	// we call this because the owner was changed
	if err := k.AfterDymNameOwnerChanged(ctx, newDymNameRecord.Name); err != nil {
		return err
	}

	// we call this because the config was cleared
	if err := k.AfterDymNameConfigChanged(ctx, newDymNameRecord.Name); err != nil {
		return err
	}

	return nil
}
