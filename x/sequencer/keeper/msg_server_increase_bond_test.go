package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestIncreaseBond() {
	suite.SetupTest()
	rollappId, pk := suite.CreateDefaultRollapp()
	// setup a default sequencer
	defaultSequencerAddress := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk)
	// setup an unbonded sequencer
	pk1 := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk1)
	unbondedSequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, unbondedSequencer, unbondedSequencer.Status)
	// setup a jailed sequencer
	pk2 := ed25519.GenPrivKey().PubKey()
	jailedSequencerAddress := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk2)
	jailedSequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, jailedSequencerAddress)
	jailedSequencer.Jailed = true
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, jailedSequencer, jailedSequencer.Status)
	// fund all the sequencers which have been setup
	bondAmount := sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100)
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, sdk.MustAccAddressFromBech32(defaultSequencerAddress), sdk.NewCoins(bondAmount))
	suite.Require().NoError(err)
	err = bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, sdk.MustAccAddressFromBech32(unbondedSequencerAddress), sdk.NewCoins(bondAmount))
	suite.Require().NoError(err)
	err = bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, sdk.MustAccAddressFromBech32(jailedSequencerAddress), sdk.NewCoins(bondAmount))
	suite.Require().NoError(err)

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
			expectedErr: types.ErrUnknownSequencer,
		},
		{
			name: "invalid sequencer status",
			msg: types.MsgIncreaseBond{
				Creator:   unbondedSequencerAddress,
				AddAmount: bondAmount,
			},
			expectedErr: types.ErrInvalidSequencerStatus,
		},
		{
			name: "jailed sequencer",
			msg: types.MsgIncreaseBond{
				Creator:   jailedSequencerAddress,
				AddAmount: bondAmount,
			},
			expectedErr: types.ErrSequencerJailed,
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
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.IncreaseBond(suite.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				suite.Require().ErrorIs(err, tc.expectedErr)
			} else {
				suite.Require().NoError(err)
				expectedBond := types.DefaultParams().MinBond.Add(bondAmount)
				seq, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, defaultSequencerAddress)
				suite.Require().Equal(expectedBond, seq.Tokens[0])
			}
		})
	}
}
