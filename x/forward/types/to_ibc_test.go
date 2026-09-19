package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestIBCMinimumValidation(t *testing.T) {
	for _, floor := range []math.Int{{}, math.ZeroInt(), math.NewInt(80), math.NewInt(-1)} {
		hook := NewHookForwardToIBC("channel-0", "destination", 1, floor)
		err := hook.ValidateBasic()
		if !floor.IsNil() && floor.IsNegative() {
			require.ErrorContains(t, err, "min_amount must be non-negative")
		} else {
			require.NoError(t, err)
		}
	}
}
