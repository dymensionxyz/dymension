package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	dacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *KeeperTestSuite) TestCreateDemandOrderOnRecv() {
	tests := []struct {
		name          string
		memo          string
		expectedErr   bool
		expectedFee   string
		expectedPrice string //considering bridging fee of 1%
	}{
		{
			name:          "fee by memo - create demand order",
			memo:          `{"eibc":{"fee":"100"}}`,
			expectedErr:   false,
			expectedFee:   "100",
			expectedPrice: "890",
		},
		{
			name:          "empty memo - create demand order",
			memo:          "",
			expectedErr:   false,
			expectedPrice: "990",
		},
		{
			name:          "memo w/o eibc - create demand order",
			memo:          `{"notEIBC":{}}`,
			expectedErr:   false,
			expectedPrice: "990",
		},
		{
			name:        "bad memo - fail",
			memo:        "bad",
			expectedErr: true,
		},
		{
			name:        "PFM memo - fail",
			memo:        `{"forward":{}}`,
			expectedErr: true,
		},
	}

	// set 1% bridging fee
	dackParams := dacktypes.NewParams("hour", sdk.NewDecWithPrec(1, 2)) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)

	amt, _ := sdk.NewIntFromString(transferPacketData.Amount)
	bridgeFee := suite.App.DelayedAckKeeper.BridgingFeeFromAmt(suite.Ctx, amt)
	suite.Require().True(bridgeFee.IsPositive())

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// modify the memo and set rollapp packet
			transferPacketData.Memo = tt.memo
			packet = channeltypes.NewPacket(transferPacketData.GetBytes(), 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
			rollappPacket.Packet = &packet
			suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)

			// Create new demand order
			order, err := suite.App.EIBCKeeper.CreateDemandOrderOnRecv(suite.Ctx, transferPacketData, rollappPacket)
			if tt.expectedErr {
				suite.Require().Error(err)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(order)
			if tt.expectedFee != "" {
				suite.Require().Equal(tt.expectedFee, order.Fee[0].Amount.String())
				suite.Require().Equal(tt.expectedPrice, order.Price[0].Amount.String())
			} else {
				suite.Require().Len(order.Fee, 0)
				suite.Require().Equal(tt.expectedPrice, order.Price[0].Amount.String())
			}
		})
	}
}
