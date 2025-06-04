package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "foo", Func: InvariantFoo},
}

// RegisterInvariants registers the module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantFoo(k Keeper) uinv.Func { // TODO: impl some
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		return nil
	})
}
