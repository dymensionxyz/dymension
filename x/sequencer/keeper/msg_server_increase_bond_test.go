package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestIncreaseBond() {
	rollappId, pk := s.CreateDefaultRollapp()
	// setup a default sequencer
	defaultSequencerAddress := s.CreateSequencer(s.Ctx, rollappId, pk)
	// setup an unbonded sequencer
	pk1 := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.CreateSequencer(s.Ctx, rollappId, pk1)
	unbondedSequencer, _ := s.App.SequencerKeeper.GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	s.App.SequencerKeeper.SetSequencer(s.Ctx, unbondedSequencer)

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
				seq, _ := s.App.SequencerKeeper.GetRealSequencer(s.Ctx, defaultSequencerAddress)
				s.Require().Equal(expectedBond, seq.Tokens[0])
			}
		})
	}
}
