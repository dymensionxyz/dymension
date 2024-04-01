package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	// Validate all other genesis fields are empty
	genesisFields := []int{
		len(genState.RollappPackets),
	}
	for _, fieldLength := range genesisFields {
		if fieldLength != 0 {
			panic("Only params can be initialized at genesis")
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:         k.GetParams(ctx),
		RollappPackets: k.GetAllRollappPackets(ctx),
	}
}
