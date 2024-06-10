package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	dacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestMsgUpdateDemandOrder() {
	// Create and fund the account
	testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, sdk.NewInt(100_000))
	eibcSupplyAddr := testAddresses[0]

	dackParams := dacktypes.NewParams("hour", sdk.NewDecWithPrec(1, 2)) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)

	denom := suite.App.StakingKeeper.BondDenom(suite.Ctx)
	// Set a rollapp packet with 1000 amount
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// Set the initial price and fee for total amount 1000 and 1% bridge fee
	initialPrice := sdk.NewInt(890)
	initialFee := sdk.NewInt(100)

	testCases := []struct {
		name          string
		fee           sdk.Int
		submittedBy   string
		expectError   bool
		expectedPrice sdk.Int
	}{
		{
			name:          "happy case",
			fee:           sdk.NewInt(400),
			submittedBy:   eibcSupplyAddr.String(),
			expectError:   false,
			expectedPrice: sdk.NewInt(590),
		},
		{
			name:          "happy case - zero eibc fee",
			fee:           sdk.NewInt(0),
			submittedBy:   eibcSupplyAddr.String(),
			expectError:   false,
			expectedPrice: sdk.NewInt(990),
		},
		{
			name:        "wrong owner",
			fee:         sdk.NewInt(400),
			submittedBy: testAddresses[1].String(),
			expectError: true,
		},
		{
			name:        "too high fee",
			fee:         sdk.NewInt(1001),
			submittedBy: eibcSupplyAddr.String(),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		// Create new demand order
		demandOrder := types.NewDemandOrder(*rollappPacket, initialPrice, initialFee, denom, eibcSupplyAddr.String())
		err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
		suite.Require().NoError(err)

		// try to update the demand order
		msg := types.NewMsgUpdateDemandOrder(demandOrder.Id, tc.submittedBy, tc.fee.String())
		_, err = suite.msgServer.UpdateDemandOrder(suite.Ctx, msg)
		if tc.expectError {
			suite.Require().Error(err, tc.name)
			continue
		}
		suite.Require().NoError(err, tc.name)
		// check if the demand order is updated
		updatedDemandOrder, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, rollappPacket.Status, demandOrder.Id)
		suite.Require().NoError(err, tc.name)
		suite.Assert().Equal(updatedDemandOrder.Fee.AmountOf(denom), tc.fee, tc.name)
		suite.Assert().Equal(updatedDemandOrder.Price.AmountOf(denom), tc.expectedPrice, tc.name)
	}
}
