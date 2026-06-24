package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const escDenom = "stake"

// base order: fee=50, price=150, creation=100, escalates to maxFee=150 over 10 blocks.
func escOrder(escalation *FeeEscalation) *DemandOrder {
	return &DemandOrder{
		Price:          sdk.NewCoins(sdk.NewCoin(escDenom, math.NewInt(150))),
		Fee:            sdk.NewCoins(sdk.NewCoin(escDenom, math.NewInt(50))),
		CreationHeight: 100,
		FeeEscalation:  escalation,
	}
}

func TestEffectiveFeeAndPrice(t *testing.T) {
	esc := &FeeEscalation{MaxFeeAmount: math.NewInt(150), DurationBlocks: 10}

	tests := []struct {
		name      string
		o         *DemandOrder
		height    uint64
		wantFee   int64
		wantPrice int64
	}{
		{"at creation", escOrder(esc), 100, 50, 150},
		{"before creation", escOrder(esc), 99, 50, 150},
		{"midpoint", escOrder(esc), 105, 100, 100},
		{"end of window", escOrder(esc), 110, 150, 50},
		{"past window (capped)", escOrder(esc), 200, 150, 50},
		{"nil escalation", escOrder(nil), 105, 50, 150},
		{"zero duration", escOrder(&FeeEscalation{MaxFeeAmount: math.NewInt(150), DurationBlocks: 0}), 105, 50, 150},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, math.NewInt(tt.wantFee), tt.o.EffectiveFeeAmount(tt.height), "fee")
			require.Equal(t, math.NewInt(tt.wantPrice), tt.o.EffectivePriceAmount(tt.height), "price")
		})
	}
}

func TestEffectiveFeePercent(t *testing.T) {
	o := escOrder(&FeeEscalation{MaxFeeAmount: math.NewInt(150), DurationBlocks: 10})
	// midpoint: fee=100, price=100 => 1.0
	require.Equal(t, math.LegacyOneDec(), o.EffectiveFeePercent(105))
	// nil escalation falls back to base GetFeePercent (50/150)
	require.Equal(t, escOrder(nil).GetFeePercent(), escOrder(nil).EffectiveFeePercent(105))
}

func TestApplyEffectiveFee(t *testing.T) {
	o := escOrder(&FeeEscalation{MaxFeeAmount: math.NewInt(150), DurationBlocks: 10})
	o.ApplyEffectiveFee(105)
	require.Nil(t, o.FeeEscalation)
	require.Equal(t, math.NewInt(100), o.GetFeeAmount())
	require.Equal(t, math.NewInt(100), o.PriceAmount())

	// no-op for static orders
	static := escOrder(nil)
	static.ApplyEffectiveFee(105)
	require.Equal(t, math.NewInt(50), static.GetFeeAmount())
	require.Equal(t, math.NewInt(150), static.PriceAmount())
}
