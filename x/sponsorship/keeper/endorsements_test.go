package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestEndorsements() {
	s.T().Skip("skip until tested")

	// Begin the very first epoch
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	/** Create a rollapp and an additional gauge **/

	// This automatically creates a rollapp gauge with ID 1
	rollappID := s.CreateDefaultRollapp()

	// Additional gauge with ID 2
	s.CreateGauges(1)

	/** Create an endorsement gauge **/

	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	dym100 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	s.FundAcc(gaugeCreator, dym100)

	// Gauge for 100 DYM and 10 epochs. This gauge has ID 3.
	endorsementGaugeID, err := s.App.IncentivesKeeper.CreateEndorsementGauge(
		s.Ctx,
		false,
		gaugeCreator,
		dym100,
		incentivestypes.EndorsementGauge{
			RollappId: rollappID,
		},
		s.Ctx.BlockTime(),
		10,
	)
	s.Require().NoError(err)

	/** Create a validator and two delegators **/

	val := s.CreateValidator()
	valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
	s.Require().NoError(err)

	// user1 delegates 40 DYM
	initial1 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))
	del1 := s.CreateDelegator(valAddr, initial1)

	// user2 delegates 80 DYM
	initial2 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(160))
	del2 := s.CreateDelegator(valAddr, initial2)

	/** Cast user votes **/

	// User1 votes 50% on each gauge, so 40 DYM goes to the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del1.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			// 50% for the rollapp gauge. 50 is a percentage, not a real DYM!
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(50)},
			// 50% for the additional gauge. 50 is a percentage, not a real DYM!
			{GaugeId: 2, Weight: commontypes.DYM.MulRaw(50)},
		},
	})

	// User2 votes 50% on each gauge, so 80 DYM goes to the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			// 50% for the rollapp gauge. 50 is a percentage, not a real DYM!
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(50)},
			// 50% for the additional gauge. 50 is a percentage, not a real DYM!
			{GaugeId: 2, Weight: commontypes.DYM.MulRaw(50)},
		},
	})

	/** Test cases **/

	params := s.App.IncentivesKeeper.GetParams(s.Ctx)
	_ = params

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
		gaugeBalance           sdk.Coins
		epochRewards           sdk.Coins
		totalShares            math.Int
		epochShares            math.Int
		endorsers              map[string]endorser // address -> endorser
	}{
		{
			event: "Epoch 1 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:           1,
			accumulatedBalance:     dym100,
			accumulatedDistributed: sdk.NewCoins(),
			gaugeBalance:           dym100,
			epochRewards:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10))),
			totalShares:            types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:            types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(3_333_333_333_333_333_333)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(6_666_666_666_666_666_666)),
				},
			},
		},
		{
			event: "User1 claims",
			simulateEvent: func() {
				_, err = s.msgServer.ClaimRewards(s.Ctx, &types.MsgClaimRewards{
					Sender:  del1.GetDelegatorAddr(),
					GaugeId: 3,
				})
				s.Require().NoError(err)
			},
			epochsFilled:           1,
			accumulatedBalance:     dym100,
			accumulatedDistributed: sdk.NewCoins(s.adym(3_333_333_333_333_333_333)),
			gaugeBalance:           dym100.Sub(s.adym(3_333_333_333_333_333_333)),
			epochRewards:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10))),
			totalShares:            types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:            types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(6_666_666_666_666_666_666)),
				},
			},
		},
		{
			event: "User2 claims",
			simulateEvent: func() {
				_, err = s.msgServer.ClaimRewards(s.Ctx, &types.MsgClaimRewards{
					Sender:  del2.GetDelegatorAddr(),
					GaugeId: 3,
				})
				s.Require().NoError(err)
			},
			epochsFilled:       1,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
			},
		},
		{
			event: "Epoch 2 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:       2,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(3_333_333_333_333_333_333)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(6_666_666_666_666_666_666)),
				},
			},
		},
		{
			event: "User1 claims",
			simulateEvent: func() {
				_, err = s.msgServer.ClaimRewards(s.Ctx, &types.MsgClaimRewards{
					Sender:  del1.GetDelegatorAddr(),
					GaugeId: 3,
				})
				s.Require().NoError(err)
			},
			epochsFilled:       2,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(6_666_666_666_666_666_666)),
				},
			},
		},
		{
			event: "Epoch 3 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:       3,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(math.NewIntFromUint64(10_833_333_333_333_333_333)))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(3_611_111_111_111_111_111)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(7_222_222_222_222_222_222)),
				},
			},
		},
		{
			event: "User2 claims",
			simulateEvent: func() {
				_, err = s.msgServer.ClaimRewards(s.Ctx, &types.MsgClaimRewards{
					Sender:  del2.GetDelegatorAddr(),
					GaugeId: 3,
				})
				s.Require().NoError(err)
			},
			epochsFilled:       3,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(math.NewIntFromUint64(10_833_333_333_333_333_333)))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(3_611_111_111_111_111_111)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
			},
		},
		{
			event: "Epoch 4 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:       4,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(math.NewIntFromUint64(11_349_206_349_206_349_206)))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(3_783_068_783_068_783_068)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(7_566_137_566_137_566_137)),
				},
			},
		},
		{
			event: "Epoch 5 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:       5,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(math.NewIntFromUint64(13_240_740_740_740_740_741)))),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(4_413_580_246_913_580_247)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(8_827_160_493_827_160_494)),
				},
			},
		},
		{
			event: "User1 un-endorses",
			simulateEvent: func() {
				_, err = s.msgServer.ClaimRewards(s.Ctx, &types.MsgClaimRewards{
					Sender:  del1.GetDelegatorAddr(),
					GaugeId: 3,
				})
				s.Require().NoError(err)

				_, err = s.msgServer.RevokeVote(s.Ctx, &types.MsgRevokeVote{
					Voter: del1.GetDelegatorAddr(),
				})
				s.Require().NoError(err)
			},
			epochsFilled:       5,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			epochRewards: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(math.NewIntFromUint64(13_240_740_740_740_740_741)))),
			totalShares:  types.DYM.MulRaw(80),  // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(8_827_160_493_827_160_494)),
				},
			},
		},
		{
			event: "Epoch 6 start",
			simulateEvent: func() {
				s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
			},
			epochsFilled:       6,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			epochRewards: sdk.NewCoins(s.adym(15_006_172_839_506_172_839)),
			totalShares:  types.DYM.MulRaw(80), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(80), // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(15_006_172_839_506_172_839)),
				},
			},
		},
		{
			event: "User1 endorses",
			simulateEvent: func() {
				// User1 votes 50% on each gauge, so 40 DYM goes to the rollapp gauge
				s.Vote(types.MsgVote{
					Voter: del1.GetDelegatorAddr(),
					Weights: []types.GaugeWeight{
						// 50% for the rollapp gauge. 50 is a percentage, not a real DYM!
						{GaugeId: 1, Weight: commontypes.DYM.MulRaw(50)},
						// 50% for the additional gauge. 50 is a percentage, not a real DYM!
						{GaugeId: 2, Weight: commontypes.DYM.MulRaw(50)},
					},
				})
			},
			epochsFilled:       6,
			accumulatedBalance: dym100,
			accumulatedDistributed: sdk.NewCoins(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			gaugeBalance: dym100.Sub(
				s.adym(3_333_333_333_333_333_333).
					Add(s.adym(6_666_666_666_666_666_666)).
					Add(s.adym(3_333_333_333_333_333_333)).
					Add(s.adym(7_222_222_222_222_222_222)).
					Add(s.adym(4_413_580_246_913_580_247)),
			),
			epochRewards: sdk.NewCoins(s.adym(15_006_172_839_506_172_839)),
			totalShares:  types.DYM.MulRaw(120), // here is DYM is just a number, not a real DYM
			epochShares:  types.DYM.MulRaw(80),  // here is DYM is just a number, not a real DYM
			endorsers: map[string]endorser{
				del1.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(40),
					canClaim:         false,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))),
					claimableBalance: sdk.NewCoins(s.adym(0)),
				},
				del2.GetDelegatorAddr(): {
					votingPower:      commontypes.DYM.MulRaw(80),
					canClaim:         true,
					balance:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(80))),
					claimableBalance: sdk.NewCoins(s.adym(15_006_172_839_506_172_839)),
				},
			},
		},
	}

	for _, event := range events {
		// No need to run SetupTest() on every iteration as the next iteration
		// uses the state from the previous one

		event.simulateEvent()

		/** Validate the state **/

		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, endorsementGaugeID)
		s.Require().NoError(err)
		s.Require().Equal(event.epochsFilled, gauge.FilledEpochs)
		s.Require().True(gauge.Coins.Equal(event.accumulatedBalance))
		s.Require().True(gauge.DistributedCoins.Equal(event.accumulatedDistributed))
		s.Require().True(gauge.Coins.Sub(gauge.DistributedCoins...).Equal(event.gaugeBalance))
		//eGauge := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement).Endorsement
		//s.Require().Truef(eGauge.EpochRewards.Equal(event.epochRewards), "exp %s\ngot %s", event.epochRewards, eGauge.EpochRewards)

		//endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, eGauge.RollappId)
		//s.Require().NoError(err)
		//s.Require().True(endorsement.TotalShares.Equal(event.totalShares), "exp %s\ngot %s", event.totalShares, endorsement.TotalShares)
		//s.Require().True(endorsement.EpochShares.Equal(event.epochShares), "exp %s\ngot %s", event.epochShares, endorsement.EpochShares)

		for addr, e := range event.endorsers {
			sdkAddr := sdk.MustAccAddressFromBech32(addr)

			//canClaim, err := s.App.SponsorshipKeeper.CanClaim(s.Ctx, sdkAddr)
			//s.Require().NoError(err)
			//s.Require().Equal(e.canClaim, canClaim)

			//if canClaim {
			result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, sdkAddr, endorsementGaugeID)
			s.Require().NoError(err)
			s.Require().Truef(result.Rewards.Equal(e.claimableBalance), "exp %s\ngot %s", e.claimableBalance, result.Rewards)
		}
	}
}

func (s *KeeperTestSuite) adym(number uint64) sdk.Coin {
	v := math.NewIntFromUint64(number)
	return sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(v))
}

func (s *KeeperTestSuite) adym1(number string) sdk.Coin {
	v, ok := math.NewIntFromString(number)
	s.Require().True(ok)
	return sdk.NewCoin(sdk.DefaultBondDenom, commontypes.ADYM.Mul(v))
}

func (s *KeeperTestSuite) BeginEpoch(epochID string) {
	info := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochID)
	s.Ctx = s.Ctx.WithBlockTime(info.CurrentEpochStartTime.Add(info.Duration).Add(time.Minute))
	s.App.EpochsKeeper.BeginBlocker(s.Ctx)
}
