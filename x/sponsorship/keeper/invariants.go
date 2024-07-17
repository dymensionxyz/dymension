package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, "streams-count", GaugeWeightsInvariant(keeper))
}

func GaugeWeightsInvariant(Keeper) sdk.Invariant {
	// TODO
	return func(ctx sdk.Context) (string, bool) {
		var broken bool
		var msg string

		return sdk.FormatInvariant(types.ModuleName, "streamer-balance", msg), broken
	}
}
