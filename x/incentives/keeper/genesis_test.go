package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// TestIncentivesExportGenesis tests export genesis command for the incentives module.
func TestIncentivesExportGenesis(t *testing.T) {
	// export genesis using default configurations
	// ensure resulting genesis params match default params
	app := apptesting.Setup(t)
	ctx := app.NewContext(false)
	genesis := app.IncentivesKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Gauges, 0)

	// create an address and fund with coins
	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := bankutil.FundAccount(ctx, app.BankKeeper, addr, coins)
	require.NoError(t, err)

	// mints LP tokens and send to address created earlier
	// this ensures the supply exists on chain
	distrTo := lockuptypes.QueryCondition{
		Denom:    "lptoken",
		Duration: time.Second,
	}
	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = bankutil.FundAccount(ctx, app.BankKeeper, addr, mintLPtokens)
	require.NoError(t, err)

	// create a gauge that distributes coins to earlier created LP token and duration
	startTime := time.Now()
	gaugeID, err := app.IncentivesKeeper.CreateAssetGauge(ctx, true, addr, coins, distrTo, startTime, 1)
	require.NoError(t, err)

	// export genesis using default configurations
	// ensure resulting genesis params match default params
	genesis = app.IncentivesKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Gauges, 1)

	// ensure the first gauge listed in the exported genesis explicitly matches expectation
	require.Equal(t, genesis.Gauges[0], types.Gauge{
		Id:                gaugeID,
		IsPerpetual:       true,
		DistributeTo:      &types.Gauge_Asset{Asset: &distrTo},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	})
}

// TestIncentivesInitGenesis takes a genesis state and tests initializing that genesis for the incentives module.
func TestIncentivesInitGenesis(t *testing.T) {
	app := apptesting.Setup(t)
	ctx := app.NewContext(false)

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.ValidateBasic()
	require.NoError(t, validateGenesis)

	// create coins, lp tokens with lockup durations, and a gauge for this lockup
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()
	distrTo := lockuptypes.QueryCondition{
		Denom:    "lptoken",
		Duration: time.Second,
	}
	gauge := types.Gauge{
		Id:                1,
		IsPerpetual:       false,
		DistributeTo:      &types.Gauge_Asset{Asset: &distrTo},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	}

	// initialize genesis with specified parameter, the gauge created earlier, and lockable durations
	app.IncentivesKeeper.InitGenesis(ctx, types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "week",
			CreateGaugeBaseFee:   math.ZeroInt(),
			AddToGaugeBaseFee:    math.ZeroInt(),
			AddDenomFee:          math.ZeroInt(),
		},
		Gauges: []types.Gauge{gauge},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	})

	// check that the gauge created earlier was initialized through initGenesis and still exists on chain
	gauges := app.IncentivesKeeper.GetGauges(ctx)
	require.Len(t, gauges, 1)
	require.Equal(t, gauges[0], gauge)
}
