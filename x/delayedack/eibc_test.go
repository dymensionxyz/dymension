package delayedack

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCalculateTimeoutAndAckFee(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mult, err := sdk.NewDecFromStr("0.0015")
		require.NoError(t, err)
		_, err = calculateTimeoutAndAckFee(mult, "100")
		require.ErrorIs(t, ErrFeeIsNotPositive, err)
	})
	t.Run("simple", func(t *testing.T) {
		mult, err := sdk.NewDecFromStr("0.0015")
		require.NoError(t, err)
		_, err = calculateTimeoutAndAckFee(mult, "1000000")
		require.NoError(t, err)
	})
}
