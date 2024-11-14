package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{k: k}
}

// OnHardFork implements the RollappHooks interface
func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappID string, _ uint64) error {
	return hook.k.rk.ClearRegisteredDenoms(ctx, rollappID)
}
