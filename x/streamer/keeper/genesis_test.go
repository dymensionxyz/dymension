package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// TestStreamerExportGenesis tests export genesis command for the streamer module.
func TestStreamerExportGenesis(t *testing.T) {
	// export genesis using default configurations
	// ensure resulting genesis params match default params
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cometbftproto.Header{})
	genesis := app.StreamerKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params, types.DefaultGenesis().Params)
	require.Len(t, genesis.Streams, 0)

	// fund the module
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	err := bankutil.FundModuleAccount(app.BankKeeper, ctx, types.ModuleName, coins)
	require.NoError(t, err)

	_, err = app.IncentivesKeeper.CreateGauge(
		ctx,
		true,
		app.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByTime,
			Denom:         "stake",
			Duration:      time.Hour,
			Timestamp:     time.Time{},
		}, time.Now(), 1)
	require.NoError(t, err)

	// create a stream that distributes coins to earlier created LP token and duration
	startTime := time.Now()
	distr := []types.DistrRecord{
		{
			GaugeId: 1,
			Weight:  math.NewInt(50),
		},
	}
	streamID, err := app.StreamerKeeper.CreateStream(ctx, coins, distr, startTime, "day", 30, NonSponsored)
	require.NoError(t, err)

	// export genesis using default configurations
	// ensure resulting genesis params match default params
	genesis = app.StreamerKeeper.ExportGenesis(ctx)
	require.Len(t, genesis.Streams, 1)

	distInfo, err := types.NewDistrInfo(distr)
	require.NoError(t, err)

	// ensure the first stream listed in the exported genesis explicitly matches expectation
	const numEpochsPaidOver = 30
	require.Equal(t, genesis.Streams[0], types.Stream{
		Id:                   streamID,
		DistributeTo:         distInfo,
		Coins:                coins,
		StartTime:            startTime.UTC(),
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    numEpochsPaidOver,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins(nil),
		Sponsored:            false,
		EpochCoins:           coins.QuoInt(math.NewInt(numEpochsPaidOver)),
	})
}

// TestStreamerInitGenesis takes a genesis state and tests initializing that genesis for the streamer module.
func TestStreamerInitGenesis(t *testing.T) {
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cometbftproto.Header{})

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)

	// create coins, lp tokens with lockup durations, and a stream for this lockup
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()

	distr := types.DistrInfo{
		TotalWeight: math.NewInt(100),
		Records: []types.DistrRecord{
			{
				GaugeId: 1,
				Weight:  math.NewInt(50),
			},
			{
				GaugeId: 2,
				Weight:  math.NewInt(50),
			},
		},
	}

	stream := types.Stream{
		Id:                   1,
		DistributeTo:         &distr,
		Coins:                coins,
		NumEpochsPaidOver:    2,
		DistrEpochIdentifier: "day",
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins(nil),
		StartTime:            startTime.UTC(),
	}

	// initialize genesis with specified parameter, the stream created earlier, and lockable durations
	expectedPointer := types.EpochPointer{
		StreamId:        1,
		GaugeId:         1,
		EpochIdentifier: "day",
	}
	app.StreamerKeeper.InitGenesis(ctx, types.GenesisState{
		Params:        types.Params{},
		Streams:       []types.Stream{stream},
		LastStreamId:  1,
		EpochPointers: []types.EpochPointer{expectedPointer},
	})

	// check that the stream created earlier was initialized through initGenesis and still exists on chain
	streams := app.StreamerKeeper.GetStreams(ctx)
	lastStreamID := app.StreamerKeeper.GetLastStreamID(ctx)
	require.Len(t, streams, 1)
	require.Equal(t, streams[0], stream)
	require.Equal(t, lastStreamID, uint64(1))
	ep, err := app.StreamerKeeper.GetEpochPointer(ctx, "day")
	require.NoError(t, err)
	require.Equal(t, expectedPointer, ep)
}

func TestStreamerOrder(t *testing.T) {
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, cometbftproto.Header{}).WithBlockTime(time.Now())

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)
	// create coins, lp tokens with lockup durations, and a stream for this lockup
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10000)}
	startTime := time.Now()

	distr := types.DistrInfo{
		TotalWeight: math.NewInt(100),
		Records: []types.DistrRecord{
			{
				GaugeId: 1,
				Weight:  math.NewInt(50),
			},
			{
				GaugeId: 2,
				Weight:  math.NewInt(50),
			},
		},
	}

	stream := types.Stream{
		Id:                   1,
		DistributeTo:         &distr,
		Coins:                coins,
		NumEpochsPaidOver:    2,
		DistrEpochIdentifier: "day",
		FilledEpochs:         2,
		DistributedCoins:     sdk.Coins(nil),
		// stream starts in 10 seconds
		StartTime: startTime.Add(time.Second * 10).UTC(),
	}

	stream2 := types.Stream{
		Id:                   2,
		DistributeTo:         &distr,
		Coins:                coins,
		NumEpochsPaidOver:    2,
		DistrEpochIdentifier: "day",
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins(nil),
		// stream starts in 1 day
		StartTime: startTime.Add(time.Second * 60 * 60 * 24 * 1).UTC(),
	}

	stream3 := types.Stream{
		Id:                   3,
		DistributeTo:         &distr,
		Coins:                coins,
		NumEpochsPaidOver:    2,
		DistrEpochIdentifier: "day",
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins(nil),
		// stream starts in 2 days
		StartTime: startTime.Add(time.Second * 60 * 60 * 24 * 2).UTC(),
	}

	app.StreamerKeeper.InitGenesis(ctx, types.GenesisState{
		Params:       types.Params{},
		Streams:      []types.Stream{stream, stream2, stream3},
		LastStreamId: 3,
	})
	// check that the stream created earlier was initialized through initGenesis and still exists on chain
	streams := app.StreamerKeeper.GetStreams(ctx)
	lastStreamID := app.StreamerKeeper.GetLastStreamID(ctx)
	require.Len(t, streams, 3)
	require.Equal(t, lastStreamID, streams[len(streams)-1].Id, "last stream id invariant broken")

	// Forward block time by 1 minute to start the stream with id 1
	ctx = ctx.WithBlockTime(startTime.Add(time.Second * 60))
	// Move stream with id 1 from upcoming to active
	err := app.StreamerKeeper.MoveUpcomingStreamToActiveStream(ctx, stream)
	require.NoError(t, err)

	streams = app.StreamerKeeper.GetStreams(ctx)
	lastStreamID = app.StreamerKeeper.GetLastStreamID(ctx)
	require.Len(t, streams, 3)
	require.Equal(t, lastStreamID, streams[len(streams)-1].Id, "last stream id invariant broken")
}
