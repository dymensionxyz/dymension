package types_test

import (
	"math/big"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestAgentSpendRecipientAllowed_EquivalentAddressEncoding(t *testing.T) {
	_, _, allowed := testdata.KeyTestPubAddr()
	_, _, denied := testdata.KeyTestPubAddr()
	agent := types.Agent{SpendRecipientAllowlist: []string{allowed.String()}}

	require.True(t, agent.SpendRecipientAllowed(strings.ToUpper(allowed.String())))
	require.False(t, agent.SpendRecipientAllowed(denied.String()))
	require.True(t, (types.Agent{}).SpendRecipientAllowed(denied.String()))
}

func windowAgent(limit int64, windowBlocks uint64) types.Agent {
	return types.Agent{
		SpendDenom:          "adym",
		SpendLimitPerWindow: math.NewInt(limit),
		SpendWindowBlocks:   windowBlocks,
	}
}

// TestAgentSpendWindow_SingleBucketCap checks that spend within one bucket is
// capped at exactly the limit.
func TestAgentSpendWindow_SingleBucketCap(t *testing.T) {
	a := windowAgent(100, 10)

	// nil window-spent decodes as zero
	require.True(t, a.SpendAllows(15, math.NewInt(100)))
	require.False(t, a.SpendAllows(15, math.NewInt(101)))

	a.RecordSpend(15, math.NewInt(60))
	require.Equal(t, uint64(10), a.SpendWindowStartHeight)
	require.True(t, a.SpendAllows(19, math.NewInt(40)))
	require.False(t, a.SpendAllows(19, math.NewInt(41)))

	a.RecordSpend(19, math.NewInt(40))
	require.Equal(t, math.NewInt(100), a.SpendWindowSpent)
	require.False(t, a.SpendAllows(19, math.NewInt(1)))
}

// TestAgentSpendWindow_BoundaryBurstBlocked is the regression test for the
// tumbling-window boundary burst: exhausting the budget in the last block of a
// bucket must not grant a fresh full budget in the first block of the next.
func TestAgentSpendWindow_BoundaryBurstBlocked(t *testing.T) {
	const limit, w = 1000, 100
	a := windowAgent(limit, w)
	a.RecordSpend(199, math.NewInt(limit))

	require.False(t, a.SpendAllows(200, math.NewInt(1)))
	for k := int64(0); k < w; k++ {
		allowed := limit - limit*(w-k)/w
		height := uint64(200 + k) //nolint:gosec // test heights are small and positive
		require.Equal(t, math.NewInt(allowed), a.RemainingWindowBudget(height), "k=%d", k)
		require.True(t, a.SpendAllows(height, math.NewInt(allowed)), "k=%d", k)
		require.False(t, a.SpendAllows(height, math.NewInt(allowed+1)), "k=%d", k)
	}
}

// TestAgentSpendWindow_OneBucketAdvanceCarriesPrev checks that a full previous
// bucket leaves zero budget at the next bucket start, rising linearly to the
// full limit by the bucket end, and that recording migrates the old bucket
// into the previous-window slot.
func TestAgentSpendWindow_OneBucketAdvanceCarriesPrev(t *testing.T) {
	a := windowAgent(100, 10)
	a.RecordSpend(12, math.NewInt(100))

	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(20))
	require.Equal(t, math.NewInt(50), a.RemainingWindowBudget(25))
	require.Equal(t, math.NewInt(90), a.RemainingWindowBudget(29))
	require.Equal(t, math.NewInt(100), a.RemainingWindowBudget(30))

	a.RecordSpend(25, math.NewInt(30))
	require.Equal(t, uint64(20), a.SpendWindowStartHeight)
	require.Equal(t, math.NewInt(30), a.SpendWindowSpent)
	require.Equal(t, math.NewInt(100), a.SpendPrevWindowSpent)
	require.Equal(t, math.NewInt(20), a.RemainingWindowBudget(25))

	// advancing one more bucket drops the oldest bucket and carries the 30
	a.RecordSpend(30, math.NewInt(10))
	require.Equal(t, uint64(30), a.SpendWindowStartHeight)
	require.Equal(t, math.NewInt(10), a.SpendWindowSpent)
	require.Equal(t, math.NewInt(30), a.SpendPrevWindowSpent)
	require.Equal(t, math.NewInt(60), a.RemainingWindowBudget(30))
}

