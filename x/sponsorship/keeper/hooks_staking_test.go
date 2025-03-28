package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestHooks() {
	testCases := []struct {
		name                 string
		prepare              func()
		initialDistr         types.Distribution // initial test distr
		expectedDistr        types.Distribution // final distr after applying all the updates
		expectedUpdateEvents int
	}{
		{
			name: "User increases their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(valAddr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				s.Delegate(delAddr, valAddr, update)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, valAddr, initial.Add(update).Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(400_000)},
					{GaugeId: 2, Power: math.NewInt(1_000_000)},
				},
			},
			expectedUpdateEvents: 1,
		},
		{
			name: "User with 2 validators",
			prepare: func() {
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				val2 := s.CreateValidator()
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				del := s.CreateDelegator(val1addr, initial) // delegator 1 -> validator 1
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.Delegate(delAddr, val2addr, initial) // delegator 1 -> validator 2

				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, val1addr, initial.Amount)
				s.AssertDelegatorValidator(delAddr, val2addr, initial.Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(400_000)},
					{GaugeId: 2, Power: math.NewInt(1_000_000)},
				},
			},
			expectedUpdateEvents: 0,
		},
		{
			name: "User with 2 validators, partial unbonding from one of them",
			prepare: func() {
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				val2 := s.CreateValidator()
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				del := s.CreateDelegator(val1addr, initial) // delegator 1 -> validator 1
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.Delegate(delAddr, val2addr, initial) // delegator 1 -> validator 2

				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				del = s.Undelegate(delAddr, val1addr, update)
				s.NotNil(del) // partial unbonding

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, val1addr, initial.Sub(update).Amount)
				s.AssertDelegatorValidator(delAddr, val2addr, initial.Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_200_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(240_000)},
					{GaugeId: 2, Power: math.NewInt(600_000)},
				},
			},
			expectedUpdateEvents: 1,
		},
		{
			name: "User with 2 validators, complete unbonding from one of them",
			prepare: func() {
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				val2 := s.CreateValidator()
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				del := s.CreateDelegator(val1addr, initial) // delegator 1 -> validator 1
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.Delegate(delAddr, val2addr, initial) // delegator 1 -> validator 2

				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, initial.Amount)
				del = s.Undelegate(delAddr, val1addr, update)
				s.Nil(del)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertNotHaveDelegatorValidator(delAddr, val1addr)
				s.AssertDelegatorValidator(delAddr, val2addr, initial.Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(200_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedUpdateEvents: 1,
		},
		{
			name: "User decreases their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(valAddr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				del = s.Undelegate(delAddr, valAddr, update)
				s.NotNil(del) // partial unbonding

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, valAddr, initial.Sub(update).Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(200_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(40_000)},
					{GaugeId: 2, Power: math.NewInt(100_000)},
				},
			},
			expectedUpdateEvents: 1,
		},
		{
			name: "User completely removes their delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(valAddr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				// Completely undelegate
				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				finalDel := s.Undelegate(delAddr, valAddr, update)
				s.Nil(finalDel)

				// Check the state
				s.AssertNotVoted(delAddr)
				s.AssertNotHaveDelegatorValidator(delAddr, valAddr)
			},
			initialDistr:         types.NewDistribution(),
			expectedDistr:        types.NewDistribution(), // empty distribution
			expectedUpdateEvents: 1,
		},
		{
			name: "User partially redelegates",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				val2 := s.CreateValidator()
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(val1addr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(300_000))
				src, dst := s.BeginRedelegate(delAddr, val1addr, val2addr, update)
				s.NotNil(src)
				s.NotNil(dst)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, val1addr, initial.Sub(update).Amount)
				s.AssertDelegatorValidator(delAddr, val2addr, update.Amount)
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
			expectedUpdateEvents: 2,
		},
		{
			name: "User completely redelegates, the vote is deleted",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				val2 := s.CreateValidator()
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(val1addr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				src, dst := s.BeginRedelegate(delAddr, val1addr, val2addr, update)
				s.Nil(src)
				s.NotNil(dst)

				// Check the state
				s.AssertNotVoted(delAddr)
				s.AssertNotHaveDelegatorValidator(delAddr, val1addr)
				s.AssertNotHaveDelegatorValidator(delAddr, val2addr)
			},
			initialDistr:         types.NewDistribution(),
			expectedDistr:        types.NewDistribution(),
			expectedUpdateEvents: 1,
		},
		{
			name: "Completely cancel unbonding delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(valAddr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				s.Ctx = s.Ctx.WithBlockHeight(10)
				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				s.Undelegate(delAddr, valAddr, update)
				s.CancelUnbondingDelegation(delAddr, valAddr, s.Ctx.BlockHeight(), update)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, valAddr, initial.Amount)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(200_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedUpdateEvents: 2,
		},
		{
			name: "Partially cancel unbonding delegation",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val := s.CreateValidator()
				valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(valAddr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				s.Ctx = s.Ctx.WithBlockHeight(10)

				update := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(800_000))
				s.Undelegate(delAddr, valAddr, update)

				partiallyCancel := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(400_000))
				s.CancelUnbondingDelegation(delAddr, valAddr, s.Ctx.BlockHeight(), partiallyCancel)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, valAddr, initial.Sub(update).Add(partiallyCancel).Amount)
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
			expectedUpdateEvents: 2,
		},
		{
			name: "User becomes a validator",
			prepare: func() {
				initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
				val1 := s.CreateValidator()
				val1addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
				s.Require().NoError(err)
				del := s.CreateDelegator(val1addr, initial)
				delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())
				s.CreateGauges(2)

				s.Vote(types.MsgVote{
					Voter: del.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: types.DYM.MulRaw(20)},
						{GaugeId: 2, Weight: types.DYM.MulRaw(50)},
					},
				})

				update := math.NewInt(1_000_000)
				val2 := s.CreateValidatorWithAddress(delAddr, update)
				val2addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
				s.Require().NoError(err)

				// Check the state
				s.AssertVoted(delAddr)
				s.AssertDelegatorValidator(delAddr, val1addr, initial.Amount)
				s.AssertDelegatorValidator(delAddr, val2addr, update)
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(400_000)},
					{GaugeId: 2, Power: math.NewInt(1_000_000)},
				},
			},
			expectedUpdateEvents: 1,
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

			s.AssertEventEmitted(s.Ctx, proto.MessageName(new(types.EventVotingPowerUpdate)), tc.expectedUpdateEvents)
		})
	}
}
