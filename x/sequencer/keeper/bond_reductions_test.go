package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

const (
	seq1 = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	seq2 = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
)

func TestGetMatureDecreasingBondIDs(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)

	t.Run("No mature bonds", func(t *testing.T) {
		ids := keeper.GetMatureDecreasingBondIDs(ctx, time.Now())
		require.Len(t, ids, 0)
	})

	t.Run("Mature bonds of multiple sequencers", func(t *testing.T) {
		bondReductionTime := time.Now()
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq1,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		// Not mature
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime.Add(time.Hour),
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})

		ids := keeper.GetMatureDecreasingBondIDs(ctx, bondReductionTime)
		require.Len(t, ids, 2)
	})
}

func TestGetBondReductionsBySequencer(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)

	t.Run("No bond reductions", func(t *testing.T) {
		ids := keeper.GetBondReductionsBySequencer(ctx, seq1)
		require.Len(t, ids, 0)
	})

	t.Run("Bond reductions of multiple sequencers", func(t *testing.T) {
		bondReductionTime := time.Now()
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq1,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime.Add(time.Hour),
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})

		ids := keeper.GetBondReductionsBySequencer(ctx, seq1)
		require.Len(t, ids, 1)

		ids = keeper.GetBondReductionsBySequencer(ctx, seq2)
		require.Len(t, ids, 2)
	})
}

func (suite *SequencerTestSuite) TestHandleBondReduction() {
	suite.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := suite.CreateDefaultRollapp()
	// Create a sequencer with bond amount of minBond + 100
	defaultSequencerAddress := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(100)), pk)
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 50),
	})
	suite.Require().NoError(err)
	expectedCompletionTime := suite.Ctx.BlockHeader().Time.Add(suite.App.SequencerKeeper.UnbondingTime(suite.Ctx))
	suite.Require().Equal(expectedCompletionTime, resp.CompletionTime)
	// Execute HandleBondReduction
	suite.Ctx = suite.Ctx.WithBlockTime(expectedCompletionTime)
	suite.App.SequencerKeeper.HandleBondReduction(suite.Ctx, suite.Ctx.BlockHeader().Time)
	// Check if the bond has been reduced
	sequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, defaultSequencerAddress)
	suite.Require().Equal(bond.AddAmount(sdk.NewInt(50)), sequencer.Tokens[0])
	// ensure the bond decresing queue is empty
	reds := suite.App.SequencerKeeper.GetMatureDecreasingBondIDs(suite.Ctx, expectedCompletionTime)
	suite.Require().Len(reds, 0)
}

func (suite *SequencerTestSuite) TestHandleBondReduction_MinBondIncrease() {
	suite.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := suite.CreateDefaultRollapp()
	// Create a sequencer with bond amount of minBond + 100
	defaultSequencerAddress := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(100)), pk)
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 50),
	})
	suite.Require().NoError(err)
	expectedCompletionTime := suite.Ctx.BlockHeader().Time.Add(suite.App.SequencerKeeper.UnbondingTime(suite.Ctx))
	suite.Require().Equal(expectedCompletionTime, resp.CompletionTime)
	curBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, sdk.MustAccAddressFromBech32(defaultSequencerAddress), bondDenom)
	suite.Require().Equal(sdk.ZeroInt(), curBalance.Amount)

	// Increase the minBond param
	params := suite.App.SequencerKeeper.GetParams(suite.Ctx)
	params.MinBond = bond.AddAmount(sdk.NewInt(60))
	suite.App.SequencerKeeper.SetParams(suite.Ctx, params)

	// Execute HandleBondReduction
	suite.Ctx = suite.Ctx.WithBlockTime(expectedCompletionTime)
	suite.App.SequencerKeeper.HandleBondReduction(suite.Ctx, suite.Ctx.BlockHeader().Time)
	// Check if the bond has been reduced - but is the same as new min bond value
	sequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, defaultSequencerAddress)
	suite.Require().Equal(bond.AddAmount(sdk.NewInt(60)), sequencer.Tokens[0])
	// ensure the bond decresing queue is empty
	reds := suite.App.SequencerKeeper.GetMatureDecreasingBondIDs(suite.Ctx, expectedCompletionTime)
	suite.Require().Len(reds, 0)
	// Ensure the bond has been refunded
	curBalance = suite.App.BankKeeper.GetBalance(suite.Ctx, sdk.MustAccAddressFromBech32(defaultSequencerAddress), bondDenom)
	suite.Require().Equal(sdk.NewInt(40), curBalance.Amount)
}
