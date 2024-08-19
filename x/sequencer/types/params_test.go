package types

import (
	"testing"
	"time"

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
				MinBond:                 sdk.Coin{Denom: "testdenom", Amount: sdk.NewInt(-5)},
				UnbondingTime:           params.UnbondingTime,
				LivenessSlashMultiplier: params.LivenessSlashMultiplier,
			},
			true,
		},
		{
			"invalid unbonding time",
			Params{
				MinBond:                 params.MinBond,
				UnbondingTime:           -time.Second,
				LivenessSlashMultiplier: params.LivenessSlashMultiplier,
			},
			true,
		},
		{
			"invalid liveness slash multiplier",
			Params{
				MinBond:                 params.MinBond,
				UnbondingTime:           params.UnbondingTime,
				LivenessSlashMultiplier: sdk.NewDec(-1),
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
