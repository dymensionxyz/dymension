package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	require.Equal(
		t,
		time.Now().Unix(), time.Now().UTC().Unix(),
		"if mis-match, 100% sure will causes AppHash",
	)
}

func Test_consumeMinimumGas(t *testing.T) {
	tests := []struct {
		name                    string
		originalConsumedGas     uint64
		overrideConsumedGas     *uint64
		minimumGas              uint64
		wantPanic               bool
		wantGasMeterConsumedGas uint64
	}{
		{
			name:                    "pass - normal gas consumption",
			originalConsumedGas:     0,
			minimumGas:              1_000,
			wantGasMeterConsumedGas: 1_000,
		},
		{
			name:                    "pass - should be stacked with previous run",
			originalConsumedGas:     20_000,
			minimumGas:              1_000,
			wantGasMeterConsumedGas: 21_000,
		},
		{
			name:                "fail - should panic if later consumed gas is less than original consumed gas",
			originalConsumedGas: 2,
			overrideConsumedGas: func() *uint64 {
				v := uint64(1)
				return &v
			}(),
			minimumGas: 1_000,
			wantPanic:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithGasMeter(sdk.NewInfiniteGasMeter())

			originalConsumedGas := tt.originalConsumedGas
			if tt.overrideConsumedGas != nil {
				originalConsumedGas = *tt.overrideConsumedGas
			}
			if originalConsumedGas > 0 {
				ctx.GasMeter().ConsumeGas(originalConsumedGas, "simulate pre-run gas consumption")
			}

			if tt.wantPanic {
				require.Panics(t, func() {
					consumeMinimumGas(ctx, tt.minimumGas, tt.originalConsumedGas, "test")
				})
				return
			}

			consumeMinimumGas(ctx, tt.minimumGas, tt.originalConsumedGas, "test")
			require.Equal(t, tt.wantGasMeterConsumedGas, ctx.GasMeter().GasConsumed())
		})
	}
}
