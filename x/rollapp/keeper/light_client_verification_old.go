package keeper

import (
	"bytes"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectionkeeper "github.com/cosmos/ibc-go/v6/modules/core/03-connection/keeper"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	conntypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channelkeeper "github.com/cosmos/ibc-go/v6/modules/core/04-channel/keeper"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type ClientKeeper interface {
	GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool)
}

type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type LCV struct {
	Keeper
	delayedAckKeeper *delayedackkeeper.Keeper
	sequencerKeeper  *sequencerkeeper.Keeper
	channelKeeper    ChannelKeeper
	connectionKeeper ConnectionKeeper
	clientKeeper     ClientKeeper
}

func NewLCV(
	k Keeper,
	delayedAckKeeper *delayedackkeeper.Keeper,
	sequencerKeeper *sequencerkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
) *LCV {
	return &LCV{
		Keeper:           k,
		delayedAckKeeper: delayedAckKeeper,
		sequencerKeeper:  sequencerKeeper,
		channelKeeper:    &channelKeeper,
		connectionKeeper: connectionKeeper,
		clientKeeper:     clientKeeper,
	}
}

func (k LCV) chainIDFromPortChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, state, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", errorsmod.Wrap(err, "get channel client state")
	}

	tmState, ok := state.(*ibctmtypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmState.ChainId, nil
}

// ensureIBCClientLatestNextValidatorsHashMatchesCurrentSequencer checks that the current sequencer' pub key hash for the rollapp
// actually matches the nextValidators hash in the ibc light client for the rollapp
//
// ASSUMPTIONS:
//
//	sequencer is fixed, see todo 1
//	sequencer is valid, see todo 2
func (k LCV) ensureIBCClientLatestNextValidatorsHashMatchesCurrentSequencer(ctx sdk.Context, raID, rollappPortOnHub string, rollappChannelOnHub string) error {
	/*
		TODO: Support sequencer changes: we use the latest nextValidators hash, but really we should check the validator set at the light
			client header corresponding to the last (finalized?) state info, because the sequencer may have changed after that.
			Currently, if the sequencer were to change suddenly, the light client may change before the state info is updated, and this
			would resolve to invalid.
	*/
	lightClientNextValidatorsHash, err := k.getNextValidatorsHash(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return errorsmod.Wrap(err, "get next validators hash")
	}

	/*
			TODO: Support trustless sequencer.
				Ask Sergi for help, also see https://github.com/dymensionxyz/research/issues/258#issue-2152850199
					"""
					This solution is with the following assumptions:
						1. The sequencer is honest.
						2. The sequencer doesn't change (i.e unbond).
						In that case we can take the matching consensus state for the packet height and validate
		 				the signature pubKey against the current active sequencer on the hub (assuming sequencer
						hasn't changed). Assuming the sequencer is honest it won't create a malicious chain with
						invalid headers and at least a 3rd party can't pretend to be the rollapp chain.
					"""
				Sergi quote:
					"""
					Get the sequencer from the latest state info update and check the validator set hash
					from the headers match with the sequencer for the raID
					As the assumption the sequencer is honest we don't check the packet proof height.
					"""
	*/
	sequencerID, latestSequencerPubKeyHash, err := k.getLatestSequencerPubKey(ctx, raID)
	if err != nil {
		return errorsmod.Wrap(err, "get latest sequencer pub key")
	}

	if !bytes.Equal(lightClientNextValidatorsHash, latestSequencerPubKeyHash) {
		return errorsmod.Wrapf(
			gerrc.ErrUnauthenticated,
			"consensus state does not match: consensus state validators: %x, raID sequencer: %x", lightClientNextValidatorsHash, sequencerID)
	}

	return nil
}

// getLatestSequencerPubKey returns the *hash* of the pub key of the latest validator
func (k LCV) getLatestSequencerPubKey(ctx sdk.Context, rollappID string) (string, []byte, error) {
	state, found := k.GetLatestStateInfo(ctx, rollappID)
	if !found {
		return "", nil, gerrc.ErrNotFound
	}

	sequencerID := state.GetSequencer()
	sequencer, ok := k.sequencerKeeper.GetSequencer(ctx, sequencerID)
	if !ok {
		return "", nil, errorsmod.Wrapf(gerrc.ErrNotFound, "sequencer: id: %s", sequencerID)
	}
	seqPubKeyHash, err := sequencer.GetDymintPubKeyHash()
	if err != nil {
		return "", nil, errorsmod.Wrap(err, "get dymint pubkey hash")
	}
	return sequencerID, seqPubKeyHash, nil
}

// getNextValidatorsHash returns the tendermint consensus state next validators hash for the latest client height associated to the channel
func (k LCV) getNextValidatorsHash(ctx sdk.Context, portID string, channelID string) ([]byte, error) {
	conn, err := k.getConnectionEnd(ctx, portID, channelID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get connection end")
	}
	client, err := k.delayedAckKeeper.GetClientState(ctx, portID, channelID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get client state")
	}

	// TODO: see todos in ensureIBCClientLatestNextValidatorsHashMatchesCurrentSequencer for discussion of latest height
	consensusState, ok := k.clientKeeper.GetClientConsensusState(ctx, conn.GetClientID(), client.GetLatestHeight())
	if !ok {
		return nil, errors.Join(gerrc.ErrNotFound, clienttypes.ErrConsensusStateNotFound)
	}
	tmConsensusState, ok := consensusState.(*ibctmtypes.ConsensusState)
	if !ok {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "expected tendermint consensus state, got: %T", consensusState)
	}
	return tmConsensusState.NextValidatorsHash, nil
}

func (k LCV) getConnectionEnd(ctx sdk.Context, portID string, channelID string) (conntypes.ConnectionEnd, error) {
	ch, ok := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !ok {
		return conntypes.ConnectionEnd{}, errorsmod.Wrap(errors.Join(gerrc.ErrNotFound, channeltypes.ErrChannelNotFound), channelID)
	}
	conn, ok := k.connectionKeeper.GetConnection(ctx, ch.ConnectionHops[0])
	if !ok {
		return conntypes.ConnectionEnd{}, errorsmod.Wrap(errors.Join(gerrc.ErrNotFound, conntypes.ErrConnectionNotFound), ch.ConnectionHops[0])
	}
	return conn, nil
}
