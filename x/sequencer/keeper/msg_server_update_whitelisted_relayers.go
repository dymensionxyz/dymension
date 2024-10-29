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

// UpdateWhitelistedRelayers defines a method for updating the sequencer's whitelisted relater list.
func (k msgServer) UpdateWhitelistedRelayers(goCtx context.Context, msg *types.MsgUpdateWhitelistedRelayers) (*types.MsgUpdateWhitelistedRelayersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, ok := k.GetSequencer(ctx, msg.Creator)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrNotFound, "sequencer")
	}

	seq.SetWhitelistedRelayers(msg.Relayers)

	k.SetSequencer(ctx, seq)

	err := uevent.EmitTypedEvent(ctx, &types.EventUpdateWhitelistedRelayers{
		Creator:  msg.Creator,
		Relayers: msg.Relayers,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateWhitelistedRelayersResponse{}, nil
}
