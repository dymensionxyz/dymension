package keeper

import (
	"bytes"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	tenderminttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		hooks      types.MultiDelayedAckHooks
		paramstore paramtypes.Subspace

		rollappKeeper   types.RollappKeeper
		sequencerKeeper types.SequencerKeeper
		porttypes.ICS4Wrapper
		channelKeeper    types.ChannelKeeper
		connectionKeeper types.ConnectionKeeper
		clientKeeper     types.ClientKeeper
		types.EIBCKeeper
		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	rollappKeeper types.RollappKeeper,
	sequencerKeeper types.SequencerKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	connectionKeeper types.ConnectionKeeper,
	clientKeeper types.ClientKeeper,
	eibcKeeper types.EIBCKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		paramstore:       ps,
		rollappKeeper:    rollappKeeper,
		sequencerKeeper:  sequencerKeeper,
		ICS4Wrapper:      ics4Wrapper,
		channelKeeper:    channelKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		bankKeeper:       bankKeeper,
		EIBCKeeper:       eibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ExtractChainIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", fmt.Errorf("failed to extract clientID from channel: %w", err)
	}

	tmClientState, ok := clientState.(*ibctypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmClientState.ChainId, nil
}

func (k Keeper) IsRollappsEnabled(ctx sdk.Context) bool {
	return k.rollappKeeper.GetParams(ctx).RollappsEnabled
}

func (k Keeper) GetRollapp(ctx sdk.Context, chainID string) (rollapptypes.Rollapp, bool) {
	return k.rollappKeeper.GetRollapp(ctx, chainID)
}

func (k Keeper) GetRollappFinalizedHeight(ctx sdk.Context, chainID string) (uint64, error) {
	// GetLatestFinalizedStateIndex
	latestFinalizedStateIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, chainID)
	if !found {
		return 0, rollapptypes.ErrNoFinalizedStateYetForRollapp
	}

	stateInfo := k.rollappKeeper.MustGetStateInfo(ctx, chainID, latestFinalizedStateIndex.Index)
	return stateInfo.StartHeight + stateInfo.NumBlocks - 1, nil
}

// GetClientState retrieves the client state for a given packet.
func (k Keeper) GetClientState(ctx sdk.Context, portID string, channelID string) (exported.ClientState, error) {
	connectionEnd, err := k.GetConnectionEnd(ctx, portID, channelID)
	if err != nil {
		return nil, err
	}
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.GetClientID())
	if !found {
		return nil, clienttypes.ErrConsensusStateNotFound
	}

	return clientState, nil
}

func (k Keeper) BlockedAddr(addr string) bool {
	account, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}
	return k.bankKeeper.BlockedAddr(account)
}

/* -------------------------------------------------------------------------- */
/*                               Hooks handling                               */
/* -------------------------------------------------------------------------- */

func (k *Keeper) SetHooks(hooks types.MultiDelayedAckHooks) {
	if k.hooks != nil {
		panic("DelayedAckHooks already set")
	}
	k.hooks = hooks
}

func (k *Keeper) GetHooks() types.MultiDelayedAckHooks {
	return k.hooks
}

/* -------------------------------------------------------------------------- */
/*                                 ICS4Wrapper                                */
/* -------------------------------------------------------------------------- */

// ExtractRollappAndTransferPacketFromData extracts the rollapp and fungible token from the packet data
func (k *Keeper) ExtractRollappAndTransferPacketFromData(
	ctx sdk.Context,
	data []byte,
	rollappPortOnHub string,
	rollappChannelOnHub string,
) (*rollapptypes.Rollapp, *transfertypes.FungibleTokenPacketData, error) {
	// no-op if the packet is not a fungible token packet
	var altPacket types.WrappedFungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(data, &altPacket); err != nil {
		// try if it's a FungibleTokenPacketData packet
		altPacket.FungibleTokenPacketData = new(transfertypes.FungibleTokenPacketData)
		if err = types.ModuleCdc.UnmarshalJSON(data, altPacket.FungibleTokenPacketData); err != nil {
			return nil, nil, fmt.Errorf("cannot unmarshal transfer packet data: %w", err)
		}
	}
	rollapp, err := k.ExtractRollappFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return nil, nil, err
	}

	return rollapp, altPacket.FungibleTokenPacketData, nil
}

