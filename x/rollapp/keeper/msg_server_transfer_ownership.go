package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) TransferOwnership(goCtx context.Context, msg *types.MsgTransferOwnership) (*types.MsgTransferOwnershipResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidRequest, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, ok := k.GetRollapp(ctx, msg.RollappId)
	if !ok {
		return nil, types.ErrUnknownRollappID
	}

	if rollapp.Owner != msg.CurrentOwner {
		return nil, types.ErrUnauthorizedSigner
	}

	if rollapp.Frozen {
		return nil, types.ErrRollappFrozen
	}

	if rollapp.Owner == msg.NewOwner {
		return nil, types.ErrSameOwner
	}

	rollapp.Owner = msg.NewOwner
	k.SetRollapp(ctx, rollapp)

	if err := uevent.EmitTypedEvent(ctx, msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgTransferOwnershipResponse{}, nil
}
