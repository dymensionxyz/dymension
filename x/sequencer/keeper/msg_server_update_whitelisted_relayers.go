package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpdateWhitelistedRelayers defines a method for updating the sequencer's whitelisted relater list.
func (k msgServer) UpdateWhitelistedRelayers(goCtx context.Context, msg *types.MsgUpdateWhitelistedRelayers) (*types.MsgUpdateWhitelistedRelayersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	seq.SetWhitelistedRelayers(msg.Relayers)

	err = uevent.EmitTypedEvent(ctx, &types.EventUpdateWhitelistedRelayers{
		Creator:  msg.Creator,
		Relayers: msg.Relayers,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateWhitelistedRelayersResponse{}, nil
}
