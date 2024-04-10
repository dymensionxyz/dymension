package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tenderminttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) TriggerGenesisEvent(goCtx context.Context, msg *types.MsgRollappGenesisEvent) (*types.MsgRollappGenesisEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the sender and validate they are in the whitelist
	if whitelist := k.DeployerWhitelist(ctx); len(whitelist) > 0 {
		if !k.IsAddressInDeployerWhiteList(ctx, msg.Address) {
			return nil, sdkerrors.ErrUnauthorized
		}
	}

	// Get the rollapp
	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	// Get the channel and validate it's connected client chain is the same as the rollapp's
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", msg.ChannelId)
	if err != nil {
		return nil, err
	}
	tmClientState, ok := clientState.(*tenderminttypes.ClientState)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrInvalidGenesisChannelId, "expected tendermint client state, got %T", clientState)
	}
	if tmClientState.GetChainID() != msg.RollappId {
		return nil, sdkerrors.Wrapf(types.ErrInvalidGenesisChannelId, "channel %s is connected to chain ID %s, expected %s",
			msg.ChannelId, tmClientState.GetChainID(), msg.RollappId)
	}

	// Update the rollapp with the channelID and trigger the genesis event
	rollapp.ChannelId = msg.ChannelId
	if err = k.TriggerRollappGenesisEvent(ctx, rollapp); err != nil {
		return nil, err
	}

	return &types.MsgRollappGenesisEventResponse{}, nil
}
