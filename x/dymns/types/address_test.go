package types

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestFallbackAddress_ValidateBasic(t *testing.T) {
	require.NoError(t, FallbackAddress(make([]byte, 20)).ValidateBasic())
	require.NoError(t, FallbackAddress(make([]byte, 32)).ValidateBasic())

	for i := 0; i < 120; i++ {
		if i == 20 || i == 32 {
			continue
		}

		require.Error(t, FallbackAddress(make([]byte, i)).ValidateBasic())
	}
}

func TestFallbackAddress_String(t *testing.T) {
	t.Run("output must be lowercase", func(t *testing.T) {
		bz := make([]byte, 20)
		copy(bz, []byte{0xab, 0xcd, 0xef})
		require.Equal(t, strings.ToLower(common.BytesToAddress(bz).String()), FallbackAddress(bz).String())
	})
}
