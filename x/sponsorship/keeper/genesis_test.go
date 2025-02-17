package keeper_test

import (
	"slices"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	val1 := s.CreateValidator()
	val1Addr, _ := sdk.ValAddressFromBech32(val1.GetOperator())
	val2 := s.CreateValidator()
	val2Addr, _ := sdk.ValAddressFromBech32(val2.GetOperator())
	initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))

	del11 := s.CreateDelegator(val1Addr, initial) // delegator 1 -> validator 1
	del1Addr := sdk.MustAccAddressFromBech32(del11.GetDelegatorAddr())
	s.Delegate(del1Addr, val2Addr, initial) // delegator 1 -> validator 2

	del22 := s.CreateDelegator(val2Addr, initial) // delegator 2 -> validator 2
	del2Addr := sdk.MustAccAddressFromBech32(del22.GetDelegatorAddr())

	testCases := []struct {
		name          string
		genesis       types.GenesisState
		expectedDistr types.Distribution
	}{
		{
			name: "Import/Export",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				VoterInfos: []types.VoterInfo{
					{
						Voter: del1Addr.String(),
						Vote: types.Vote{
							VotingPower: math.NewInt(600),
							Weights: []types.GaugeWeight{
								{GaugeId: 1, Weight: types.DYM.MulRaw(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: val1Addr.String(), Power: math.NewInt(400)},
							{Validator: val2Addr.String(), Power: math.NewInt(200)},
						},
					},
					{
						Voter: del2Addr.String(),
						Vote: types.Vote{
							VotingPower: math.NewInt(400),
							Weights: []types.GaugeWeight{
								{GaugeId: 2, Weight: types.DYM.MulRaw(100)},
							},
						},
						Validators: []types.ValidatorVotingPower{
							{Validator: val2Addr.String(), Power: math.NewInt(400)},
						},
					},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(600)},
					{GaugeId: 2, Power: math.NewInt(400)},
				},
			},
		},
		{
			name:          "Default is valid",
			genesis:       *types.DefaultGenesis(),
			expectedDistr: types.NewDistribution(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			err := s.App.SponsorshipKeeper.ImportGenesis(s.Ctx, tc.genesis)
			s.Require().NoError(err)

			// Check the distribution is correct
			distr := s.GetDistribution()
			err = distr.Validate()
			s.Require().NoError(err)
			s.Require().True(tc.expectedDistr.Equal(distr), "expect: %v\nactual: %v", tc.expectedDistr, distr)

			// Check all values are in the state
			for _, info := range tc.genesis.VoterInfos {
				voter, err := sdk.AccAddressFromBech32(info.Voter)
				s.Require().NoError(err)

				voted, err := s.App.SponsorshipKeeper.Voted(s.Ctx, voter)
				s.Require().NoError(err)
				s.Require().True(voted)

				for _, val := range info.Validators {
					validator, err := sdk.ValAddressFromBech32(val.Validator)
					s.Require().NoError(err)

					has, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, voter, validator)
					s.Require().NoError(err)
					s.Require().True(has)
				}
			}

			actual, err := s.App.SponsorshipKeeper.ExportGenesis(s.Ctx)
			s.Require().NoError(err)

			err = actual.Validate()
			s.Require().NoError(err)

			sortGenState(tc.genesis)
			sortGenState(actual)
			s.Require().Equal(tc.genesis.Params, actual.Params, "expect: %v\nactual: %v", tc.genesis, actual)
			s.Require().ElementsMatch(tc.genesis.VoterInfos, actual.VoterInfos, "expect: %v\nactual: %v", tc.genesis, actual)
		})
	}
}

// sortGenState sorts all underlying slices in the ascending order.
func sortGenState(gs types.GenesisState) {
	for i := range gs.VoterInfos {
		slices.SortFunc(gs.VoterInfos[i].Validators, func(a, b types.ValidatorVotingPower) int {
			switch {
			case a.Validator < b.Validator:
				return -1
			case a.Validator > b.Validator:
				return 1
			}
			return 0
		})
	}

	slices.SortFunc(gs.VoterInfos, func(a, b types.VoterInfo) int {
		switch {
		case a.Voter < b.Voter:
			return -1
		case a.Voter > b.Voter:
			return 1
		}
		return 0
	})
}
