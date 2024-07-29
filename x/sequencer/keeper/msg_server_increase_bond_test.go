package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestIncreaseBond() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)
	bondAmount := sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100)
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, sdk.MustAccAddressFromBech32(seqAddr), sdk.NewCoins(bondAmount))
	suite.Require().NoError(err)

	msg := types.MsgIncreaseBond{
		Creator: seqAddr,
		Amount:  bondAmount,
	}
	_, err = suite.msgServer.IncreaseBond(suite.Ctx, &msg)

	suite.Require().NoError(err)
	expectedBond := types.DefaultParams().MinBond.Add(bondAmount)
	seq, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().Equal(expectedBond, seq.Tokens[0])
}

func (suite *SequencerTestSuite) TestIncreaseBondInvalidSequencer() {
	suite.SetupTest()
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey1.Address())
	bondAmount := sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100)

	msg := types.MsgIncreaseBond{
		Creator: addr.String(),
		Amount:  bondAmount,
	}
	_, err := suite.msgServer.IncreaseBond(suite.Ctx, &msg)

	suite.Require().ErrorIs(types.ErrUnknownSequencer, err)
}

func (suite *SequencerTestSuite) TestIncreaseBondInvalidSequencerStatus() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)
	seq.Status = types.Unbonded
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, seq, seq.Status)

	msg := types.MsgIncreaseBond{
		Creator: seqAddr,
		Amount:  sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100),
	}
	_, err := suite.msgServer.IncreaseBond(suite.Ctx, &msg)

	suite.Require().ErrorIs(types.ErrInvalidSequencerStatus, err)
}

func (suite *SequencerTestSuite) TestIncreaseBondSequencerJailed() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)
	seq.Jailed = true
	suite.App.SequencerKeeper.UpdateSequencer(suite.Ctx, seq, seq.Status)

	msg := types.MsgIncreaseBond{
		Creator: seqAddr,
		Amount:  sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100),
	}
	_, err := suite.msgServer.IncreaseBond(suite.Ctx, &msg)

	suite.Require().ErrorIs(types.ErrSequencerJailed, err)
}

func (suite *SequencerTestSuite) TestIncreaseBondInsufficientBalance() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	_, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)

	msg := types.MsgIncreaseBond{
		Creator: seqAddr,
		Amount:  sdk.NewInt64Coin(types.DefaultParams().MinBond.Denom, 100),
	}
	_, err := suite.msgServer.IncreaseBond(suite.Ctx, &msg)

	suite.Require().ErrorContains(err, "insufficient funds")
}
