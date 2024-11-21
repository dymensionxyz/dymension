package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// This test is made with respect to the Cosmos SDK issue with shares calculation
// https://github.com/cosmos/cosmos-sdk/issues/11084#issuecomment-2202729511.
// The figures are from the scenario faced in Dymension testnet.
func (s *KeeperTestSuite) TestSponsorshipStakingPower() {
	// Create a new validator
	valI := s.CreateValidator()

	// Create a new delegator
	delCoin := sdk.NewCoin(s.App.StakingKeeper.BondDenom(s.Ctx), types.DYM.MulRaw(100))
	delI := s.CreateDelegator(valI.GetOperator(), delCoin)

	// Get the validator and delegator
	del, found := s.App.StakingKeeper.GetDelegation(s.Ctx, delI.GetDelegatorAddr(), delI.GetValidatorAddr())
	s.Require().True(found)
	val, found := s.App.StakingKeeper.GetValidator(s.Ctx, delI.GetValidatorAddr())
	s.Require().True(found)

	// Modify the validator and delegator shares with specific values from the scenario
	valTokens, ok := math.NewIntFromString("147832774220793166606172162")
	s.Require().True(ok)
	val.Tokens = valTokens
	val.DelegatorShares = math.LegacyMustNewDecFromStr("36367374852345403688683780.037360536022838465")
	del.Shares = math.LegacyMustNewDecFromStr("24600346603811628012.902035514609719356")

	// Save the modified validator and delegator
	s.App.StakingKeeper.SetValidator(s.Ctx, val)
	s.App.StakingKeeper.SetDelegation(s.Ctx, del)

	// Query the delegation from x/staking
	stakingQuerier := stakingkeeper.Querier{Keeper: s.App.StakingKeeper}
	resp, err := stakingQuerier.Delegation(s.Ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delI.GetDelegatorAddr().String(),
		ValidatorAddr: delI.GetValidatorAddr().String(),
	})
	s.Require().NoError(err)

	// Validate that the error is reproduced in the current SDK version
	// The valid expected amount is 100 DYM since
	// 24600346603811628012.902035514609719356 / 36367374852345403688683780.037360536022838465 * 147832774220793166606172162 == 100000000000000000000
	// But in x/staking the final float value is truncated, and 99...99,(9) becomes 99...99 (without the decimal part).
	expectedAmt, ok := math.NewIntFromString("99999999999999999999")
	s.Require().True(ok)
	s.Require().True(resp.DelegationResponse.Balance.Amount.Equal(expectedAmt))

	// Now compare the values with the sponsorship module
	s.CreateGauges(1)
	s.Vote(types.MsgVote{
		Voter: del.GetDelegatorAddr().String(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: types.DYM.MulRaw(50)},
		},
	})

	// Staking power should be the same as the x/staking module
	s.AssertVoted(del.GetDelegatorAddr())
	vote := s.GetVote(del.GetDelegatorAddr().String())
	s.Require().True(vote.VotingPower.Equal(expectedAmt))
}

func (s *KeeperTestSuite) TestStakingPowerTruncation() {
	tokens, ok := math.NewIntFromString("147832774220793166606172162")
	s.Require().True(ok)
	delegatorShares := math.LegacyMustNewDecFromStr("36367374852345403688683780.037360536022838465")
	shares := math.LegacyMustNewDecFromStr("24600346603811628012.902035514609719356")

	val1 := shares.MulInt(tokens).Quo(delegatorShares)
	val2 := shares.MulInt(tokens).Quo(delegatorShares).TruncateInt()
	val3 := shares.MulInt(tokens).Quo(delegatorShares).Ceil().TruncateInt()

	s.T().Log(val1) // 99999999999999999999.99999999999999999
	s.T().Log(val2) // 99999999999999999999
	s.T().Log(val3) // 100000000000000000000
}
