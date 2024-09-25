package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	hooks      types.MultiRollappHooks
	paramstore paramtypes.Subspace
	authority  string // authority is the x/gov module account

	ibcClientKeeper       types.IBCClientKeeper
	canonicalClientKeeper types.CanonicalLightClientKeeper
	channelKeeper         types.ChannelKeeper
	sequencerKeeper       types.SequencerKeeper
	bankKeeper            types.BankKeeper

	vulnerableDRSVersions collections.KeySet[string]

	finalizePending func(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	channelKeeper types.ChannelKeeper,
	ibcclientKeeper types.IBCClientKeeper,
	sequencerKeeper types.SequencerKeeper,
	bankKeeper types.BankKeeper,
	authority string,
	canonicalClientKeeper types.CanonicalLightClientKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid x/rollapp authority address: %w", err))
	}

	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		paramstore:      ps,
		hooks:           nil,
		channelKeeper:   channelKeeper,
		authority:       authority,
		ibcClientKeeper: ibcclientKeeper,
		sequencerKeeper: sequencerKeeper,
		bankKeeper:      bankKeeper,
		vulnerableDRSVersions: collections.NewKeySet(
			collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey)),
			collections.NewPrefix(types.VulnerableDRSVersionsKeyPrefix),
			"vulnerable_drs_versions",
			collections.StringKey,
		),
		finalizePending:       nil,
		canonicalClientKeeper: canonicalClientKeeper,
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

func (k *Keeper) SetSequencerKeeper(sk types.SequencerKeeper) {
	k.sequencerKeeper = sk
}

func (k *Keeper) SetCanonicalClientKeeper(kk types.CanonicalLightClientKeeper) {
	k.canonicalClientKeeper = kk
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

func (k *Keeper) SetHooks(sh types.MultiRollappHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiRollappHooks {
	return k.hooks
}
