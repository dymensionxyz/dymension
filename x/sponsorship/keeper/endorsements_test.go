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
			RollappId:    rollappID,
			EpochRewards: nil,
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
		eGauge := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement).Endorsement
		s.Require().Truef(eGauge.EpochRewards.Equal(event.epochRewards), "exp %s\ngot %s", event.epochRewards, eGauge.EpochRewards)

		endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, eGauge.RollappId)
		s.Require().NoError(err)
		s.Require().True(endorsement.TotalShares.Equal(event.totalShares), "exp %s\ngot %s", event.totalShares, endorsement.TotalShares)
		s.Require().True(endorsement.EpochShares.Equal(event.epochShares), "exp %s\ngot %s", event.epochShares, endorsement.EpochShares)

		for addr, e := range event.endorsers {
			sdkAddr := sdk.MustAccAddressFromBech32(addr)

			canClaim, err := s.App.SponsorshipKeeper.CanClaim(s.Ctx, sdkAddr)
			s.Require().NoError(err)
			s.Require().Equal(e.canClaim, canClaim)

			if canClaim {
				result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, sdkAddr, endorsementGaugeID)
				s.Require().NoError(err)
				s.Require().Truef(result.Rewards.Equal(e.claimableBalance), "exp %s\ngot %s", e.claimableBalance, result.Rewards)
			}
		}
	}
}

