package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestMsgVote() {
	testCases := []struct {
		name          string
		delegation    sdk.Coin
		msg           types.MsgVote
		expectedDistr types.Distribution
		expectedVote  types.Vote
		expectedErr   error
		errorContains string
	}{
		{
			name:       "Valid",
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000_000)),
			msg: types.MsgVote{
				Voter:   "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft",
				Weights: nil,
			},
			expectedDistr: types.Distribution{},
			expectedVote:  types.Vote{},
			expectedErr:   nil,
			errorContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get current validator
			bondedValidators := s.App.StakingKeeper.GetBondedValidatorsByPower(s.Ctx)
			s.Require().NotEmpty(bondedValidators)
			validator := bondedValidators[0]

			// Convert voter and validator addresses from bech32
			voterAddr, err := sdk.AccAddressFromBech32(tc.msg.Voter)
			s.Require().NoError(err)
			valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
			s.Require().NoError(err)

			// Fund voter's address
			apptesting.FundAccount(s.App, s.Ctx, voterAddr, sdk.Coins{tc.delegation})

			// Delegate to the validator
			delegateHandler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgDelegate))
			delegateResp, err := delegateHandler(s.Ctx, stakingtypes.NewMsgDelegate(voterAddr, valAddr, tc.delegation))
			s.Require().NoError(err)
			s.Require().NotNil(delegateResp)

			// Cast the vote
			voteHandler := s.App.MsgServiceRouter().Handler(new(types.MsgVote))
			voteResp, err := voteHandler(s.Ctx, &tc.msg)
			s.Require().NoError(err)

			distr := s.GetDistribution()
			vote := s.GetVote(tc.msg.Voter)

			_ = voteResp
			_ = distr
			_ = vote
		})
	}
}
