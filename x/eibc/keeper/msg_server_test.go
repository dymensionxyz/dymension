package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestMsgFulfillOrder() {
	tests := []struct {
		name                                    string
		demandOrderPacketKey                    string
		demandOrderPrice                        string
		demandOrderFee                          string
		underlyingRollappPacket                 *commontypes.RollappPacket
		isFulfillSuccesfull                     bool
		isDemandOrderFulfilled                  bool
		eIBCdemandAddrBalance                   math.Int
		expectedPostCreationEventsType          string
		expectedPostCreationEventsCount         int
		expectedPostCreationEventsAttributes    []sdk.Attribute
		expectedPostFulfillmentEventsType       string
		expectedPostFulfillmentEventsCount      int
		expectedPostFulfillmentEventsAttributes []sdk.Attribute
	}{
		{
			name:                            "Test successful demand order fulfillment",
			demandOrderPrice:                "150",
			demandOrderFee:                  "50",
			underlyingRollappPacket:         rollappPacket,
			isFulfillSuccesfull:             true,
			isDemandOrderFulfilled:          true,
			eIBCdemandAddrBalance:           math.NewInt(1000),
			expectedPostCreationEventsType:  eibcEventType,
			expectedPostCreationEventsCount: 1,
			expectedPostCreationEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 1,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "true"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
		},
		{
			name:                            "Test demand order fulfillment insufficient balance",
			demandOrderPrice:                "150",
			demandOrderFee:                  "50",
			isFulfillSuccesfull:             false,
			isDemandOrderFulfilled:          false,
			eIBCdemandAddrBalance:           math.NewInt(130),
			expectedPostCreationEventsType:  eibcEventType,
			expectedPostCreationEventsCount: 1,
			expectedPostCreationEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 0,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
		},
	}
	totalEventsEmitted := 0
	for _, tc := range tests {
		// Set the rollapp packet
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		// Create and fund the account
		testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, tc.eIBCdemandAddrBalance)
		eibcSupplyAddr := testAddresses[0]
		eibcDemandAddr := testAddresses[1]
		demandOrder, err := types.NewDemandOrder(*rollappPacket, tc.demandOrderPrice, tc.demandOrderFee, sdk.DefaultBondDenom, eibcSupplyAddr.String())
		suite.Require().NoError(err)
		err = suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
		suite.Require().NoError(err)
		// Validate creation events emitted
		suite.AssertEventEmitted(suite.Ctx, tc.expectedPostCreationEventsType, tc.expectedPostCreationEventsCount+totalEventsEmitted)
		totalEventsEmitted += tc.expectedPostCreationEventsCount
		lastEvent, ok := suite.FindLastEventOfType(suite.Ctx.EventManager().Events(), tc.expectedPostCreationEventsType)
		suite.Require().True(ok)
		suite.AssertAttributes(lastEvent, tc.expectedPostCreationEventsAttributes)
		// try fulfill the demand order
		demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, demandOrder.Id)
		suite.Require().NoError(err)
		_, err = suite.msgServer.FulfillOrder(suite.Ctx, &types.MsgFulfillOrder{
			FulfillerAddress: eibcDemandAddr.String(),
			OrderId:          demandOrder.Id,
		})
		if !tc.isFulfillSuccesfull {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}
		// Check that the demand fulfillment
		demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, demandOrder.Id)
		suite.Require().NoError(err)
		suite.Assert().Equal(tc.isDemandOrderFulfilled, demandOrder.IsFullfilled)
		// Validate events
		suite.AssertEventEmitted(suite.Ctx, tc.expectedPostFulfillmentEventsType, tc.expectedPostFulfillmentEventsCount+totalEventsEmitted)
		if tc.expectedPostFulfillmentEventsCount == 0 {
			continue
		}
		totalEventsEmitted += tc.expectedPostFulfillmentEventsCount
		lastEvent, ok = suite.FindLastEventOfType(suite.Ctx.EventManager().Events(), tc.expectedPostFulfillmentEventsType)
		suite.Require().True(ok)
		suite.AssertAttributes(lastEvent, tc.expectedPostFulfillmentEventsAttributes)
	}
}
