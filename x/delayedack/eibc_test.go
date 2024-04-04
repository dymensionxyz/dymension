package delayedack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCalculateTimeoutAndAckFee(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		mult, err := sdk.NewDecFromStr("0.0015")
		require.NoError(t, err)
		result, err := calculateTimeoutAndAckFee(mult, "100")
		require.NoError(t, err)
		fmt.Println(result)
	})
}