// TestClaim_LazyModel tests the Claim function with the lazy accumulator model.
func (s *KeeperTestSuite) TestClaim_LazyModel() {
	ctx := s.Ctx
	sponsorshipKeeper := s.App.SponsorshipKeeper
	incentivesKeeper := s.App.IncentivesKeeper
	bankKeeper := s.App.BankKeeper

	// Setup
	rollappID := "testrollapp_claim"
	baseGaugeID := uint64(100) // Using a higher base ID to avoid collision with other tests
	endorsementGaugeID := baseGaugeID + 1
	rollappGaugeID := baseGaugeID + 2 // The gauge users vote on for rollapp power

	userAddr := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(userAddr, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000000)))) // Fund user for gas or other potential costs

	// Test cases
	defaultGaugeCoins := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000)))
	defaultUserPower := math.NewInt(100)
	defaultTotalShares := math.NewInt(1000)
	expectedRewardsForDefaultCase := sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(1000))) // (100/1000) * 10000

	initialUserBalance := bankKeeper.GetAllBalances(ctx, userAddr)

	tests := []struct {
		name                 string
		setup                func(testCtx sdk.Context) // Function to setup specific conditions for the test
		gaugeIDToClaim       uint64
		claimer              sdk.AccAddress
		expectError          bool
		expectedErrorMessage string
		verifyPostClaim      func(testCtx sdk.Context, initialBalance sdk.Coins) // Function to verify state after claim attempt
	}{
		{
			name: "Successful claim",
			setup: func(testCtx sdk.Context) {
				// Setup gauge in incentives keeper
				endorsementGauge := incentivestypes.Gauge{
					Id:          endorsementGaugeID,
					IsPerpetual: false,
					DistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
					Coins:       defaultGaugeCoins,
					StartTime:   testCtx.BlockTime().Add(-time.Hour),
				}
				err := incentivesKeeper.SetGauge(testCtx, &endorsementGauge)
				s.Require().NoError(err)

				// Setup endorsement in sponsorship keeper
				sponsorshipKeeper.SetEndorsement(testCtx, types.Endorsement{
					RollappId:      rollappID,
					RollappGaugeId: rollappGaugeID,
					TotalShares:    defaultTotalShares,
					EpochShares:    defaultTotalShares,
				})
				// Setup vote for the user
				sponsorshipKeeper.SetVote(testCtx, types.Vote{
					Voter: userAddr.String(),
					Weights: []types.GaugeWeight{{GaugeId: rollappGaugeID, Weight: defaultUserPower}},
					VotingPower: defaultUserPower,
				})
				// Ensure user can claim (not blacklisted)
				sponsorshipKeeper.DeleteClaimBlacklist(testCtx, userAddr)
			},
			gaugeIDToClaim: endorsementGaugeID,
			claimer:        userAddr,
			expectError:    false,
			verifyPostClaim: func(testCtx sdk.Context, initialBalance sdk.Coins) {
				finalBalance := bankKeeper.GetAllBalances(testCtx, userAddr)
				s.Require().True(finalBalance.Sub(initialBalance...).IsEqual(expectedRewardsForDefaultCase), "User balance should increase by claimed rewards. Diff: %s", finalBalance.Sub(initialBalance...))
				
				// Verify user is blacklisted
				canClaim, err := sponsorshipKeeper.CanClaim(testCtx, userAddr)
				s.Require().NoError(err)
				s.Require().False(canClaim, "User should be blacklisted after a successful claim")

				// Verify gauge's DistributedCoins are updated
				gauge, err := incentivesKeeper.GetGaugeByID(testCtx, endorsementGaugeID)
				s.Require().NoError(err)
				s.Require().True(gauge.DistributedCoins.Equal(expectedRewardsForDefaultCase), "Gauge DistributedCoins should be updated")
			},
		},
		{
			name: "Claim not allowed - blacklisted",
			setup: func(testCtx sdk.Context) {
				// Setup as per successful claim initially
				endorsementGauge := incentivestypes.Gauge{
					Id:          endorsementGaugeID,
					IsPerpetual: false,
					DistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
					Coins:       defaultGaugeCoins,
					StartTime:   testCtx.BlockTime().Add(-time.Hour),
				}
				err := incentivesKeeper.SetGauge(testCtx, &endorsementGauge)
				s.Require().NoError(err)
				sponsorshipKeeper.SetEndorsement(testCtx, types.Endorsement{RollappId: rollappID, RollappGaugeId: rollappGaugeID, TotalShares: defaultTotalShares, EpochShares: defaultTotalShares})
				sponsorshipKeeper.SetVote(testCtx, types.Vote{Voter: userAddr.String(), Weights: []types.GaugeWeight{{GaugeId: rollappGaugeID, Weight: defaultUserPower}}, VotingPower: defaultUserPower})
				
				// Blacklist the user
				err = sponsorshipKeeper.BlacklistClaim(testCtx, userAddr)
				s.Require().NoError(err)
			},
			gaugeIDToClaim:       endorsementGaugeID,
			claimer:              userAddr,
			expectError:          true,
			expectedErrorMessage: "user is not allowed to claim",
			verifyPostClaim: func(testCtx sdk.Context, initialBalance sdk.Coins) {
				finalBalance := bankKeeper.GetAllBalances(testCtx, userAddr)
				s.Require().True(finalBalance.IsEqual(initialBalance), "User balance should not change if claim fails due to blacklist")
			},
		},
		{
			name: "EstimateClaim fails - gauge not found",
			setup: func(testCtx sdk.Context) {
				// No gauge setup, so EstimateClaim will fail
				sponsorshipKeeper.SetEndorsement(testCtx, types.Endorsement{RollappId: rollappID, RollappGaugeId: rollappGaugeID, TotalShares: defaultTotalShares, EpochShares: defaultTotalShares})
				sponsorshipKeeper.SetVote(testCtx, types.Vote{Voter: userAddr.String(), Weights: []types.GaugeWeight{{GaugeId: rollappGaugeID, Weight: defaultUserPower}}, VotingPower: defaultUserPower})
				sponsorshipKeeper.DeleteClaimBlacklist(testCtx, userAddr)
			},
			gaugeIDToClaim:       endorsementGaugeID + 10, // Non-existent gauge ID
			claimer:              userAddr,
			expectError:          true,
			expectedErrorMessage: "estimate claim: get gauge by ID", // Error from GetGaugeByID
			verifyPostClaim: func(testCtx sdk.Context, initialBalance sdk.Coins) {
				finalBalance := bankKeeper.GetAllBalances(testCtx, userAddr)
				s.Require().True(finalBalance.IsEqual(initialBalance), "User balance should not change if EstimateClaim fails")
			},
		},
		{
			name: "Claim with zero rewards available (gauge.Coins is zero)",
			setup: func(testCtx sdk.Context) {
				endorsementGauge := incentivestypes.Gauge{
					Id:          endorsementGaugeID,
					IsPerpetual: false,
					DistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
					Coins:       sdk.NewCoins(), // Zero coins in gauge
					StartTime:   testCtx.BlockTime().Add(-time.Hour),
				}
				err := incentivesKeeper.SetGauge(testCtx, &endorsementGauge)
				s.Require().NoError(err)
				sponsorshipKeeper.SetEndorsement(testCtx, types.Endorsement{RollappId: rollappID, RollappGaugeId: rollappGaugeID, TotalShares: defaultTotalShares, EpochShares: defaultTotalShares})
				sponsorshipKeeper.SetVote(testCtx, types.Vote{Voter: userAddr.String(), Weights: []types.GaugeWeight{{GaugeId: rollappGaugeID, Weight: defaultUserPower}}, VotingPower: defaultUserPower})
				sponsorshipKeeper.DeleteClaimBlacklist(testCtx, userAddr)
			},
			gaugeIDToClaim: endorsementGaugeID,
			claimer:        userAddr,
			expectError:    false, // Claiming zero rewards is not an error itself
			verifyPostClaim: func(testCtx sdk.Context, initialBalance sdk.Coins) {
				finalBalance := bankKeeper.GetAllBalances(testCtx, userAddr)
				s.Require().True(finalBalance.IsEqual(initialBalance), "User balance should not change if zero rewards are claimed")
				
				// User should still be blacklisted as the claim attempt was "successful" (even if for 0 rewards)
				canClaim, err := sponsorshipKeeper.CanClaim(testCtx, userAddr)
				s.Require().NoError(err)
				s.Require().False(canClaim, "User should be blacklisted even after claiming zero rewards")

				gauge, err := incentivesKeeper.GetGaugeByID(testCtx, endorsementGaugeID)
				s.Require().NoError(err)
				s.Require().True(gauge.DistributedCoins.IsZero(), "Gauge DistributedCoins should be zero if zero rewards claimed")
			},
		},
		{
			name: "Claim when user has no power for the specific rollapp (EstimateClaim returns error)",
			setup: func(testCtx sdk.Context) {
				endorsementGauge := incentivestypes.Gauge{
					Id:          endorsementGaugeID,
					IsPerpetual: false,
					DistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
					Coins:       defaultGaugeCoins,
					StartTime:   testCtx.BlockTime().Add(-time.Hour),
				}
				err := incentivesKeeper.SetGauge(testCtx, &endorsementGauge)
				s.Require().NoError(err)
				sponsorshipKeeper.SetEndorsement(testCtx, types.Endorsement{RollappId: rollappID, RollappGaugeId: rollappGaugeID, TotalShares: defaultTotalShares, EpochShares: defaultTotalShares})
				// User vote does not include power for rollappGaugeID, or power is zero
				sponsorshipKeeper.SetVote(testCtx, types.Vote{
					Voter: userAddr.String(),
					Weights: []types.GaugeWeight{{GaugeId: rollappGaugeID + 1, Weight: defaultUserPower}}, // Vote for a different gauge
					VotingPower: defaultUserPower,
				})
				sponsorshipKeeper.DeleteClaimBlacklist(testCtx, userAddr)
			},
			gaugeIDToClaim:       endorsementGaugeID,
			claimer:              userAddr,
			expectError:          true,
			expectedErrorMessage: "user has no endorsement power for RA gauge", // Error from EstimateClaim
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create a fresh context for each test case to ensure isolation
			// This is important if keepers modify state during setup or execution.
			// However, s.Ctx is usually reset by s.SetupTest() or similar in suites.
			// For this structure, we'll operate on a locally scoped ctx if needed, or ensure
			// suite's Ctx is clean. Let's assume s.Ctx is managed by the suite per test run.
			
			// Perform test-specific setup
			tt.setup(ctx)
			
			// Get initial balance before claim, specific to this test run's context and setup
			currentInitialBalance := bankKeeper.GetAllBalances(ctx, tt.claimer)

			// Execute Claim
			err := sponsorshipKeeper.Claim(ctx, tt.claimer, tt.gaugeIDToClaim)

			// Assertions
			if tt.expectError {
				s.Require().Error(err, "Expected an error")
				if tt.expectedErrorMessage != "" {
					s.Require().Contains(err.Error(), tt.expectedErrorMessage, "Error message mismatch")
				}
			} else {
				s.Require().NoError(err, "Expected no error")
			}

			if tt.verifyPostClaim != nil {
				tt.verifyPostClaim(ctx, currentInitialBalance)
			}
			
			// Teardown / Reset state if necessary
			// (e.g., delete gauges, endorsements, votes set during the test case)
			// This is important to prevent interference between test cases.
			// If SetGauge, SetEndorsement, SetVote overwrite, that's fine.
			// Otherwise, explicit deletion might be needed.
			// For now, assuming SetupTest in the suite handles major resets.
			// Explicitly delete the test gauge to be safe.
			_ = incentivesKeeper.DeleteGauge(ctx, endorsementGaugeID) // Ignores error if gauge didn't exist
			_ = incentivesKeeper.DeleteGauge(ctx, endorsementGaugeID+10) 
		})
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

