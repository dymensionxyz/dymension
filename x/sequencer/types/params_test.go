package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/stretchr/testify/require"
)

var (
	DefaultMinBond = ucoin.SimpleMul(rollapptypes.OneDymCoin, 100)
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
			"invalid notice period",
			Params{
				NoticePeriod:               0,
				LivenessSlashMinMultiplier: params.LivenessSlashMinMultiplier,
				LivenessSlashMinAbsolute:   params.LivenessSlashMinAbsolute,
				KickThreshold:              params.KickThreshold,
			},
			true,
		},
		{
			"invalid liveness slash multiplier",
			Params{
				NoticePeriod:               params.NoticePeriod,
				LivenessSlashMinMultiplier: sdk.NewDec(-1),
				LivenessSlashMinAbsolute:   params.LivenessSlashMinAbsolute,
				KickThreshold:              params.KickThreshold,
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
