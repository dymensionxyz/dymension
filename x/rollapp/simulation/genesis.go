package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {

	p := types.DefaultParams()
	p.MinSequencerBondGlobal = commontypes.ADym(sdk.NewInt(100))
	g := types.GenesisState{
		Params: p,
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&g)
}
