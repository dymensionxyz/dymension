package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// intended to be called by relayer, but can be called by anyone
// verifies that the suggested client is safe to designate canonical and matches state updates from the sequencer
func (m msgServer) SetCanonicalClient(goCtx context.Context, msg *types.MsgSetCanonicalClient) (*types.MsgSetCanonicalClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	clientStateI, ok := m.ibcClientKeeper.GetClientState(ctx, msg.ClientId)
	if !ok {
		return nil, gerrc.ErrNotFound.Wrap("client")
	}

	clientState, ok := clientStateI.(*ibctm.ClientState)
	if !ok {
		return nil, gerrc.ErrInvalidArgument.Wrap("not tm client")
	}

	chainID := clientState.ChainId
	_, ok = m.rollappKeeper.GetRollapp(ctx, chainID)
	if !ok {
		return nil, gerrc.ErrNotFound.Wrap("rollapp")
	}
	rollappID := chainID

	_, ok = m.GetCanonicalClient(ctx, rollappID)
	if ok {
		return nil, gerrc.ErrAlreadyExists.Wrap("canonical client for rollapp")
	}

	latestHeight, ok := m.rollappKeeper.GetLatestHeight(ctx, rollappID)
	if !ok {
		return nil, gerrc.ErrNotFound.Wrap("latest rollapp height")
	}

	err := m.validClient(ctx, msg.ClientId, clientState, rollappID, latestHeight)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unsafe to mark client canonical: check that sequencer has posted a recent state update")
	}

	m.Keeper.SetCanonicalClient(ctx, rollappID, msg.ClientId)

	if err := uevent.EmitTypedEvent(ctx, &types.EventSetCanonicalClient{
		RollappId: rollappID,
		ClientId:  msg.ClientId,
	}); err != nil {
		return nil, errorsmod.Wrap(err, "emit typed event")
	}

	return &types.MsgSetCanonicalClientResponse{}, nil
}
