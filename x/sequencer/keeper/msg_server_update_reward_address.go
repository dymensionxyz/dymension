package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpdateRewardAddress defines a method for updating the sequencer's reward address.
func (k msgServer) UpdateRewardAddress(goCtx context.Context, msg *types.MsgUpdateRewardAddress) (*types.MsgUpdateRewardAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	seq.RewardAddr = msg.RewardAddr

	err = uevent.EmitTypedEvent(ctx, &types.EventUpdateRewardAddress{
		Creator:    msg.Creator,
		RewardAddr: msg.RewardAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRewardAddressResponse{}, nil
}
