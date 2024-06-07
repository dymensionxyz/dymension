package keeper

import (
	"bytes"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/dymension/v3/utils/gerr"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetValidTransferDataWithFinalizationInfo(
	ctx sdk.Context,
	packet channeltypes.Packet,
	packetType commontypes.RollappPacket_Type,
) (data types.TransferDataWithFinalization, err error) {
	rollappPortOnHub, rollappChannelOnHub := packet.DestinationPort, packet.DestinationChannel

	transferData, err := k.GetValidTransferData(ctx, packet)
	if err != nil {
		err = errorsmod.Wrap(err, "get valid transfer data")
	}
	data.TransferData = transferData

	packetId := commontypes.NewPacketUID(packetType, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	height, ok := types.PacketProofHeightFromCtx(ctx, packetId)
	if !ok {
		// TODO: should probably be a panic
		err = errorsmod.Wrapf(gerr.ErrNotFound, "get proof height from context: packetID: %s", packetId)
		return
	}
	data.ProofHeight = height.RevisionHeight

	finalizedHeight, err := k.getRollappFinalizedHeight(ctx, data.RollappID)
	if err != nil && !errorsmod.IsOf(err, rollapptypes.ErrNoFinalizedStateYetForRollapp) {
		err = errorsmod.Wrap(err, "get rollapp finalized height")
		return
	}
	data.Finalized = err == nil && finalizedHeight >= data.ProofHeight
	return
}

func (k Keeper) GetValidTransferData(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (data types.TransferData, err error) {
	rollappPortOnHub, rollappChannelOnHub := packet.DestinationPort, packet.DestinationChannel

	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = errorsmod.Wrap(err, "unmarshal transfer data")
		return
	}
	if err := data.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(err, "validate basic")
		return
	}
	chainID, err := k.chainIDFromPortChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		err = errorsmod.Wrap(err, "chain id from port and channel")
		return
	}
	rollapp, ok := k.GetRollapp(ctx, chainID)
	if !ok {
		// no problem, it corresponds to a regular non-rollapp chain
		return
	}
	data.RollappID = chainID
	if rollapp.ChannelId == "" {
		err = errorsmod.Wrapf(rollapptypes.ErrGenesisEventNotTriggered, "empty channel id: rollap id: %s", chainID)
		return
	}
	if rollapp.ChannelId != rollappChannelOnHub {
		err = errorsmod.Wrapf(
			rollapptypes.ErrMismatchedChannelID,
			"channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, rollappChannelOnHub,
		)
		return
	}

	err = k.validateRollappID(ctx, data.RollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		err = errorsmod.Wrap(err, "validate rollapp id")
		return
	}

	return
}

func (k Keeper) chainIDFromPortChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
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

// validateRollappID checks that the rollapp id from the ibc connection matches the rollapp, checking the sequencer registered with the consensus state validator set
func (k Keeper) validateRollappID(ctx sdk.Context, raID, rollappPortOnHub string, rollappChannelOnHub string) error {
	// Compare the validators set hash of the consensus state to the sequencer hash.
	// TODO (srene): We compare the validator set of the last consensus height, because it fails to  get consensus for a different height,
	// but we should compare the validator set at the height of the last state info, because sequencer may have changed after that.
	// If the sequencer is changed, then the validation will fail till the new sequencer sends a new state info update.
	nextValidatorsHash, err := k.getNextValidatorsHash(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return errorsmod.Wrap(err, "get next validators hash")
	}

	// Get the sequencer from the latest state info update and check the validator set hash
	// from the headers match with the sequencer for the raID
	// As the assumption the sequencer is honest we don't check the packet proof height.
	sequencerID, sequencerPubKeyHash, err := k.getLatestSequencerPubKey(ctx, raID)
	if err != nil {
		return errorsmod.Wrap(err, "get latest sequencer pub key")
	}

	// It compares the validator set hash from the consensus state with the one we recreated from the sequencer. If its true it means the chain corresponds to the raID chain
	if !bytes.Equal(nextValidatorsHash, sequencerPubKeyHash) {
		return errorsmod.Wrapf(
			gerr.ErrUnauthenticated,
			"consensus state does not match: consensus state validators: %x, raID sequencer: %x", nextValidatorsHash, sequencerID)
	}

	return nil
}

// getLatestSequencerPubKey returns the *hash* of the pub key of the latest validator
func (k Keeper) getLatestSequencerPubKey(ctx sdk.Context, rollappID string) (string, []byte, error) {
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
func (k Keeper) getNextValidatorsHash(ctx sdk.Context, portID string, channelID string) ([]byte, error) {
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

func (k Keeper) getConnectionEnd(ctx sdk.Context, portID string, channelID string) (conntypes.ConnectionEnd, error) {
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
