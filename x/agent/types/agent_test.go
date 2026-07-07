package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// TestAgentSpendWindow mirrors the eIBC on-demand LP bucket tests: the window
// is an absolute-aligned tumbling bucket that fully resets on rollover.
func TestAgentSpendWindow(t *testing.T) {
	a := types.Agent{
		SpendDenom:          "adym",
		SpendLimitPerWindow: math.NewInt(100),
		SpendWindowBlocks:   10,
	}

	// nil window-spent decodes as zero
	require.True(t, a.SpendAllows(15, math.NewInt(100)))
	require.False(t, a.SpendAllows(15, math.NewInt(101)))

	a.RecordSpend(15, math.NewInt(60))
	require.Equal(t, uint64(10), a.SpendWindowStartHeight)
	require.True(t, a.SpendAllows(19, math.NewInt(40)))
	require.False(t, a.SpendAllows(19, math.NewInt(41)))

	// rollover: capacity resets even though the prior window was exhausted
	a.RecordSpend(19, math.NewInt(40))
	require.False(t, a.SpendAllows(19, math.NewInt(1)))
	require.True(t, a.SpendAllows(20, math.NewInt(100)))

	a.RecordSpend(20, math.NewInt(100))
	require.Equal(t, uint64(20), a.SpendWindowStartHeight)
	require.Equal(t, math.NewInt(100), a.SpendWindowSpent)
	require.False(t, a.SpendAllows(29, math.NewInt(1)))
	require.True(t, a.SpendAllows(30, math.NewInt(100)))
}

func TestAgentSpendDisabled(t *testing.T) {
	a := types.Agent{}
	require.False(t, a.SpendEnabled())
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(100))
	require.False(t, a.SpendAllows(100, math.NewInt(1)))
}
