package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestDecreaseBond() {
	s.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.CreateDefaultRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.CreateSequencerWithBond(s.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)
	// setup an unbonded sequencer
	unbondedPk := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.CreateSequencer(s.Ctx, rollappId, unbondedPk)
	unbondedSequencer, _ := s.App.SequencerKeeper.GetSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	s.App.SequencerKeeper.UpdateSequencerLeg(s.Ctx, &unbondedSequencer, unbondedSequencer.Status)
	// setup a jailed sequencer
	jailedPk := ed25519.GenPrivKey().PubKey()
	jailedSequencerAddress := s.CreateSequencer(s.Ctx, rollappId, jailedPk)
	jailedSequencer, _ := s.App.SequencerKeeper.GetSequencer(s.Ctx, jailedSequencerAddress)
	jailedSequencer.Jailed = true
	s.App.SequencerKeeper.UpdateSequencerLeg(s.Ctx, &jailedSequencer, jailedSequencer.Status)

	testCase := []struct {
		name        string
		msg         types.MsgDecreaseBond
		expectedErr error
	}{
		{
			name: "invalid sequencer",
			msg: types.MsgDecreaseBond{
				Creator:        "invalid_address",
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrSequencerNotFound,
		},
		{
			name: "sequencer is not bonded",
			msg: types.MsgDecreaseBond{
				Creator:        unbondedSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrInvalidSequencerStatus,
		},
		{
			name: "decreased bond value to less than minimum bond value",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 100),
			},
			expectedErr: types.ErrInsufficientBond,
		},
		{
			name: "trying to decrease more bond than they have tokens bonded",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: bond.AddAmount(sdk.NewInt(30)),
			},
			expectedErr: types.ErrInsufficientBond,
		},
		{
			name: "valid decrease bond",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.DecreaseBond(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				expectedCompletionTime := s.Ctx.BlockHeader().Time.Add(s.App.SequencerKeeper.UnbondingTime(s.Ctx))
				s.Require().Equal(expectedCompletionTime, resp.CompletionTime)
				// check if the unbonding is set correctly
				bondReductionIDs := s.App.SequencerKeeper.GetMatureDecreasingBondIDs(s.Ctx, expectedCompletionTime)
				s.Require().Len(bondReductionIDs, 1)
				bondReduction, found := s.App.SequencerKeeper.GetBondReduction(s.Ctx, bondReductionIDs[0])
				s.Require().True(found)
				s.Require().Equal(tc.msg.Creator, bondReduction.SequencerAddress)
				s.Require().Equal(tc.msg.DecreaseAmount, bondReduction.DecreaseBondAmount)
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBond_BondDecreaseInProgress() {
	s.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.CreateDefaultRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.CreateSequencerWithBond(s.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)
	// decrease the bond of the sequencer
	_, err := s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should be fine as still not below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should err as below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().ErrorIs(err, types.ErrInsufficientBond)
}
