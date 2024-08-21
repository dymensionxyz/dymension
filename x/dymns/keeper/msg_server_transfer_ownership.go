package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// TransferDymNameOwnership is message handler,
// handles transfer of ownership of a Dym-Name, performed by the owner.
func (k msgServer) TransferDymNameOwnership(goCtx context.Context, msg *dymnstypes.MsgTransferDymNameOwnership) (*dymnstypes.MsgTransferDymNameOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateTransferDymNameOwnership(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err := k.transferDymNameOwnership(ctx, *dymName, msg.NewOwner); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgTransferDymNameOwnershipResponse{}, nil
}

// validateTransferDymNameOwnership handles validation for message handled by TransferDymNameOwnership
func (k msgServer) validateTransferDymNameOwnership(ctx sdk.Context, msg *dymnstypes.MsgTransferDymNameOwnership) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "Dym-Name is already expired")
	}

	so := k.GetSellOrder(ctx, msg.Name, dymnstypes.TypeName)
	if so != nil {
		// by ignoring SO, can fall into case that SO not completed/lost funds of bidder,...

		return nil, errorsmod.Wrap(
			gerrc.ErrFailedPrecondition,
			"can not transfer ownership while there is an active Sell Order",
		)
	}

	return dymName, nil
}

// transferDymNameOwnership transfers ownership of a Dym-Name to a new owner.
func (k Keeper) transferDymNameOwnership(ctx sdk.Context, dymName dymnstypes.DymName, newOwner string) error {
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
