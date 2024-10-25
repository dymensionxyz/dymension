package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestIncreaseBond() {
	rollappId, pk := s.createRollapp()
	// setup a default sequencer
	defaultSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, pk)
	// setup an unbonded sequencer
	pk1 := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, pk1)
	unbondedSequencer, _ := s.k().GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	s.k().SetSequencer(s.Ctx, unbondedSequencer)

	// fund all the sequencers which have been setup
	bondAmount := sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100)
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, sdk.MustAccAddressFromBech32(defaultSequencerAddress), sdk.NewCoins(bondAmount))
	s.Require().NoError(err)
	err = bankutil.FundAccount(s.App.BankKeeper, s.Ctx, sdk.MustAccAddressFromBech32(unbondedSequencerAddress), sdk.NewCoins(bondAmount))
	s.Require().NoError(err)

	testCase := []struct {
		name        string
		msg         types.MsgIncreaseBond
		expectedErr error
	}{
		{
			name: "valid",
			msg: types.MsgIncreaseBond{
				Creator:   defaultSequencerAddress,
				AddAmount: bondAmount,
			},
			expectedErr: nil,
		},
		{
			name: "invalid sequencer",
			msg: types.MsgIncreaseBond{
				Creator:   sample.AccAddress(), // a random address which is not a registered sequencer
				AddAmount: bondAmount,
			},
			expectedErr: types.ErrSequencerNotFound,
		},
		{
			name: "sequencer doesn't have enough balance",
			msg: types.MsgIncreaseBond{
				Creator:   defaultSequencerAddress,
				AddAmount: sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 99999999), // very high amount which sequencer doesn't have
			},
			expectedErr: sdkerrors.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			_, err := s.msgServer.IncreaseBond(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				expectedBond := types.DefaultParams().MinBond.Add(bondAmount)
				seq, _ := s.k().GetRealSequencer(s.Ctx, defaultSequencerAddress)
				s.Require().Equal(expectedBond, seq.Tokens[0])
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBond() {
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(sdk.NewInt(20)))
	// setup an unbonded sequencer
	unbondedPk := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.createSequencerWithPk(s.Ctx, rollappId, unbondedPk)
	unbondedSequencer, _ := s.k().GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	s.k().SetSequencer(s.Ctx, unbondedSequencer)

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
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBond_BondDecreaseInProgress() {
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(sdk.NewInt(20)))
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

func (s *SequencerTestSuite) TestUnbondingNonProposer() {
	rollappId, pk := s.createRollapp()
	proposerAddr := s.createSequencerWithPk(s.Ctx, rollappId, pk)

	bondedAddr := s.CreateDefaultSequencer(s.Ctx, rollappId)
	s.Require().NotEqual(proposerAddr, bondedAddr)

	proposer := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, proposer.Address)

	/* ------------------------- unbond non proposer sequencer ------------------------ */
	bondedSeq, err := s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types.Bonded, bondedSeq.Status)

	unbondMsg := types.MsgUnbond{Creator: bondedAddr}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check sequencer operating status
	bondedSeq, err = s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types.Unbonding, bondedSeq.Status)

	s.k().UnbondAllMatureSequencers(s.Ctx, bondedSeq.UnbondTime.Add(10*time.Second))
	bondedSeq, err = s.k().GetRealSequencer(s.Ctx, bondedAddr)
	s.Require().True(found)
	s.Equal(types.Unbonded, bondedSeq.Status)

	// check proposer not changed
	proposer, ok = s.k().GetProposerLegacy(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, proposer.Address)

	// try to unbond again. already unbonded, we expect error
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().Error(err)
}

func (s *SequencerTestSuite) TestUnbondingProposer() {
	s.Ctx = s.Ctx.WithBlockHeight(10)

	rollappId, proposerAddr := s.CreateDefaultRollappAndProposer()
	_ = s.createSequencerWithPk(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: proposerAddr}
	_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)

	// check proposer still bonded and notice period started
	p := s.k().GetProposer(s.Ctx, rollappId)
	s.Require().True(ok)
	s.Equal(proposerAddr, p.Address)
	s.Equal(s.Ctx.BlockHeight(), p.UnbondRequestHeight)

	// unbonding again, we expect error as sequencer is in notice period
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().Error(err)

	// next proposer should not be set yet
	_, ok = s.k().GetNextProposer(s.Ctx, rollappId)
	s.Require().False(ok)

	// check notice period queue
	m, err := s.k().NoticeElapsedSequencers(s.Ctx, p.NoticePeriodTime.Add(-1*time.Second))
	s.Require().NoError(err)
	s.Require().Len(m, 0)
	m, err = s.k().NoticeElapsedSequencers(s.Ctx, p.NoticePeriodTime.Add(1*time.Second))
	s.Require().NoError(err)
	s.Require().Len(m, 1)
}
