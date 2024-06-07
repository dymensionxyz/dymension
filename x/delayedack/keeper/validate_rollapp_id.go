package keeper

import (
	"bytes"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/types"
)

// ValidateRollappId checks that the rollapp id from the ibc connection matches the rollapp, checking the sequencer registered with the consensus state validator set
func (k Keeper) ValidateRollappId(ctx types.Context, rollappID, rollappPortOnHub string, rollappChannelOnHub string) error {
	// Compare the validators set hash of the consensus state to the sequencer hash.
	// TODO (srene): We compare the validator set of the last consensus height, because it fails to  get consensus for a different height,
	// but we should compare the validator set at the height of the last state info, because sequencer may have changed after that.
	// If the sequencer is changed, then the validation will fail till the new sequencer sends a new state info update.
	nextValidatorsHash, err := k.getNextValidatorsHash(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return errorsmod.Wrap(err, "get next validators hash")
	}

	// Get the sequencer from the latest state info update and check the validator set hash
	// from the headers match with the sequencer for the rollappID
	// As the assumption the sequencer is honest we don't check the packet proof height.
	sequencerID, sequencerPubKeyHash, err := k.getLatestSequencerPubKey(ctx, rollappID)
	if err != nil {
		return errorsmod.Wrap(err, "get latest sequencer pub key")
	}

	// It compares the validator set hash from the consensus state with the one we recreated from the sequencer. If its true it means the chain corresponds to the rollappID chain
	if !bytes.Equal(nextValidatorsHash, sequencerPubKeyHash) {
		return errorsmod.Wrapf(
			gerr.ErrUnauthenticated,
			"consensus state does not match: consensus state validators: %x, rollappID sequencer: %x", nextValidatorsHash, sequencerID)
	}
	return nil
}

// getLatestSequencerPubKey returns the *hash* of the pub key of the latest validator
func (k Keeper) getLatestSequencerPubKey(ctx types.Context, rollappID string) (string, []byte, error) {
	state, ok := k.rollappKeeper.GetLatestStateInfo(ctx, rollappID)
	if !ok {
		return "", nil, errorsmod.Wrap(gerr.ErrNotFound, "latest state info")
	}
	sequencerID := state.GetSequencer()
	sequencer, ok := k.sequencerKeeper.GetSequencer(ctx, sequencerID)
	if !ok {
		return "", nil, errorsmod.Wrapf(gerr.ErrNotFound, "sequencer: id: %s", sequencerID)
	}
	seqPubKeyHash, err := sequencer.GetDymintPubKeyHash()
	if err != nil {
		return "", nil, errorsmod.Wrap(err, "get dymint pubkey hash")
	}
	return sequencerID, seqPubKeyHash, nil
}

// getNextValidatorsHash returns the tendermint consensus state next validators hash for the latest client height associated to the channel
func (k Keeper) getNextValidatorsHash(ctx types.Context, portID string, channelID string) ([]byte, error) {
	conn, err := k.getConnectionEnd(ctx, portID, channelID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get connection end")
	}
	client, err := k.GetClientState(ctx, portID, channelID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get client state")
	}

	// TODO(srene) : consensus state is only obtained when getting it for latestheight.
	// 	this can be an issue when sequencer changes. i have to figure out why is only returned for latest height

	/*
		TODO:
			Person to ask: srene
			It fails if now passing the latest client height
			If the sequencer changes, we


	*/

	consensusState, ok := k.clientKeeper.GetClientConsensusState(ctx, conn.GetClientID(), client.GetLatestHeight())
	if !ok {
		return nil, clienttypes.ErrConsensusStateNotFound
	}
	tmConsensusState, ok := consensusState.(*ibctmtypes.ConsensusState)
	if !ok {
		return nil, errorsmod.Wrapf(gerr.ErrInvalidArgument, "expected tendermint consensus state, got: %T", consensusState)
	}
	return tmConsensusState.NextValidatorsHash, nil
}

func (k Keeper) getConnectionEnd(ctx types.Context, portID string, channelID string) (conntypes.ConnectionEnd, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return conntypes.ConnectionEnd{}, errorsmod.Wrap(channeltypes.ErrChannelNotFound, channelID)
	}
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])

	if !found {
		return conntypes.ConnectionEnd{}, errorsmod.Wrap(conntypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}
	return connectionEnd, nil
}
