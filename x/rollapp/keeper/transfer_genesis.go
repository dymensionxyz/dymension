package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// VerifyAndRecordGenesisTransfer TODO: could just pass the rollapp object?
func (k Keeper) VerifyAndRecordGenesisTransfer(ctx sdk.Context, rollappID string, ix int, n int) error {
	ra := k.MustGetRollapp(ctx, rollappID)

	return nil
}

func (k Keeper) GetAllGenesisTransfers(ctx sdk.Context) []types.GenesisTransfers {
	var ret []types.GenesisTransfers
	return ret
}
