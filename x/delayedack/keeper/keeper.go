package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type Keeper struct {
	rollapptypes.StubRollappCreatedHooks

	cdc                   codec.Codec
	storeKey              storetypes.Key
	channelKeeperStoreKey storetypes.Key // we need direct access to the IBC channel store
	hooks                 types.MultiDelayedAckHooks
	authority             string

	// pendingPacketsByAddress is an index of all pending packets associated with a Hub address.
	// In case of ON_RECV packet (Rollapp -> Hub), the address is the packet receiver.
	// In case of ON_ACK/ON_TIMEOUT packet (Hub -> Rollapp), the address is the packet sender.
	// Index key: receiver address + packet key.
	pendingPacketsByAddress collections.KeySet[collections.Pair[string, []byte]]

	rollappKeeper types.RollappKeeper
	porttypes.ICS4Wrapper
	channelKeeper types.ChannelKeeper
	types.EIBCKeeper

	// TODO: refac https://github.com/dymensionxyz/dymension/issues/1849
	completionHooks map[string]CompletionHookInstance
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.Key,
	channelKeeperStoreKey storetypes.Key,
	authority string,
	rollappKeeper types.RollappKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	eibcKeeper types.EIBCKeeper,
) *Keeper {
	return &Keeper{
		cdc:                   cdc,
		storeKey:              storeKey,
		channelKeeperStoreKey: channelKeeperStoreKey,
		authority:             authority,
		pendingPacketsByAddress: collections.NewKeySet(
			collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey)),
			collections.NewPrefix(types.PendingPacketsByAddressKeyPrefix),
			"pending_packets_by_receiver",
			collections.PairKeyCodec(collections.StringKey, collcodec.NewBytesKey[[]byte]()),
		),
		rollappKeeper:   rollappKeeper,
		ICS4Wrapper:     ics4Wrapper,
		channelKeeper:   channelKeeper,
		EIBCKeeper:      eibcKeeper,
		completionHooks: make(map[string]CompletionHookInstance),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// expose codec to be used by the delayedack middleware
func (k Keeper) Cdc() codec.Codec {
	return k.cdc
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
