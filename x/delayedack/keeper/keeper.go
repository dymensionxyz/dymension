package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	tenderminttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		hooks      types.MultiDelayedAckHooks
		paramstore paramtypes.Subspace

		rollappKeeper    types.RollappKeeper
		sequencerKeeper  types.SequencerKeeper
		ics4Wrapper      porttypes.ICS4Wrapper
		channelKeeper    types.ChannelKeeper
		connectionKeeper types.ConnectionKeeper
		clientKeeper     types.ClientKeeper
		types.EIBCKeeper
		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
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
		memKey:           memKey,
		paramstore:       ps,
		rollappKeeper:    rollappKeeper,
		sequencerKeeper:  sequencerKeeper,
		ics4Wrapper:      ics4Wrapper,
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
	res, err := k.rollappKeeper.StateInfo(ctx, &rollapptypes.QueryGetStateInfoRequest{
		RollappId: chainID,
		Finalized: true,
	})
	if err != nil {
		return 0, err
	}

	return (res.StateInfo.StartHeight + res.StateInfo.NumBlocks - 1), nil
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

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (k *Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, acknowledgement exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper GetAppVersion function.
func (k *Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}

// VallidateRollappId checks that the rollappid from the ibc connection matches the rollapp checking the sequencer registered with the consensus state validator set
func (k *Keeper) ValidateRollappId(ctx sdk.Context, rollapp, portID, channelID string) error {

	// Get the sequencer from the latest state info update and check the validator set hash
	// from the headers match with the sequencer for the rollapp
	// As the assumption the sequencer is honest we don't check the packet proof height.
	latestStateIndex, found := k.rollappKeeper.GetLatestStateInfoIndex(ctx, rollapp)
	if !found {
		return sdkerrors.Wrapf(rollapptypes.ErrUnknownRollappID, "state index not found for the rollapp: %s", rollapp)
	}
	stateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollapp, latestStateIndex.Index)
	if !found {
		return sdkerrors.Wrapf(rollapptypes.ErrUnknownRollappID, "state info not found for the rollapp: %s with index: %d", rollapp, latestStateIndex.Index)
	}

	// Compare the validators set hash of the consensus state to the sequencer hash.
	// TODO (srene): We compare the validator set of the last consensus height, because it fails to  get consensus for a different height,
	// but we should compare the validator set at the height of the last state info, because sequencer may have changed after that.
	// If the sequencer is changed, then the validation will faill till the new sequencer sends a new state info update.
	tmConsensusState, err := k.getTmConsensusState(ctx, portID, channelID)
	if err != nil {
		k.Logger(ctx).Error("error consensus state", err)
		return err
	}

	//Gets sequencer information from the sequencer address found in the latest state info
	sequencer, found := k.sequencerKeeper.GetSequencer(ctx, stateInfo.Sequencer)
	if !found {
		return sdkerrors.Wrapf(sequencertypes.ErrUnknownSequencer, "sequencer %s not found for the rollapp %s", stateInfo.Sequencer, rollapp)
	}

	//Gets the validator set hash made out of the pub key for the sequencer
	seqPubKeyHash, err := sequencer.GetDymintPubKeyHash()
	if err != nil {
		return err
	}

	//It compares the validator set hash from the consensus state with the one we recreated from the sequencer. If its true it means the chain corresponds to the rollapp chain
	if !bytes.Equal(tmConsensusState.NextValidatorsHash, seqPubKeyHash) {
		errMsg := fmt.Sprintf("consensus state does not match: consensus state validators %x, rollapp sequencer %x",
			tmConsensusState.NextValidatorsHash, stateInfo.Sequencer)
		return sdkerrors.Wrap(types.ErrMismatchedSequencer, errMsg)
	}
	return nil
}

func (k Keeper) GetConnectionEnd(ctx sdk.Context, portID string, channelID string) (connectiontypes.ConnectionEnd, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return connectiontypes.ConnectionEnd{}, sdkerrors.Wrap(channeltypes.ErrChannelNotFound, channelID)
	}
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])

	if !found {
		return connectiontypes.ConnectionEnd{}, sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
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

	//TODO(srene) : consensus state is only obtained when getting it for latestheight. this can be an issue when sequencer changes. i have to figure out why is only returned for latest height

	consensusState, found := k.clientKeeper.GetClientConsensusState(ctx, connectionEnd.GetClientID(), clientState.GetLatestHeight())
	if !found {
		return nil, clienttypes.ErrConsensusStateNotFound
	}
	tmConsensusState, ok := consensusState.(*tenderminttypes.ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected tendermint consensus state, got %T", consensusState)
	}
	return tmConsensusState, nil
}
