package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/dymensionxyz/dymension/app"

	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/dymensionxyz/dymension/x/streamer/types"
)

// TestStreamerExportGenesis tests export genesis command for the streamer module.
func TestStreamerExportGenesis(t *testing.T) {
	// export genesis using default configurations
	// ensure resulting genesis params match default params
	app := osmoapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	genesis := app.StreamerKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params, types.DefaultGenesis().Params)
	require.Len(t, genesis.Streams, 0)

	// create an address and fund with coins
	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := bankutil.FundAccount(app.BankKeeper, ctx, addr, coins)
	require.NoError(t, err)

	mintLPtokens := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	err = bankutil.FundAccount(app.BankKeeper, ctx, addr, mintLPtokens)
	require.NoError(t, err)

	// create a stream that distributes coins to earlier created LP token and duration
	startTime := time.Now()
	streamID, err := app.StreamerKeeper.CreateStream(ctx, coins, distrTo, startTime, "day", 1)
	require.NoError(t, err)

	// export genesis using default configurations
	// ensure resulting genesis params match default params
	genesis = app.StreamerKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params.DistrEpochIdentifier, "week")
	require.Len(t, genesis.Streams, 1)

	// ensure the first stream listed in the exported genesis explicitly matches expectation
	require.Equal(t, genesis.Streams[0], types.Stream{
		Id:                streamID,
		IsPerpetual:       true,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	})
}

// TestStreamerInitGenesis takes a genesis state and tests initializing that genesis for the streamer module.
func TestStreamerInitGenesis(t *testing.T) {
	app := osmoapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)

	// create coins, lp tokens with lockup durations, and a stream for this lockup
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()

	stream := types.Stream{
		Id:                1,
		DistributeTo:      distrTo,
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins(nil),
		StartTime:         startTime.UTC(),
	}

	// initialize genesis with specified parameter, the stream created earlier, and lockable durations
	app.StreamerKeeper.InitGenesis(ctx, types.GenesisState{
		Params: types.Params{
			DistrEpochIdentifier: "week",
		},
		Streams: []types.Stream{stream},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	})

	// check that the stream created earlier was initialized through initGenesis and still exists on chain
	streams := app.StreamerKeeper.GetStreams(ctx)
	require.Len(t, streams, 1)
	require.Equal(t, streams[0], stream)
}
