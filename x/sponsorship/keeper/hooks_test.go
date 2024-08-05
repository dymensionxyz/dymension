package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestHooks() {
	testCases := []struct {
		name          string
		prepare       func()
		initialDistr  types.Distribution // initial test distr
		expectedDistr types.Distribution // final distr after applying all the updates
	}{
		{
			name: "User increases their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				del := s.CreateDelegator(val.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				s.Delegate(del.GetDelegatorAddr(), val.GetOperator(), update)

				// Check the corresponding power
				vp, err := s.App.SponsorshipKeeper.GetDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val.GetOperator())
				s.Require().NoError(err)
				s.Require().Equal(initial.Amount.Add(update.Amount), vp)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(400_000)},
					{GaugeId: 2, Power: math.NewInt(1_000_000)},
				},
			},
		},
		{
			name: "User decreases their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				del := s.CreateDelegator(val.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				s.Undelegate(del.GetDelegatorAddr(), val.GetOperator(), update)

				// Check the corresponding power
				vp, err := s.App.SponsorshipKeeper.GetDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val.GetOperator())
				s.Require().NoError(err)
				s.Require().Equal(initial.Amount.Sub(update.Amount), vp)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(200_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(40_000)},
					{GaugeId: 2, Power: math.NewInt(100_000)},
				},
			},
		},
		{
			name: "User completely removes their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				del := s.CreateDelegator(val.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				// Completely undelegate
				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				finalDel := s.Undelegate(del.GetDelegatorAddr(), val.GetOperator(), update)
				s.Require().Nil(finalDel)

				hasRecord, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val.GetOperator())
				s.Require().NoError(err)
				s.Require().False(hasRecord)
			},
			initialDistr:  types.NewDistribution(),
			expectedDistr: types.NewDistribution(), // empty distribution
		},
		{
			name: "User partially redelegates",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val1 := s.CreateValidator()
				val2 := s.CreateValidator()
				del := s.CreateDelegator(val1.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(300_000))
				s.BeginRedelegate(del.GetDelegatorAddr(), val1.GetOperator(), val2.GetOperator(), update)

				// Check the corresponding power
				vp1, err := s.App.SponsorshipKeeper.GetDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val1.GetOperator())
				s.Require().NoError(err)
				s.Require().Equal(initial.Amount.Sub(update.Amount), vp1)

				vp2, err := s.App.SponsorshipKeeper.GetDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val2.GetOperator())
				s.Require().NoError(err)
				s.Require().Equal(update.Amount, vp2)
			},
			initialDistr: types.NewDistribution(),
			// final distribution is the same, but validator indexes are updated in the state
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(200_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
		},
		{
			name: "User completely redelegates, the vote is deleted",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val1 := s.CreateValidator()
				val2 := s.CreateValidator()
				del := s.CreateDelegator(val1.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				s.BeginRedelegate(del.GetDelegatorAddr(), val1.GetOperator(), val2.GetOperator(), update)

				// Check the corresponding power
				hasVP1Record, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val1.GetOperator())
				s.Require().NoError(err)
				s.Require().False(hasVP1Record)

				// Check the corresponding power
				hasVP2Record, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, del.GetDelegatorAddr(), val2.GetOperator())
				s.Require().NoError(err)
				s.Require().False(hasVP2Record)
			},
			initialDistr:  types.NewDistribution(),
			expectedDistr: types.NewDistribution(),
		},
		{
			name: "Completely cancel unbonding delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				del := s.CreateDelegator(val.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				s.Undelegate(del.GetDelegatorAddr(), val.GetOperator(), update)

				s.CancelUnbondingDelegation(del.GetDelegatorAddr(), val.GetOperator(), s.Ctx.BlockHeight(), update)

				// TODO: check balances
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(200_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
		},
		{
			name: "Partially cancel unbonding delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				del := s.CreateDelegator(val.GetOperator(), initial)
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr().String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				s.Undelegate(del.GetDelegatorAddr(), val.GetOperator(), update)

				partiallyCancel := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(400_000))
				s.CancelUnbondingDelegation(del.GetDelegatorAddr(), val.GetOperator(), s.Ctx.BlockHeight(), partiallyCancel)

				// TODO: check balances
			},
			initialDistr: types.NewDistribution(),
			// 600_000 = 1_000_000 - 800_000 + 400_000
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(600_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(120_000)},
					{GaugeId: 2, Power: math.NewInt(300_000)},
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set initial distribution
			err := s.App.SponsorshipKeeper.SaveDistribution(s.Ctx, tc.initialDistr)
			s.Require().NoError(err)

			tc.prepare()

			// Check the distribution is correct
			distr := s.GetDistribution()
			err = distr.Validate()
			s.Require().NoError(err)
			s.Require().True(tc.expectedDistr.Equal(distr), "expect: %v\nactual: %v", tc.expectedDistr, distr)
		})
	}
}
