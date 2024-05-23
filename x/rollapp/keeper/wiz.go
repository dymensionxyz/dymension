package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	tenderminttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) TriggerGen(goCtx context.Context, msg *types.MsgRollappGenesisEvent) (*types.MsgRollappGenesisEventResponse, error) {
	/*
		What does this do?

		Get the client for the channel, and check the tendermint chain id matches the rollapp id
		Sets the rollapp channel id
	*/

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	// NOTE: whitelist check removed here

	// Get the rollapp
	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	// Get the channel and validate it's connected client chain is the same as the rollapp's
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", msg.ChannelId)
	if err != nil {
		return nil, fmt.Errorf("get channel client state: %w", err)
	}
	tmClientState, ok := clientState.(*tenderminttypes.ClientState)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidGenesisChannelId, "expected tendermint client state, got %T", clientState)
	}
	if tmClientState.GetChainID() != msg.RollappId {
		return nil, errors.Wrapf(types.ErrInvalidGenesisChannelId, "channel connected to wrong chain: channel: %s: got: %s: expect: %s",
			msg.ChannelId, tmClientState.GetChainID(), msg.RollappId)
	}

	// Update the rollapp with the channelID and trigger the genesis event
	rollapp.ChannelId = msg.ChannelId
	if err = k.TriggerRollappGenesisEvent(ctx, rollapp); err != nil {
		return nil, fmt.Errorf("trigger rollapp genesis event: %w", err)
	}

	return &types.MsgRollappGenesisEventResponse{}, nil
}
