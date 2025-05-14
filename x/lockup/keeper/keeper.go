package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	stroretypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// Keeper provides a way to manage module storage.
type Keeper struct {
	storeKey stroretypes.StoreKey
	cdc      codec.BinaryCodec
	hooks    types.LockupHooks

	ak types.AccountKeeper
	bk types.BankKeeper
	tk types.TxFeesKeeper

	authority string
}

// NewKeeper returns an instance of Keeper.
func NewKeeper(storeKey stroretypes.StoreKey, cdc codec.BinaryCodec, ak types.AccountKeeper, bk types.BankKeeper, tk types.TxFeesKeeper, authority string) *Keeper {

	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		ak:        ak,
		bk:        bk,
		tk:        tk,
		authority: authority,
	}
}

// GetParams returns the total set of lockup parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.KeyParams)
	k.cdc.MustUnmarshal(b, &params)
	return params
}

// SetParams sets the total set of lockup parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.KeyParams, b)
}

func (k Keeper) GetForceUnlockAllowedAddresses(ctx sdk.Context) (forceUnlockAllowedAddresses []string) {
	return k.GetParams(ctx).ForceUnlockAllowedAddresses
}

func (k Keeper) GetLockCreationFee(ctx sdk.Context) math.Int {
	return k.GetParams(ctx).LockCreationFee
}

// Logger returns a logger instance.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Set the lockup hooks.
func (k *Keeper) SetHooks(lh types.LockupHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set lockup hooks twice")
	}

	k.hooks = lh

	return k
}
