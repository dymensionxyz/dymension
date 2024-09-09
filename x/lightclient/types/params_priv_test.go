package types

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExpectedTrustPeriod(t *testing.T) {
	t.Run("happy path - 21 days unbonding", func(t *testing.T) {
		u := time.Hour * 24 * 21
		trust := expectedTrustPeriod(u)
		expect := time.Hour*327 + time.Minute*36
		require.Equal(t, expect, trust)
	})
	t.Run("happy path - 14 days unbonding", func(t *testing.T) {
		u := time.Hour * 24 * 14
		trust := expectedTrustPeriod(u)
		expect := time.Hour*218 + time.Minute*24
		require.Equal(t, expect, trust)
	})
	t.Run("edge case, shortened for testing - 60s unbonding", func(t *testing.T) {
		u := time.Second * 60
		trust := expectedTrustPeriod(u)
		expect := time.Second * 39
		require.Equal(t, expect, trust)
	})
	t.Run("edge case, math - large value", func(t *testing.T) {
		u := time.Duration(math.MaxInt64)
		t.Log(math.MaxInt64)
		scale := int64(10000000000)
		expectApprox := 9223372 * scale * 65
		trust := expectedTrustPeriod(u)
		require.Less(t, int64(trust), expectApprox+2*scale)
		require.Greater(t, expectApprox-2*scale, int64(trust))
	})
}
