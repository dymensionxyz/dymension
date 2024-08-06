package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestMsgVote() {
	addr := apptesting.CreateRandomAccounts(3)

	type delegation struct {
		delegator  sdk.AccAddress
		delegation sdk.Coin
	}

	testCases := []struct {
		name          string
		params        types.Params       // module params
		numGauges     int                // number of initial gauges, IDs fall between [1; numGauges]
		delegations   []delegation       // initial delegations
		votes         []types.MsgVote    // all votes, votes are applied one by one; only one element is expected in case of the error
		initialDistr  types.Distribution // initial test distr
		expectedDistr types.Distribution // final distr after applying all the votes
		expectErr     bool               // if the error is expected, the vote slice must contain only one element
		errorContains string
	}{
		{
			name:      "Valid, 1 voter, empty initial",
			params:    types.DefaultParams(),
			numGauges: 2,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
			},
			initialDistr: types.NewDistribution(),
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(200_000)},
					{GaugeId: 2, Power: math.NewInt(300_000)},
				},
			},
			expectErr:     false,
			errorContains: "",
		},
		{
			name:      "Valid, 1 voter, non-empty initial",
			params:    types.DefaultParams(),
			numGauges: 2,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(700_000)},
					{GaugeId: 2, Power: math.NewInt(800_000)},
				},
			},
			expectErr:     false,
			errorContains: "",
		},
		{
			name:      "Valid, 3 voters, non-empty initial",
			params:    types.DefaultParams(),
			numGauges: 3,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
				{
					delegator:  addr[1],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
				{
					delegator:  addr[2],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				// [gauge1, 20%] [gauge2,  0%] [gauge3, 40%] power = 1_000_000
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 3, Weight: math.NewInt(40)},
					},
				},
				// [gauge1,  0%] [gauge2, 30%] [gauge3, 20%] power = 1_000_000
				{
					Voter: addr[1].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 2, Weight: math.NewInt(30)},
						{GaugeId: 3, Weight: math.NewInt(20)},
					},
				},
				// [gauge1, 40%] [gauge2, 20%] [gauge3,  0%] power = 1_000_000
				{
					Voter: addr[2].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(40)},
						{GaugeId: 2, Weight: math.NewInt(20)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(200_000)},
					{GaugeId: 3, Power: math.NewInt(300_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(4_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(1_100_000)},
					{GaugeId: 2, Power: math.NewInt(700_000)},
					{GaugeId: 3, Power: math.NewInt(900_000)},
				},
			},
			expectErr:     false,
			errorContains: "",
		},
		{
			name:      "Voter re-votes",
			params:    types.DefaultParams(),
			numGauges: 2,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(50)},
						{GaugeId: 2, Weight: math.NewInt(40)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(2_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(1_000_000)},
					{GaugeId: 2, Power: math.NewInt(900_000)},
				},
			},
			expectErr:     false,
			errorContains: "",
		},
		{
			name:      "Unknown gauge",
			params:    types.DefaultParams(),
			numGauges: 1,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
				},
			},
			expectErr:     true,
			errorContains: "failed to get gauge by id '2'",
		},
		{
			name: "Weight is less than the min allocation",
			params: types.Params{
				MinAllocationWeight: math.NewInt(30),
				MinVotingPower:      types.DefaultMinVotingPower,
			},
			numGauges: 2,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectErr:     true,
			errorContains: "gauge weight '20' is less than min allocation weight '30'",
		},
		{
			name: "Not enough voting power",
			params: types.Params{
				MinAllocationWeight: types.DefaultMinAllocationWeight,
				MinVotingPower:      math.NewInt(2_000_000),
			},
			numGauges: 2,
			delegations: []delegation{
				{
					delegator:  addr[0],
					delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
				},
			},
			votes: []types.MsgVote{
				{
					Voter: addr[0].String(),
					Weights: []types.GaugeWeight{
						{GaugeId: 1, Weight: math.NewInt(20)},
						{GaugeId: 2, Weight: math.NewInt(30)},
					},
				},
			},
			initialDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectedDistr: types.Distribution{
				VotingPower: math.NewInt(1_000_000),
				Gauges: []types.Gauge{
					{GaugeId: 1, Power: math.NewInt(500_000)},
					{GaugeId: 2, Power: math.NewInt(500_000)},
				},
			},
			expectErr:     true,
			errorContains: "voting power '1000000' is less than min voting power expected '2000000'",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Create expected num of gauges
			s.CreateGauges(tc.numGauges)

			// Create a validator
			val := s.CreateValidator()

			// Fund voter addresses and delegate to the validator
			for _, d := range tc.delegations {
				// Fund delegator's address
				apptesting.FundAccount(s.App, s.Ctx, d.delegator, sdk.Coins{d.delegation})

				// Delegate to the validator
				_ = s.Delegate(d.delegator, val.GetOperator(), d.delegation)
			}

			// Set the initial distribution
			err := s.App.SponsorshipKeeper.SaveDistribution(s.Ctx, tc.initialDistr)
			s.Require().NoError(err)

			// Set module params
			err = s.App.SponsorshipKeeper.SetParams(s.Ctx, tc.params)
			s.Require().NoError(err)

			// Cast the votes
			for _, v := range tc.votes {
				voter, err := sdk.AccAddressFromBech32(v.Voter)
				s.Require().NoError(err)

				voteResp, err := s.msgServer.Vote(s.Ctx, &v)

				if tc.expectErr {
					s.Require().Error(err)
					s.Require().ErrorContains(err, tc.errorContains)

					// Check the vote is not in the state
					voted, err := s.App.SponsorshipKeeper.Voted(s.Ctx, voter)
					s.Require().NoError(err)
					s.Require().False(voted)
				} else {
					s.Require().NoError(err)
					s.Require().NotNil(voteResp)

					// Check the vote is in the state
					breakdown, err := s.App.SponsorshipKeeper.GetValidatorBreakdown(s.Ctx, voter)
					s.Require().NoError(err)

					expectedVote := types.Vote{
						VotingPower: breakdown.TotalPower,
						Weights:     v.Weights,
					}
					actualVote := s.GetVote(v.Voter)
					s.Require().Equal(expectedVote, actualVote, "expect: %v\nactual: %v", expectedVote, actualVote)
				}
			}

			// Check events emitted
			if !tc.expectErr {
				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(types.EventVote)), len(tc.votes))
			}

			// Check the distribution is correct
			distr := s.GetDistribution()
			err = distr.Validate()
			s.Require().NoError(err)
			s.Require().True(tc.expectedDistr.Equal(distr), "expect: %v\nactual: %v", tc.expectedDistr, distr)
		})
	}
}

