package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/tendermint/tendermint/libs/log"
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

// ExtractRollappIDAndTransferPacket extracts the rollapp ID from the packet
func (k Keeper) ExtractRollappIDAndTransferPacket(ctx sdk.Context, packet channeltypes.Packet, rollappPortOnHub string, rollappChannelOnHub string) (string, *transfertypes.FungibleTokenPacketData, error) {
	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return "", nil, err
	}
	// Check if the packet is destined for a rollapp
	chainID, err := k.ExtractChainIDFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return "", &data, err
	}
	rollapp, found := k.GetRollapp(ctx, chainID)
	if !found {
		return "", &data, nil
	}
	if rollapp.ChannelId == "" {
		return "", &data, errorsmod.Wrapf(rollapptypes.ErrGenesisEventNotTriggered, "empty channel id: rollap id: %s", chainID)
	}
	// check if the channelID matches the rollappID's channelID
	if rollapp.ChannelId != rollappChannelOnHub {
		return "", &data, errorsmod.Wrapf(
			rollapptypes.ErrMismatchedChannelID,
			"channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, rollappChannelOnHub,
		)
	}

	return chainID, &data, nil
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
	connectionEnd, err := k.getConnectionEnd(ctx, portID, channelID)
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

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}