// ExtractRollappFromChannel extracts the rollapp from the IBC port and channel
func (k *Keeper) ExtractRollappFromChannel(
	ctx sdk.Context,
	rollappPortOnHub string,
	rollappChannelOnHub string,
) (*rollapptypes.Rollapp, error) {
	// Check if the packet is destined for a rollapp
	chainID, err := k.ExtractChainIDFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return nil, err
	}
	rollapp, found := k.GetRollapp(ctx, chainID)
	if !found {
		return nil, nil
	}
	if rollapp.ChannelId == "" {
		return nil, errorsmod.Wrapf(rollapptypes.ErrGenesisEventNotTriggered, "empty channel id: rollap id: %s", chainID)
	}
	// check if the channelID matches the rollappID's channelID
	if rollapp.ChannelId != rollappChannelOnHub {
		return nil, errorsmod.Wrapf(
			rollapptypes.ErrMismatchedChannelID,
			"channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, rollappChannelOnHub,
		)
	}

	return &rollapp, nil
}

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}

// ValidateRollappId checks that the rollapp id from the ibc connection matches the rollapp, checking the sequencer registered with the consensus state validator set
func (k *Keeper) ValidateRollappId(ctx sdk.Context, rollappID, rollappPortOnHub string, rollappChannelOnHub string) error {
	// Get the sequencer from the latest state info update and check the validator set hash
	// from the headers match with the sequencer for the rollappID
	// As the assumption the sequencer is honest we don't check the packet proof height.
	latestStateIndex, found := k.rollappKeeper.GetLatestStateInfoIndex(ctx, rollappID)
	if !found {
		return errorsmod.Wrapf(rollapptypes.ErrUnknownRollappID, "state index not found for the rollappID: %s", rollappID)
	}
	stateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollappID, latestStateIndex.Index)
	if !found {
		return errorsmod.Wrapf(rollapptypes.ErrUnknownRollappID, "state info not found for the rollappID: %s with index: %d", rollappID, latestStateIndex.Index)
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
		return errorsmod.Wrapf(sequencertypes.ErrUnknownSequencer, "sequencer %s not found for the rollappID %s", stateInfo.Sequencer, rollappID)
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
		return errorsmod.Wrap(types.ErrMismatchedSequencer, errMsg)
	}
	return nil
}

func (k Keeper) GetConnectionEnd(ctx sdk.Context, portID string, channelID string) (connectiontypes.ConnectionEnd, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return connectiontypes.ConnectionEnd{}, errorsmod.Wrap(channeltypes.ErrChannelNotFound, channelID)
	}
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])

	if !found {
		return connectiontypes.ConnectionEnd{}, errorsmod.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}
	return connectionEnd, nil
}

// getTmConsensusState returns the tendermint consensus state for the channel for specific height
func (k Keeper) getTmConsensusState(ctx sdk.Context, portID string, channelID string) (*tenderminttypes.ConsensusState, error) {
	// Get the client state for the channel for specific height
	connectionEnd, err := k.GetConnectionEnd(ctx, portID, channelID)
	if err != nil {
		return &tenderminttypes.ConsensusState{}, err
	}
	clientState, err := k.GetClientState(ctx, portID, channelID)
	if err != nil {
		return &tenderminttypes.ConsensusState{}, err
	}

	// TODO(srene) : consensus state is only obtained when getting it for latestheight. this can be an issue when sequencer changes. i have to figure out why is only returned for latest height

	consensusState, found := k.clientKeeper.GetClientConsensusState(ctx, connectionEnd.GetClientID(), clientState.GetLatestHeight())
	if !found {
		return nil, clienttypes.ErrConsensusStateNotFound
	}
	tmConsensusState, ok := consensusState.(*tenderminttypes.ConsensusState)
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrInvalidType, "expected tendermint consensus state, got %T", consensusState)
	}
	return tmConsensusState, nil
}
