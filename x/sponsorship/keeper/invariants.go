package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/invar"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var invs = invar.NamedFuncsList[Keeper]{
	{"delegator-validator-power", InvariantDelegatorValidatorPower},
	{"distribution", InvariantDistribution},
	{"votes", InvariantVotes},
	{"general", InvariantGeneral},
}

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantDelegatorValidatorPower(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
	}
}

func InvariantDistribution(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
	}
}
func InvariantVotes(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
	}
}
func InvariantGeneral(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
	}
}
