package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func DymNSKeeper(t *testing.T) (dymnskeeper.Keeper, dymnstypes.BankKeeper, rollappkeeper.Keeper, sdk.Context) {
	app := apptesting.Setup(t)

	k := app.DymNSKeeper
	bk := app.BankKeeper
	rk := app.RollappKeeper

	ctx := app.NewContext(false)
	// Initialize params
	moduleParams := dymnstypes.DefaultParams()
	moduleParams.Chains.AliasesOfChainIds = nil
	err := k.SetParams(ctx, moduleParams)
	require.NoError(t, err)

	return k, bk, *rk, ctx
}
