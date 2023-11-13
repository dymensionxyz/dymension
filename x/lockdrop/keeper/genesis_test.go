package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/dymensionxyz/dymension/app"

	"github.com/dymensionxyz/dymension/x/lockdrop"
	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

var (
	now         = time.Now().UTC()
	testGenesis = types.GenesisState{
		Params: types.Params{
			MintedDenom: "udym",
		},
		DistrInfo: &types.DistrInfo{
			TotalWeight: sdk.NewInt(1),
			Records: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(1),
				},
			},
		},
	}
)

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	encodingConfig := simapp.MakeEncodingConfig()
	appCodec := encodingConfig.Codec
	am := lockdrop.NewAppModule(*app.LockdropKeeper)

	genesis := testGenesis
	app.LockdropKeeper.InitGenesis(ctx, &genesis)
	assert.Equal(t, app.LockdropKeeper.GetDistrInfo(ctx), *testGenesis.DistrInfo)

	genesisExported := am.ExportGenesis(ctx, appCodec)
	assert.NotPanics(t, func() {
		app := simapp.Setup(t, false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := lockdrop.NewAppModule(*app.LockdropKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func TestInitGenesis(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.LockdropKeeper.InitGenesis(ctx, &genesis)

	params := app.LockdropKeeper.GetParams(ctx)
	require.Equal(t, params, genesis.Params)

	distrInfo := app.LockdropKeeper.GetDistrInfo(ctx)
	require.Equal(t, distrInfo, *genesis.DistrInfo)
}

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.App.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	suite.App.LockdropKeeper.InitGenesis(ctx, &genesis)

	genesisExported := suite.App.LockdropKeeper.ExportGenesis(ctx)
	suite.Equal(genesisExported.Params, genesis.Params)
	suite.Equal(genesisExported.DistrInfo, genesis.DistrInfo)
}
