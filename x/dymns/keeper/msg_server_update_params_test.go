package keeper_test

import (
	"reflect"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_UpdateParams() {
	govModuleAccount := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	updatedPriceParams := func() dymnstypes.PriceParams {
		updated := dymnstypes.DefaultPriceParams()
		updated.AliasPriceSteps = append([]sdkmath.Int{
			updated.AliasPriceSteps[0].AddRaw(1),
		}, updated.AliasPriceSteps...)
		return updated
	}()
	updatedChainsParams := func() dymnstypes.ChainsParams {
		updated := dymnstypes.DefaultChainsParams()
		updated.AliasesOfChainIds = append(updated.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
			ChainId: "pseudo_1-1",
			Aliases: []string{"pseudo"},
		})
		return updated
	}()
	updatedMiscParams := func() dymnstypes.MiscParams {
		updated := dymnstypes.DefaultMiscParams()
		updated.GracePeriodDuration = updated.GracePeriodDuration * 2
		return updated
	}()

	tests := []struct {
		name            string
		msg             *dymnstypes.MsgUpdateParams
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "pass - can update",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  &updatedPriceParams,
				NewChainsParams: &updatedChainsParams,
				NewMiscParams:   &updatedMiscParams,
			},
			wantErr: false,
		},
		{
			name: "fail - reject if not from gov module",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:      dymNsModuleAccAddr.String(),
				NewPriceParams: nil,
				NewChainsParams: &dymnstypes.ChainsParams{
					AliasesOfChainIds: []dymnstypes.AliasesOfChainId{
						{
							ChainId: "pseudo_1-1",
							Aliases: []string{"pseudo"},
						},
					},
				},
				NewMiscParams: nil,
			},
			wantErr:         true,
			wantErrContains: "only the gov module can update params",
		},
		{
			name: "pass - can update price params",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  &updatedPriceParams,
				NewChainsParams: nil,
				NewMiscParams:   nil,
			},
			wantErr: false,
		},
		{
			name: "pass - can update chains params",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  nil,
				NewChainsParams: &updatedChainsParams,
				NewMiscParams:   nil,
			},
			wantErr: false,
		},
		{
			name: "pass - can update price params",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  nil,
				NewChainsParams: nil,
				NewMiscParams:   &updatedMiscParams,
			},
			wantErr: false,
		},
		{
			name: "fail - can not update if all params are nil",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  nil,
				NewChainsParams: nil,
				NewMiscParams:   nil,
			},
			wantErr:         true,
			wantErrContains: "at least one of the new params must be provided",
		},
		{
			name: "fail - can not update if any params is invalid",
			msg: &dymnstypes.MsgUpdateParams{
				Authority:       govModuleAccount,
				NewPriceParams:  &dymnstypes.PriceParams{}, // invalid
				NewChainsParams: &updatedChainsParams,
				NewMiscParams:   &updatedMiscParams,
			},
			wantErr:         true,
			wantErrContains: "invalid argument",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			// use default params
			s.updateModuleParams(func(_ dymnstypes.Params) dymnstypes.Params {
				return dymnstypes.DefaultParams()
			})

			_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).UpdateParams(s.ctx, tt.msg)

			laterModuleParams := s.dymNsKeeper.GetParams(s.ctx)

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Require().True(reflect.DeepEqual(dymnstypes.DefaultParams(), laterModuleParams), "params should not be updated")
				return
			}

			s.Require().NoError(err)

			expectNewModuleParams := dymnstypes.DefaultParams()
			if tt.msg.NewPriceParams != nil {
				expectNewModuleParams.Price = *tt.msg.NewPriceParams
			}
			if tt.msg.NewChainsParams != nil {
				expectNewModuleParams.Chains = *tt.msg.NewChainsParams
			}
			if tt.msg.NewMiscParams != nil {
				expectNewModuleParams.Misc = *tt.msg.NewMiscParams
			}
			s.Require().Falsef(
				reflect.DeepEqual(dymnstypes.DefaultParams(), expectNewModuleParams),
				"bad setup testcase, must provide altered params to test",
			)
			s.Require().Truef(
				reflect.DeepEqual(expectNewModuleParams, laterModuleParams),
				`params should be updated
want: %v
got: %v`, expectNewModuleParams, laterModuleParams,
			)
		})
	}
}
