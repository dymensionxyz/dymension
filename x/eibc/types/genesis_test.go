package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func TestGenesisState_Validate(t *testing.T) {
	validDemandOrder := types.DemandOrder{
		Id:             "1",
		Price:          sdk.Coins{sdk.NewInt64Coin("denom", 2)},
		Fee:            sdk.Coins{sdk.NewInt64Coin("denom", 1)},
		Recipient:      sample.AccAddress(),
		CreationHeight: 1,
	}

	validParams := types.Params{
		EpochIdentifier: "hour",
		TimeoutFee:      math.LegacyNewDecWithPrec(1, 1),
		ErrackFee:       math.LegacyNewDecWithPrec(1, 1),
	}

	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		}, {
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params:       validParams,
				DemandOrders: []types.DemandOrder{validDemandOrder},
			},
			valid: true,
		}, {
			desc: "invalid params",
			genState: &types.GenesisState{
				Params: types.Params{
					TimeoutFee: math.LegacyNewDec(-1),
					ErrackFee:  math.LegacyNewDec(-1),
				},
			},
			valid: false,
		}, {
			desc:     "invalid demand order",
			genState: &types.GenesisState{DemandOrders: []types.DemandOrder{{}}, Params: types.DefaultParams()},
			valid:    false,
		}, {
			desc: "duplicate demand order",
			genState: &types.GenesisState{DemandOrders: []types.DemandOrder{
				validDemandOrder,
				validDemandOrder,
			}, Params: types.DefaultParams()},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