// TestAgentSpendWindow_QuietGapFullReset checks that after two or more empty
// buckets both buckets are treated as empty, both when reading and recording.
func TestAgentSpendWindow_QuietGapFullReset(t *testing.T) {
	a := windowAgent(100, 10)
	a.RecordSpend(12, math.NewInt(100))
	a.RecordSpend(25, math.NewInt(0))
	require.Equal(t, math.NewInt(100), a.SpendPrevWindowSpent)

	require.Equal(t, math.NewInt(100), a.RemainingWindowBudget(40))
	a.RecordSpend(40, math.NewInt(70))
	require.Equal(t, uint64(40), a.SpendWindowStartHeight)
	require.Equal(t, math.NewInt(70), a.SpendWindowSpent)
	require.True(t, a.SpendPrevWindowSpent.IsZero())
}

// TestAgentSpendWindow_TrailingWindowNearDoubleExposure pins the accepted
// limitation of the weighted estimate: spend concentrated late in the previous
// bucket is under-weighted, so a single W-block interval (heights 199..298) can
// carry nearly 2x the limit.
func TestAgentSpendWindow_TrailingWindowNearDoubleExposure(t *testing.T) {
	const limit, w = 1000, 100
	a := windowAgent(limit, w)
	a.RecordSpend(199, math.NewInt(limit))

	require.Equal(t, math.NewInt(980), a.RemainingWindowBudget(298))
	a.RecordSpend(298, math.NewInt(980))
	require.Equal(t, math.NewInt(1980), a.SpendWindowSpent.Add(a.SpendPrevWindowSpent))
	require.Equal(t, w, 298-199+1, "heights 199..298 span exactly one window")
}

// TestAgentSpendWindow_MaxValuesDoNotOverflow covers spend state at the top of
// math.Int's range, which validation accepts: weighting must not overflow in
// the intermediate product and the budget must saturate at zero rather than
// overflow when current plus weighted previous spend exceeds the limit.
func TestAgentSpendWindow_MaxValuesDoNotOverflow(t *testing.T) {
	maxInt := math.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))
	half := maxInt.QuoRaw(2)

	a := windowAgent(0, 2)
	a.SpendLimitPerWindow = maxInt
	a.SpendWindowStartHeight = 2
	a.SpendWindowSpent = math.ZeroInt()
	a.SpendPrevWindowSpent = maxInt
	require.NoError(t, a.ValidateSpendState(types.DefaultParams().SpendRecipientAllowlistMax))

	require.Equal(t, maxInt, types.PrevWindowWeighted(maxInt, 0, 2))
	require.Equal(t, half, types.PrevWindowWeighted(maxInt, 1, 2))
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(2))
	require.Equal(t, maxInt.Sub(half), a.RemainingWindowBudget(3))
	require.False(t, a.SpendAllows(2, math.OneInt()))

	// current and previous bucket both at the limit: their sum exceeds 256 bits
	a.SpendWindowSpent = maxInt
	require.NoError(t, a.ValidateSpendState(types.DefaultParams().SpendRecipientAllowlistMax))
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(2))
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(3))
	require.False(t, a.SpendAllows(3, math.OneInt()))
}

func TestPrevWindowWeighted(t *testing.T) {
	tests := []struct {
		prev    int64
		elapsed uint64
		w       uint64
		want    int64
	}{
		{1000, 0, 100, 1000},
		{1000, 1, 100, 990},
		{1000, 50, 100, 500},
		{1000, 99, 100, 10},
		{7, 99, 100, 0},
		{7, 1, 3, 4},
		{100, 9, 10, 10},
		{0, 5, 10, 0},
		{5, 0, 1, 5},
	}
	for _, tc := range tests {
		got := types.PrevWindowWeighted(math.NewInt(tc.prev), tc.elapsed, tc.w)
		require.Equal(t, math.NewInt(tc.want), got, "prev=%d elapsed=%d w=%d", tc.prev, tc.elapsed, tc.w)
	}
}

func TestAgentSpendDisabled(t *testing.T) {
	a := types.Agent{}
	require.False(t, a.SpendEnabled())
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(100))
	require.False(t, a.SpendAllows(100, math.NewInt(1)))
}
