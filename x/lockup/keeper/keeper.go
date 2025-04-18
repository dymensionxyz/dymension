package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	stroretypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// Keeper provides a way to manage module storage.
type Keeper struct {
	storeKey stroretypes.StoreKey

	hooks types.LockupHooks

	paramSpace paramtypes.Subspace

	ak types.AccountKeeper
	bk types.BankKeeper
	tk types.TxFeesKeeper
}

// NewKeeper returns an instance of Keeper.
func NewKeeper(storeKey stroretypes.StoreKey, paramSpace paramtypes.Subspace, ak types.AccountKeeper, bk types.BankKeeper, tk types.TxFeesKeeper) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		bk:         bk,
		tk:         tk,
	}
}

// GetParams returns the total set of lockup parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams sets the total set of lockup parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
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

// AdminKeeper defines a god privilege keeper functions to remove tokens from locks and create new locks
// For the governance system of token pools, we want a "ragequit" feature
// So governance changes will take 1 week to go into effect
// During that time, people can choose to "ragequit" which means they would leave the original pool
// and form a new pool with the old parameters but if they still had 2 months of lockup left,
// their liquidity still needs to be 2 month lockup-ed, just in the new pool
// And we need to replace their pool1 LP tokens with pool2 LP tokens with the same lock duration and end time.
type AdminKeeper struct {
	Keeper
}
