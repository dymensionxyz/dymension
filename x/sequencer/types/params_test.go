package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateBasic(t *testing.T) {
	params := DefaultParams()

	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			"valid params",
			params,
			false,
		},
		{
			"invalid min bond",
			Params{
				MinBond:                    sdk.Coin{Denom: "testdenom", Amount: sdk.NewInt(-5)},
				LivenessSlashMinMultiplier: params.LivenessSlashMinMultiplier,
			},
			true,
		},
		{
			"invalid unbonding time",
			Params{
				MinBond:                    params.MinBond,
				LivenessSlashMinMultiplier: params.LivenessSlashMinMultiplier,
			},
			true,
		},
		{
			"invalid liveness slash multiplier",
			Params{
				MinBond:                    params.MinBond,
				LivenessSlashMinMultiplier: sdk.NewDec(-1),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
