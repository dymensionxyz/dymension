package lockup

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
)

// EndBlocker is called every block to automatically unlock matured locks.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	// disable automatic withdraw before specific block height
	// it is actually for testing with legacy
	MinBlockHeightToBeginAutoWithdrawing := int64(6)
	if ctx.BlockHeight() < MinBlockHeightToBeginAutoWithdrawing {
		return []abci.ValidatorUpdate{}
	}

	// withdraw and delete locks
	k.WithdrawAllMaturedLocks(ctx)
	return []abci.ValidatorUpdate{}
}
