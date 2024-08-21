package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/app"
	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

func TestInitExportGenesis(t *testing.T) {
	const (
		appID1 = "app1"
		appID2 = "app2"
	)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		AppList: []types.App{
			{
				Name: appID1,
			},
			{
				Name: appID2,
			},
		},
	}

	k, ctx := keepertest.AppKeeper(t)
	app.InitGenesis(ctx, *k, genesisState)
	got := app.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(genesisState)
	nullify.Fill(*got)

	require.ElementsMatch(t, genesisState.AppList, got.AppList)
}
