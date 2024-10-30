package denom

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func FuzzIRODenom(f *testing.F) {
	f.Add("exampleRollappID")
	f.Add("")
	f.Add("123456")
	f.Add("🚀🌕")

	f.Fuzz(func(t *testing.T, rollappID string) {
		denom := IRODenom(rollappID)
		id, ok := RollappIDFromIRODenom(denom)
		require.True(t, ok)
		require.Equal(t, rollappID, id)
	})
}

func FuzzRollappIDFromIRODenom(f *testing.F) {
	f.Add(types.IROTokenPrefix + "exampleRollappID")
	f.Add(types.IROTokenPrefix)
	f.Add("notfuture_prefix")
	f.Add(types.IROTokenPrefix + "🚀🌕")

	f.Fuzz(func(t *testing.T, denom string) {
		rollappID, ok := RollappIDFromIRODenom(denom)
		if ok {
			// Ensure that reconstructing the denom gives the original denom
			reconstructedDenom := IRODenom(rollappID)
			require.Equal(t, denom, reconstructedDenom)
		} else {
			// Denom do not have the prefix
			require.False(t, strings.HasPrefix(denom, types.IROTokenPrefix))
		}
	})
}
