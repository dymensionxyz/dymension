package keeper

import (
	"bytes"
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// ValidateRollappId checks that the rollapp id from the ibc connection matches the rollapp, checking the sequencer registered with the consensus state validator set
func (k *Keeper) ValidateRollappId(ctx types.Context, rollappID, rollappPortOnHub string, rollappChannelOnHub string) error {
	// Get the sequencer from the latest state info update and check the validator set hash
	// from the headers match with the sequencer for the rollappID
	// As the assumption the sequencer is honest we don't check the packet proof height.
	latestStateIndex, found := k.rollappKeeper.GetLatestStateInfoIndex(ctx, rollappID)
	if !found {
		return errors.Wrapf(rollapptypes.ErrUnknownRollappID, "state index not found for the rollappID: %s", rollappID)
	}
	stateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollappID, latestStateIndex.Index)
	if !found {
		return errors.Wrapf(rollapptypes.ErrUnknownRollappID, "state info not found for the rollappID: %s with index: %d", rollappID, latestStateIndex.Index)
	}
	// Compare the validators set hash of the consensus state to the sequencer hash.
	// TODO (srene): We compare the validator set of the last consensus height, because it fails to  get consensus for a different height,
	// but we should compare the validator set at the height of the last state info, because sequencer may have changed after that.
	// If the sequencer is changed, then the validation will fail till the new sequencer sends a new state info update.
	tmConsensusState, err := k.getTmConsensusState(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		k.Logger(ctx).Error("error consensus state", err)
		return err
	}

	// Gets sequencer information from the sequencer address found in the latest state info
	sequencer, found := k.sequencerKeeper.GetSequencer(ctx, stateInfo.Sequencer)
	if !found {
		return errors.Wrapf(sequencertypes.ErrUnknownSequencer, "sequencer %s not found for the rollappID %s", stateInfo.Sequencer, rollappID)
	}

	// Gets the validator set hash made out of the pub key for the sequencer
	seqPubKeyHash, err := sequencer.GetDymintPubKeyHash()
	if err != nil {
		return err
	}

	// It compares the validator set hash from the consensus state with the one we recreated from the sequencer. If its true it means the chain corresponds to the rollappID chain
	if !bytes.Equal(tmConsensusState.NextValidatorsHash, seqPubKeyHash) {
		errMsg := fmt.Sprintf("consensus state does not match: consensus state validators %x, rollappID sequencer %x",
			tmConsensusState.NextValidatorsHash, stateInfo.Sequencer)
		return errors.Wrap(delayedacktypes.ErrMismatchedSequencer, errMsg)
	}
	return nil
}

// getTmConsensusState returns the tendermint consensus state for the channel for specific height
func (k Keeper) getTmConsensusState(ctx types.Context, portID string, channelID string) (*ibctmtypes.ConsensusState, error) {
	// Get the client state for the channel for specific height
	connectionEnd, err := k.getConnectionEnd(ctx, portID, channelID)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, err
	}
	clientState, err := k.GetClientState(ctx, portID, channelID)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, err
	}

	// TODO(srene) : consensus state is only obtained when getting it for latestheight. this can be an issue when sequencer changes. i have to figure out why is only returned for latest height

	consensusState, found := k.clientKeeper.GetClientConsensusState(ctx, connectionEnd.GetClientID(), clientState.GetLatestHeight())
	if !found {
		return nil, clienttypes.ErrConsensusStateNotFound
	}
	tmConsensusState, ok := consensusState.(*ibctmtypes.ConsensusState)
	if !ok {
		return nil, errors.Wrapf(delayedacktypes.ErrInvalidType, "expected tendermint consensus state, got %T", consensusState)
	}
	return tmConsensusState, nil
}

func (k Keeper) getConnectionEnd(ctx types.Context, portID string, channelID string) (conntypes.ConnectionEnd, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return conntypes.ConnectionEnd{}, errors.Wrap(channeltypes.ErrChannelNotFound, channelID)
	}
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])

	if !found {
		return conntypes.ConnectionEnd{}, errors.Wrap(conntypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}
	return connectionEnd, nil
}
