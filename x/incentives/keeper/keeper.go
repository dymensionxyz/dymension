package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// Keeper provides a way to manage incentives module storage.
type Keeper struct {
	storeKey  storetypes.Key
	cdc       codec.BinaryCodec
	hooks     types.IncentiveHooks
	bk        types.BankKeeper
	lk        types.LockupKeeper
	ek        types.EpochKeeper
	tk        types.TxFeesKeeper
	rk        types.RollappKeeper
	sk        types.SequencerKeeper
	spk       types.SponsorshipKeeper
	authority string
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(
	storeKey storetypes.Key,
	cdc codec.BinaryCodec,
	bk types.BankKeeper,
	lk types.LockupKeeper,
	ek types.EpochKeeper,
	txfk types.TxFeesKeeper,
	rk types.RollappKeeper,
	sk types.SequencerKeeper,
	spk types.SponsorshipKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		bk:        bk,
		lk:        lk,
		ek:        ek,
		tk:        txfk,
		rk:        rk,
		sk:        sk,
		spk:       spk,
		authority: authority,
	}
}

// SetHooks sets the incentives hooks.
func (k *Keeper) SetHooks(ih types.IncentiveHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set incentive hooks twice")
	}

	k.hooks = ih

	return k
}

// Logger returns a logger instance for the incentives module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetLockableDurations sets which lockable durations will be incentivized.
func (k Keeper) SetLockableDurations(ctx sdk.Context, lockableDurations []time.Duration) {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{LockableDurations: lockableDurations}
	osmoutils.MustSet(store, types.LockableDurationsKey, &info)
}

// GetLockableDurations returns all incentivized lockable durations.
func (k Keeper) GetLockableDurations(ctx sdk.Context) []time.Duration {
	store := ctx.KVStore(k.storeKey)
	info := types.LockableDurationsInfo{}
	osmoutils.MustGet(store, types.LockableDurationsKey, &info)
	return info.LockableDurations
}

// GetEpochInfo returns EpochInfo struct given context.
func (k Keeper) GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo {
	params := k.GetParams(ctx)
	return k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)
}
