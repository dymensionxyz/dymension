package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	defaultLPDenom           string        = "lptoken"
	defaultLPSyntheticDenom  string        = "lptoken/superbonding"
	defaultLPTokens          sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPDenom, 10)}
	defaultLPSyntheticTokens sdk.Coins     = sdk.Coins{sdk.NewInt64Coin(defaultLPSyntheticDenom, 10)}
	defaultLiquidTokens      sdk.Coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	defaultLockDuration      time.Duration = time.Second
	oneLockupUser            userLocks     = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPTokens},
	}
	twoLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
	oneSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{time.Second},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens},
	}
	twoSyntheticLockupUser userLocks = userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPSyntheticTokens, defaultLPSyntheticTokens},
	}
	defaultRewardDenom string = "rewardDenom"
)

// TODO: Switch more code to use userLocks and perpStreamDesc
// TODO: Create issue for the above.
type userLocks struct {
	lockDurations []time.Duration
	lockAmounts   []sdk.Coins
}

type perpStreamDesc struct {
	lockDenom    string
	lockDuration time.Duration
	rewardAmount sdk.Coins
}

// SetupStreams takes an array of perpStreamDesc structs. Then returns the corresponding array of Stream structs.
func (suite *KeeperTestSuite) SetupStreams(streamDescriptors []perpStreamDesc, denom string) []types.Stream {
	streams := make([]types.Stream, len(streamDescriptors))
	for i, desc := range streamDescriptors {
		_, streamPtr, _, _ := suite.setupNewStreamWithDuration(desc.rewardAmount, denom)
		streams[i] = *streamPtr
	}
	return streams
}

// CreateStream creates a stream struct given the required params.
func (suite *KeeperTestSuite) CreateStream(distrTo sdk.AccAddress, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	suite.FundAcc(suite.moduleAddress, coins)
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

// setupNewStreamWithDuration creates a stream with the specified duration.
func (suite *KeeperTestSuite) setupNewStreamWithDuration(coins sdk.Coins, duration time.Duration, denom string) (
	uint64, *types.Stream, sdk.Coins, time.Time,
) {
	addr := sdk.AccAddress([]byte("Stream_Creation_Addr_"))
	startTime2 := time.Now()

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	suite.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	streamID, stream := suite.CreateStream(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return streamID, stream, coins, startTime2
}

// SetupNewStream creates a stream with the default lock duration.
func (suite *KeeperTestSuite) SetupNewStream(isPerpetual bool, coins sdk.Coins) (uint64, *types.Stream, sdk.Coins, time.Time) {
	return suite.setupNewStreamWithDuration(isPerpetual, coins, defaultLockDuration, "lptoken")
}

// setupNewStreamWithDenom creates a stream with the specified duration and denom.
func (suite *KeeperTestSuite) setupNewStreamWithDenom(isPerpetual bool, coins sdk.Coins, duration time.Duration, denom string) (
	uint64, *types.Stream, sdk.Coins, time.Time,
) {
	addr := sdk.AccAddress([]byte("Stream_Creation_Addr_"))
	startTime2 := time.Now()
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         denom,
		Duration:      duration,
	}

	// mints coins so supply exists on chain
	mintCoins := sdk.Coins{sdk.NewInt64Coin(distrTo.Denom, 200)}
	suite.FundAcc(addr, mintCoins)

	numEpochsPaidOver := uint64(2)
	if isPerpetual {
		numEpochsPaidOver = uint64(1)
	}
	streamID, stream := suite.CreateStream(isPerpetual, addr, coins, distrTo, startTime2, numEpochsPaidOver)
	return streamID, stream, coins, startTime2
}

// SetupNewStreamWithDenom creates a stream with the specified duration and denom.
func (suite *KeeperTestSuite) SetupNewStreamWithDenom(isPerpetual bool, coins sdk.Coins, denom string) (uint64, *types.Stream, sdk.Coins, time.Time) {
	return suite.setupNewStreamWithDenom(isPerpetual, coins, defaultLockDuration, denom)
}