func (s *KeeperTestSuite) TestMsgRevokeVote() {
	addr := apptesting.CreateRandomAccounts(3)

	testCases := []struct {
		name          string
		numGauges     int // number of initial gauges, IDs fall between [1; numGauges]
		vote          types.MsgVote
		expectErr     bool
		errorContains string
	}{
		{
			name:      "Valid",
			numGauges: 2,
			vote: types.MsgVote{
				Voter: addr[0].String(),
				Weights: []types.GaugeWeight{
					{GaugeId: 1, Weight: math.NewInt(20)},
					{GaugeId: 2, Weight: math.NewInt(30)},
				},
			},
			expectErr:     false,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Create expected num of gauges
			s.CreateGauges(tc.numGauges)

			// Delegate to the validator
			val := s.CreateValidator()
			delAddr, err := sdk.AccAddressFromBech32(tc.vote.Voter)
			initial := sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))
			apptesting.FundAccount(s.App, s.Ctx, delAddr, sdk.NewCoins(initial))
			_ = s.Delegate(delAddr, val.GetOperator(), initial)

			// Set tne initial distribution
			err = s.App.SponsorshipKeeper.SaveDistribution(s.Ctx, types.NewDistribution())
			s.Require().NoError(err)

			// Cast the vote
			voteResp, err := s.msgServer.Vote(s.Ctx, &tc.vote)
			s.Require().NoError(err)
			s.Require().NotNil(voteResp)

			// Revoke the vote
			revokeResp, err := s.msgServer.RevokeVote(s.Ctx, &types.MsgRevokeVote{Voter: tc.vote.Voter})
			s.Require().NoError(err)
			s.Require().NotNil(revokeResp)

			// Check events emitted
			s.AssertEventEmitted(s.Ctx, proto.MessageName(new(types.EventRevokeVote)), 1)

			// Check the distribution is correct (empty as the initial)
			distr := s.GetDistribution()
			err = distr.Validate()
			s.Require().NoError(err)
			s.Require().True(distr.Equal(types.NewDistribution()))
		})
	}
}
