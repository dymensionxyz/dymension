package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// finalizationQueueIndex is a set of indexes for the finalization queue.
type finalizationQueueIndex struct {
	// RollappIDReverseLookup is a reverse lookup index for the finalization queue.
	// It helps to find all available heights to finalize by rollapp.
	RollappIDReverseLookup *indexes.ReversePair[uint64, string, types.BlockHeightToFinalizationQueue]
}

func (b finalizationQueueIndex) IndexesList() []collections.Index[collections.Pair[uint64, string], types.BlockHeightToFinalizationQueue] {
	return []collections.Index[collections.Pair[uint64, string], types.BlockHeightToFinalizationQueue]{b.RollappIDReverseLookup}
}

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	hooks      types.MultiRollappHooks
	paramstore paramtypes.Subspace
	authority  string // authority is the x/gov module account

	canonicalClientKeeper CanonicalLightClientKeeper
	channelKeeper         ChannelKeeper
	sequencerKeeper       SequencerKeeper
	bankKeeper            BankKeeper
	transferKeeper        TransferKeeper

	obsoleteDRSVersions     collections.KeySet[uint32]
	registeredRollappDenoms collections.KeySet[collections.Pair[string, string]]
	// finalizationQueue is a map from creation height and rollapp to the finalization queue.
	// Key: (creation height, rollappID), Value: state indexes to finalize.
	// Contains a special index that helps reverse lookup: finalization queue (all available heights) by rollapp.
	// Index key: (rollappID, creation height), Value: state indexes to finalize.
	finalizationQueue *collections.IndexedMap[collections.Pair[uint64, string], types.BlockHeightToFinalizationQueue, finalizationQueueIndex]

	finalizePending        func(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error
	seqToUnfinalizedHeight collections.KeySet[collections.Pair[string, uint64]]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	channelKeeper ChannelKeeper,
	sequencerKeeper SequencerKeeper,
	bankKeeper BankKeeper,
	transferKeeper TransferKeeper,
	authority string,
	canonicalClientKeeper CanonicalLightClientKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid x/rollapp authority address: %w", err))
	}

	service := collcompat.NewKVStoreService(storeKey)
	sb := collections.NewSchemaBuilder(service)

	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		paramstore:      ps,
		hooks:           nil,
		channelKeeper:   channelKeeper,
		authority:       authority,
		sequencerKeeper: sequencerKeeper,
		bankKeeper:      bankKeeper,
		transferKeeper:  transferKeeper,
		obsoleteDRSVersions: collections.NewKeySet(
			sb,
			collections.NewPrefix(types.ObsoleteDRSVersionsKeyPrefix),
			"obsolete_drs_versions",
			collections.Uint32Key,
		),
		registeredRollappDenoms: collections.NewKeySet[collections.Pair[string, string]](
			sb,
			collections.NewPrefix(types.KeyRegisteredDenomPrefix),
			"registered_rollapp_denoms",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		),
		finalizationQueue: collections.NewIndexedMap(
			sb,
			collections.NewPrefix(types.HeightRollappToFinalizationQueueKeyPrefix),
			"height_rollapp_to_finalization_queue",
			collections.PairKeyCodec(collections.Uint64Key, collections.StringKey),
			collcompat.ProtoValue[types.BlockHeightToFinalizationQueue](cdc),
			finalizationQueueIndex{
				RollappIDReverseLookup: indexes.NewReversePair[types.BlockHeightToFinalizationQueue](
					sb,
					collections.NewPrefix(types.RollappHeightToFinalizationQueueKeyPrefix),
					"rollapp_id_reverse_lookup",
					collections.PairKeyCodec(collections.Uint64Key, collections.StringKey),
				),
			},
		),
		finalizePending:       nil,
		canonicalClientKeeper: canonicalClientKeeper,
		seqToUnfinalizedHeight: collections.NewKeySet(
			sb,
			types.SeqToUnfinalizedHeightKeyPrefix,
			"seq_to_unfinalized_height",
			collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
		),
	}
	k.SetFinalizePendingFn(k.finalizePendingState)
	return k
}

func (k *Keeper) SetFinalizePendingFn(fn func(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error) {
	k.finalizePending = fn
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetSequencerKeeper(sk SequencerKeeper) {
	k.sequencerKeeper = sk
}

func (k *Keeper) SetCanonicalClientKeeper(kk CanonicalLightClientKeeper) {
	k.canonicalClientKeeper = kk
}

func (k *Keeper) SetTransferKeeper(transferKeeper TransferKeeper) {
	k.transferKeeper = transferKeeper
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

func (k *Keeper) SetHooks(sh types.MultiRollappHooks) {
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiRollappHooks {
	return k.hooks
}