// TestEstimateClaim_LazyModel tests the EstimateClaim function with the lazy accumulator model.
// It verifies that rewards are estimated based on gauge.Coins (total accumulated rewards)
// and the user's share of total endorsement power.
func (s *KeeperTestSuite) TestEstimateClaim_LazyModel() {
	ctx := s.Ctx
	sponsorshipKeeper := s.App.SponsorshipKeeper
	incentivesKeeper := s.App.IncentivesKeeper // Assuming this is the mock or actual keeper

	// Setup
	rollappID := "testrollapp"
	gaugeID := uint64(1)
	userAddr := apptesting.CreateRandomAccounts(1)[0]

	// 1. Create Endorsement and Vote
	// Total shares for the endorsement (e.g., sum of all users' voting power for this rollapp)
	totalEndorsementShares := math.NewInt(1000)
	sponsorshipKeeper.SetEndorsement(ctx, types.Endorsement{
		RollappId:      rollappID,
		RollappGaugeId: 2, // ID of the gauge users vote on to get power for this rollapp
		TotalShares:    totalEndorsementShares, // Represents total shares for the rollapp
		EpochShares:    totalEndorsementShares, // Assuming EpochShares is repurposed or aligned with TotalShares
	})

	// User's voting power for the specific rollapp gauge
	userPowerForRollapp := math.NewInt(100)
	sponsorshipKeeper.SetVote(ctx, types.Vote{
		Voter: userAddr.String(),
		Weights: []types.GaugeWeight{
			{GaugeId: 2, Weight: userPowerForRollapp}, // User votes on rollapp gauge ID 2
		},
		VotingPower: userPowerForRollapp, // Total voting power of user, assume it's all for this for simplicity
	})

	// 2. Setup Mock Incentives Keeper
	// Define test cases for different gauge states
	testCases := []struct {
		name                 string
		gaugeCoins           sdk.Coins // Total accumulated rewards in the endorsement gauge
		userPower            math.Int  // User's power for the specific rollapp
		totalShares          math.Int  // Total shares for the endorsement
		expectedRewards      sdk.Coins
		expectError          bool
		setupVote            bool // To test cases where user might not have voted
		setupEndorsement     bool // To test cases where endorsement might not exist
		setupGauge           bool // To test cases where gauge might not exist
		mockGaugeDistributeTo *incentivestypes.Gauge_Endorsement // To simulate gauge type
	}{
		{
			name:            "Regular claim",
			gaugeCoins:      sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:       userPowerForRollapp,
			totalShares:     totalEndorsementShares,
			expectedRewards: sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(1000))), // (100/1000) * 10000
			expectError:     false,
			setupVote:       true,
			setupEndorsement:true,
			setupGauge:      true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
		{
			name:            "Zero gauge coins",
			gaugeCoins:      sdk.NewCoins(),
			userPower:       userPowerForRollapp,
			totalShares:     totalEndorsementShares,
			expectedRewards: sdk.NewCoins(),
			expectError:     false,
			setupVote:       true,
			setupEndorsement:true,
			setupGauge:      true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
		{
			name:            "Zero user power",
			gaugeCoins:      sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:       math.NewInt(0),
			totalShares:     totalEndorsementShares,
			// Expect error because EstimateClaim returns error if power is zero
			expectedRewards: sdk.NewCoins(),
			expectError:     true, // "user does not endorse respective RA gauge"
			setupVote:       true, // Vote object exists, but power for gauge is zero
			setupEndorsement:true,
			setupGauge:      true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
		{
			name:            "Zero total shares (division by zero protection)",
			gaugeCoins:      sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:       userPowerForRollapp,
			totalShares:     math.NewInt(0),
			expectedRewards: sdk.NewCoins(), // No panic, returns empty coins
			expectError:     false,          // Code returns empty coins, not error
			setupVote:       true,
			setupEndorsement:true,
			setupGauge:      true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
		{
			name:        "Gauge not found",
			gaugeCoins:  sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:   userPowerForRollapp,
			totalShares: totalEndorsementShares,
			expectError: true,
			setupVote:   true,
			setupEndorsement: true,
			setupGauge:  false, // Trigger gauge not found
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
		{
			name:        "Gauge is not an endorsement gauge",
			gaugeCoins:  sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:   userPowerForRollapp,
			totalShares: totalEndorsementShares,
			expectError: true,
			setupVote:   true,
			setupEndorsement: true,
			setupGauge:  true,
			mockGaugeDistributeTo: nil, // Simulate not an endorsement gauge or incorrect type
		},
		{
			name:        "Endorsement not found",
			gaugeCoins:  sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:   userPowerForRollapp,
			totalShares: totalEndorsementShares,
			expectError: true,
			setupVote:   true,
			setupEndorsement: false, // Trigger endorsement not found
			setupGauge:  true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: "otherRollappID"}}, // point to a different rollappID
		},
		{
			name:        "Vote not found",
			gaugeCoins:  sdk.NewCoins(sdk.NewCoin("adym", math.NewInt(10000))),
			userPower:   userPowerForRollapp, // This won't be used as vote isn't found
			totalShares: totalEndorsementShares,
			expectError: true,
			setupVote:   false, // Trigger vote not found
			setupEndorsement: true,
			setupGauge:  true,
			mockGaugeDistributeTo: &incentivestypes.Gauge_Endorsement{Endorsement: &incentivestypes.EndorsementGauge{RollappId: rollappID}},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset context for each test case if needed, or ensure mocks are clean
			// For this test, we mostly mock direct calls based on inputs.
			// The suite's Ctx might be okay if not heavily modified by other tests in parallel.

			// Setup mocks for this specific test case
			if tc.setupGauge {
				mockGauge := incentivestypes.Gauge{
					Id:          gaugeID,
					IsPerpetual: false,
					DistributeTo: tc.mockGaugeDistributeTo, // This needs to be the correct oneof type
					Coins:       tc.gaugeCoins, // Total accumulated rewards
					StartTime:   ctx.BlockTime().Add(-time.Hour),
				}
				// If mockGaugeDistributeTo is nil, we need to provide a valid but different oneof type
				if tc.mockGaugeDistributeTo == nil {
					mockGauge.DistributeTo = &incentivestypes.Gauge_Asset{Asset: &lockuptypes.QueryCondition{}}
				}

				// Mock GetGaugeByID - this is tricky if incentivesKeeper is not a mock
				// For simplicity, let's assume we can setup the keeper with this gauge
				// If not, this test would need proper mocking framework for incentivesKeeper
				err := incentivesKeeper.SetGauge(ctx, &mockGauge)
				s.Require().NoError(err, "Failed to set mock gauge")
			} else {
				// Ensure gauge does not exist or GetGaugeByID returns error
				// This might require deleting the gauge if it was set by a previous test or using a unique ID
				// For now, assume GetGaugeByID will fail if not explicitly set up.
			}

			currentEndorsement := types.Endorsement{
				RollappId:      rollappID,
				RollappGaugeId: 2,
				TotalShares:    tc.totalShares,
				EpochShares:    tc.totalShares, // Assuming EpochShares is TotalShares for lazy model
			}
			if tc.setupEndorsement {
				if tc.mockGaugeDistributeTo != nil && tc.mockGaugeDistributeTo.Endorsement != nil {
					currentEndorsement.RollappId = tc.mockGaugeDistributeTo.Endorsement.RollappId
				}
				sponsorshipKeeper.SetEndorsement(ctx, currentEndorsement)
			} else {
				// Ensure endorsement does not exist for the specific rollappID used by the gauge
				// For the "Endorsement not found" case, the gauge might point to "otherRollappID"
				// while we ensure no endorsement for "otherRollappID" exists.
				// If the gauge points to rollappID, we'd need to remove sponsorshipKeeper.GetEndorsement(ctx, rollappID)
			}
			
			currentUserVote := types.Vote{
				Voter: userAddr.String(),
				Weights: []types.GaugeWeight{
					// Power for the rollapp gauge (ID 2)
					{GaugeId: currentEndorsement.RollappGaugeId, Weight: tc.userPower},
				},
				VotingPower: tc.userPower, // Total power of user
			}
			if tc.setupVote {
				sponsorshipKeeper.SetVote(ctx, currentUserVote)
			} else {
				// Ensure vote does not exist for userAddr
				// sponsorshipKeeper.DeleteVote(ctx, userAddr.String()) // If such a method exists
			}


			// Actual call
			result, err := sponsorshipKeeper.EstimateClaim(ctx, userAddr, gaugeID)

			// Assertions
			if tc.expectError {
				s.Require().Error(err, "Expected an error for test case: %s", tc.name)
			} else {
				s.Require().NoError(err, "Expected no error for test case: %s", tc.name)
				s.Require().NotNil(result, "Expected result to be non-nil for test case: %s", tc.name)
				s.Require().True(tc.expectedRewards.Equal(result.Rewards),
					"Rewards mismatch for test case: %s. Expected %s, got %s", tc.name, tc.expectedRewards, result.Rewards)
				if tc.mockGaugeDistributeTo != nil && tc.mockGaugeDistributeTo.Endorsement != nil {
					s.Require().Equal(tc.mockGaugeDistributeTo.Endorsement.RollappId, result.RollappId, "RollappId mismatch for test case: %s", tc.name)
				}
				s.Require().True(result.EndorsedAmount.Equal(tc.userPower), "EndorsedAmount mismatch for test case: %s", tc.name)
			}

			// Cleanup: remove gauge after test to avoid interference if using actual keeper
			if tc.setupGauge {
				// incentivesKeeper.DeleteGauge(ctx, gaugeID) // If such a method exists
			}
		})
	}
}
