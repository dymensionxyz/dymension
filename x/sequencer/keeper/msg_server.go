package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) sequencertypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ sequencertypes.MsgServer = msgServer{}

// PunishSequencer implements the Msg/PunishSequencer RPC method
func (k msgServer) PunishSequencer(goCtx context.Context, msg *sequencertypes.MsgPunishSequencer) (*sequencertypes.MsgPunishSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	rewardee, _ := sdk.AccAddressFromBech32(msg.Rewardee)

	if err := k.Keeper.PunishSequencer(ctx, msg.PunishSequencerAddress, &rewardee); err != nil {
		return nil, err
	}

	return &sequencertypes.MsgPunishSequencerResponse{}, nil
}
