package irc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/irc/keeper"
	"github.com/dymensionxyz/dymension/x/irc/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	// Set all the ircRequest
	for _, elem := range genState.IRCRequestList {
		k.SetIRCRequest(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.IRCRequestList = k.GetAllIRCRequest(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
