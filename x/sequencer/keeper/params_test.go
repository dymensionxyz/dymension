package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.SequencerKeeper(t)
	params := types.DefaultParams()
	params.MinBond = sdk.NewCoin("testdenom", sdk.NewInt(100))

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}

// test ValidateParams
func (s *SequencerTestSuite) TestValidateParams() {
	k := s.App.SequencerKeeper

	tests := []struct {
		name    string
		params  func() types.Params
		wantErr bool
	}{
		{
			"stateful validation: default params",
			func() types.Params {
				return types.DefaultParams()
			},
			false,
		},
		{
			"stateful validation: min bond denom not equal to base denom",
			func() types.Params {
				params := types.DefaultParams()
				params.MinBond.Denom = "testdenom"
				return params
			},
			true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := k.ValidateParams(s.Ctx, tt.params())
			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
