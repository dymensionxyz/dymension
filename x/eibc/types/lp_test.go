package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func baseLP() OnDemandLP {
	return OnDemandLP{
		FundsAddr:  "cosmos1w3jhxazld3c97ctyv3e97vfjxv6r2djlerhzw4",
		Rollapp:    "rollapp_1234-1",
		Denom:      "stake",
		MaxPrice:   math.NewInt(100),
		MinFee:     math.LegacyZeroDec(),
		SpendLimit: math.NewInt(1000),
	}
}

func order(price int64) *DemandOrder {
	return &DemandOrder{
		Price:     sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(price))),
		Fee:       sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(price))),
		RollappId: "rollapp_1234-1",
	}
}

func TestValidateRateLimit(t *testing.T) {
	t.Run("both unset disables feature", func(t *testing.T) {
		lp := baseLP()
		require.NoError(t, lp.Validate())
	})
	t.Run("amount without blocks fails", func(t *testing.T) {
		lp := baseLP()
		lp.RateLimitAmount = math.NewInt(50)
		require.Error(t, lp.Validate())
	})
	t.Run("blocks without amount fails", func(t *testing.T) {
		lp := baseLP()
		lp.RateLimitBlocks = 10
		require.Error(t, lp.Validate())
	})
	t.Run("both set passes", func(t *testing.T) {
		lp := baseLP()
		lp.RateLimitAmount = math.NewInt(50)
		lp.RateLimitBlocks = 10
		require.NoError(t, lp.Validate())
	})
	t.Run("nil amount with blocks fails", func(t *testing.T) {
		lp := baseLP()
		lp.RateLimitAmount = math.Int{}
		lp.RateLimitBlocks = 10
		require.Error(t, lp.Validate())
	})
}

func TestAcceptsValidityWindow(t *testing.T) {
	lp := baseLP()
	lp.ValidUntilHeight = 100
	r := OnDemandLPRecord{Lp: &lp, Spent: math.ZeroInt(), WindowSpent: math.ZeroInt()}

	require.True(t, r.Accepts(99, order(10)), "accepted just below expiry")
	require.False(t, r.Accepts(100, order(10)), "dead at expiry (exclusive)")
	require.False(t, r.Accepts(101, order(10)), "dead past expiry")

	// 0 = no expiry
	lp.ValidUntilHeight = 0
	require.True(t, r.Accepts(1_000_000, order(10)))
}

func TestRateAllowsTumblingWindow(t *testing.T) {
	lp := baseLP()
	lp.RateLimitAmount = math.NewInt(100)
	lp.RateLimitBlocks = 10
	r := OnDemandLPRecord{Lp: &lp, Spent: math.ZeroInt(), WindowSpent: math.ZeroInt()}

	// fresh window at height 5 (bucket 0): full capacity
	require.True(t, r.RateAllows(5, math.NewInt(100)))
	require.False(t, r.RateAllows(5, math.NewInt(101)))

	// spend 60 in bucket 0
	r.RecordSpend(5, math.NewInt(60))
	require.Equal(t, uint64(0), r.WindowStartHeight)
	require.Equal(t, math.NewInt(60), r.WindowSpent)
	require.True(t, r.RateAllows(7, math.NewInt(40)), "remaining 40 in same bucket")
	require.False(t, r.RateAllows(7, math.NewInt(41)), "over remaining in same bucket")

	// advance into bucket 10: capacity resets even though WindowSpent still 60
	require.True(t, r.RateAllows(12, math.NewInt(100)))

	// recording in new bucket rolls the window
	r.RecordSpend(12, math.NewInt(30))
	require.Equal(t, uint64(10), r.WindowStartHeight)
	require.Equal(t, math.NewInt(30), r.WindowSpent)
}

func TestRateLimitDisabledNoPanic(t *testing.T) {
	lp := baseLP() // RateLimitBlocks == 0
	r := OnDemandLPRecord{Lp: &lp, Spent: math.ZeroInt(), WindowSpent: math.ZeroInt()}
	require.True(t, r.RateAllows(5, math.NewInt(99999)), "disabled always allows")
	require.NotPanics(t, func() { r.RecordSpend(5, math.NewInt(10)) }, "no divide-by-zero when disabled")
	require.Equal(t, math.NewInt(10), r.Spent)
}

// pre-upgrade records decode with a nil WindowSpent; must be treated as zero.
func TestNilWindowSpentSafe(t *testing.T) {
	lp := baseLP()
	lp.RateLimitAmount = math.NewInt(100)
	lp.RateLimitBlocks = 10
	r := OnDemandLPRecord{Lp: &lp, Spent: math.ZeroInt()} // WindowSpent left nil
	require.NotPanics(t, func() { r.RateAllows(5, math.NewInt(50)) })
	require.True(t, r.RateAllows(5, math.NewInt(100)))
}

// new record fields must round-trip through proto marshal/unmarshal, and a
// record encoded without them must decode cleanly with the feature disabled.
func TestRecordProtoRoundTrip(t *testing.T) {
	lp := baseLP()
	lp.ValidUntilHeight = 100
	lp.RateLimitAmount = math.NewInt(50)
	lp.RateLimitBlocks = 10
	r := OnDemandLPRecord{Id: 1, Lp: &lp, Spent: math.NewInt(5), WindowStartHeight: 10, WindowSpent: math.NewInt(30)}

	bz, err := r.Marshal()
	require.NoError(t, err)
	var got OnDemandLPRecord
	require.NoError(t, got.Unmarshal(bz))
	require.Equal(t, r.WindowStartHeight, got.WindowStartHeight)
	require.Equal(t, r.WindowSpent, got.WindowSpent)
	require.Equal(t, r.Lp.ValidUntilHeight, got.Lp.ValidUntilHeight)
	require.Equal(t, r.Lp.RateLimitAmount, got.Lp.RateLimitAmount)
	require.Equal(t, r.Lp.RateLimitBlocks, got.Lp.RateLimitBlocks)

	// legacy record: no new fields set
	old := baseLP()
	legacy := OnDemandLPRecord{Id: 2, Lp: &old, Spent: math.NewInt(1)}
	bz, err = legacy.Marshal()
	require.NoError(t, err)
	var decoded OnDemandLPRecord
	require.NoError(t, decoded.Unmarshal(bz))
	require.Equal(t, uint64(0), decoded.WindowStartHeight)
	require.Equal(t, uint64(0), decoded.Lp.ValidUntilHeight)
	require.True(t, decoded.RateAllows(5, math.NewInt(99999)), "feature disabled")
}
