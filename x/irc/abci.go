package irc

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dymensionxyz/dymension/x/irc/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k *keeper.Keeper) {
}

// Called every block to finalize states that their dispute period over.
func EndBlocker(ctx sdk.Context, k *keeper.Keeper) {
}
