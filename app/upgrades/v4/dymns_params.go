package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func migrateDymnsParams(ctx sdk.Context, dk dymnskeeper.Keeper) error {
	params := dymnstypes.DefaultParams()
	return dk.SetParams(ctx, params)
}
