package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/ibc-go/v6/modules/core/exported"
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
		hooks      types.MultiDelayedAckHooks
		paramstore paramtypes.Subspace

		rollappKeeper types.RollappKeeper
		porttypes.ICS4Wrapper
		channelKeeper types.ChannelKeeper
		types.EIBCKeeper
		bankKeeper types.BankKeeper
	}
)

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, ps paramtypes.Subspace, rollappKeeper types.RollappKeeper, ics4Wrapper porttypes.ICS4Wrapper, channelKeeper types.ChannelKeeper, eibcKeeper types.EIBCKeeper, bankKeeper types.BankKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramstore:    ps,
		rollappKeeper: rollappKeeper,
		ICS4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		bankKeeper:    bankKeeper,
		EIBCKeeper:    eibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) IsRollappsEnabled(ctx sdk.Context) bool {
	return k.rollappKeeper.GetParams(ctx).RollappsEnabled
}

func (k Keeper) getRollappFinalizedHeight(ctx sdk.Context, chainID string) (uint64, error) {
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

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}
