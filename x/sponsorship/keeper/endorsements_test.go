package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

func (s *KeeperTestSuite) TestEndorsements() {
	addrs := apptesting.CreateRandomAccounts(2)
	_ = addrs

	type endorser struct {
		votingPower      math.Int
		canClaim         bool
		balance          sdk.Coins
		claimableBalance sdk.Coins
	}

	events := []struct {
		event                  string
		simulateEvent          func()
		epochsFilled           uint64
		accumulatedBalance     sdk.Coins
		accumulatedDistributed sdk.Coins
		claimedAmount          sdk.Coins
		gaugeBalance           sdk.Coins
		epochRewards           sdk.Coins
		totalShares            math.Int
		epochShares            math.Int
		endorsers              map[string]endorser // address -> endorser
	}{
		{},
	}

	rollappID := s.CreateDefaultRollapp()
	_ = rollappID
	// Create endorsement gauge
	var gaugeID uint64

	s.SetupTest()
	for _, event := range events {
		// No not run SetupTest() on every iteration as the next iteration
		// uses the state from the previous one
		s.Run(event.event, func() {
			event.simulateEvent()

			/** Validate the state **/

			gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeID)
			s.Require().NoError(err)
			s.Require().True(gauge.Coins.Equal(event.accumulatedBalance))
			s.Require().True(gauge.DistributedCoins.Equal(event.accumulatedDistributed))
			s.Require().True(gauge.DistributedCoins.Equal(event.accumulatedDistributed))
			eGauge := gauge.DistributeTo.(*types.Gauge_Endorsement).Endorsement
			s.Require().True(eGauge.EpochRewards.Equal(event.epochRewards))

			endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, eGauge.RollappId)
			s.Require().NoError(err)
			s.Require().True(endorsement.TotalShares.Equal(event.totalShares))
			s.Require().True(endorsement.EpochShares.Equal(event.epochShares))

			for addr, e := range event.endorsers {
				sdkAddr := sdk.MustAccAddressFromBech32(addr)

				canClaim, err := s.App.SponsorshipKeeper.CanClaim(s.Ctx, sdkAddr)
				s.Require().NoError(err)
				s.Require().Equal(e.canClaim, canClaim)

				balance := s.App.BankKeeper.GetAllBalances(s.Ctx, sdkAddr)
				s.Require().NoError(err)
				s.Require().True(balance.Equal(e.balance))

				result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, sdkAddr, gaugeID)
				s.Require().NoError(err)
				s.Require().True(result.Rewards.Equal(e.claimableBalance))
			}
		})
	}
}
