package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"

	"github.com/dymensionxyz/dymension/v3/app/params"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// deprecateCrisisModule sets the constant fee to an unreachable value to effectively disable the crisis module.
// The crisis module was used to handle the invariant checks and halt the chain if an invariant was broken.
func deprecateCrisisModule(ctx sdk.Context, keeper *crisiskeeper.Keeper) error {
	const unreachableFee = 1_000_000_000 // 1B DYM
	return keeper.SetConstantFee(ctx, sdk.NewCoin(params.BaseDenom, commontypes.DYM.MulRaw(unreachableFee)))
}
