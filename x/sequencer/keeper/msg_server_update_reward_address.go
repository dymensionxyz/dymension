package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpdateRewardAddress defines a method for updating the sequencer's reward address.
func (k msgServer) UpdateRewardAddress(goCtx context.Context, msg *types.MsgUpdateRewardAddress) (*types.MsgUpdateRewardAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, ok := k.GetSequencer(ctx, msg.Creator)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrNotFound, "sequencer")
	}

	seq.RewardAddr = msg.RewardAddr

	k.SetSequencer(ctx, seq)

	err := uevent.EmitTypedEvent(ctx, &types.EventUpdateRewardAddress{
		Creator:    msg.Creator,
		RewardAddr: msg.RewardAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRewardAddressResponse{}, nil
}
