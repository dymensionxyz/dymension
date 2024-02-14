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
		demandOrderFulfillmentStatus            bool
		demandOrderUnderlyingPacketStatus       commontypes.Status
		demandOrderDenom                        string
		underlyingRollappPacket                 *commontypes.RollappPacket
		expectedFulfillmentError                error
		eIBCdemandAddrBalance                   math.Int
		expectedDemandOrdefFulfillmentStatus    bool
		expectedPostCreationEventsType          string
		expectedPostCreationEventsCount         int
		expectedPostCreationEventsAttributes    []sdk.Attribute
		expectedPostFulfillmentEventsType       string
		expectedPostFulfillmentEventsCount      int
		expectedPostFulfillmentEventsAttributes []sdk.Attribute
	}{
		{
			name:                                 "Test demand order fulfillment - success",
			demandOrderPrice:                     "150",
			demandOrderFee:                       "50",
			demandOrderFulfillmentStatus:         false,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_PENDING,
			demandOrderDenom:                     sdk.DefaultBondDenom,
			underlyingRollappPacket:              rollappPacket,
			expectedFulfillmentError:             nil,
			eIBCdemandAddrBalance:                math.NewInt(1000),
			expectedDemandOrdefFulfillmentStatus: true,
			expectedPostCreationEventsType:       eibcEventType,
			expectedPostCreationEventsCount:      1,
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
			name:                                 "Test demand order fulfillment - insufficient balance same denom",
			demandOrderPrice:                     "150",
			demandOrderFee:                       "50",
			demandOrderFulfillmentStatus:         false,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_PENDING,
			demandOrderDenom:                     sdk.DefaultBondDenom,
			expectedFulfillmentError:             types.ErrFullfillerInsufficientBalance,
			eIBCdemandAddrBalance:                math.NewInt(130),
			expectedDemandOrdefFulfillmentStatus: false,
			expectedPostCreationEventsType:       eibcEventType,
			expectedPostCreationEventsCount:      1,
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
		{
			name:                                 "Test demand order fulfillment - insufficient balance different denom",
			demandOrderPrice:                     "150",
			demandOrderFee:                       "50",
			demandOrderFulfillmentStatus:         false,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_PENDING,
			demandOrderDenom:                     "adym",
			expectedFulfillmentError:             types.ErrFullfillerInsufficientBalance,
			eIBCdemandAddrBalance:                math.NewInt(130),
			expectedDemandOrdefFulfillmentStatus: false,
			expectedPostCreationEventsType:       eibcEventType,
			expectedPostCreationEventsCount:      1,
			expectedPostCreationEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150adym"),
				sdk.NewAttribute(types.AttributeKeyFee, "50adym"),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 0,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150adym"),
				sdk.NewAttribute(types.AttributeKeyFee, "50adym"),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
		},
		{
			name:                                 "Test demand order fulfillment - already fulfilled",
			demandOrderPrice:                     "150",
			demandOrderFee:                       "50",
			demandOrderFulfillmentStatus:         true,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_PENDING,
			demandOrderDenom:                     sdk.DefaultBondDenom,
			expectedFulfillmentError:             types.ErrDemandAlreadyFulfilled,
			eIBCdemandAddrBalance:                math.NewInt(300),
			expectedDemandOrdefFulfillmentStatus: true,
			expectedPostCreationEventsType:       eibcEventType,
			expectedPostCreationEventsCount:      1,
			expectedPostCreationEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "true"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 0,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "true"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
		},
		{
			name:                                 "Test demand order fulfillment - status not pending",
			demandOrderPrice:                     "150",
			demandOrderFee:                       "50",
			demandOrderFulfillmentStatus:         false,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_FINALIZED,
			demandOrderDenom:                     sdk.DefaultBondDenom,
			expectedFulfillmentError:             types.ErrDemandOrderDoesNotExist,
			eIBCdemandAddrBalance:                math.NewInt(300),
			expectedDemandOrdefFulfillmentStatus: false,
			expectedPostCreationEventsType:       eibcEventType,
			expectedPostCreationEventsCount:      2,
			expectedPostCreationEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_FINALIZED.String()),
			},
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 0,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "150"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFullfilled, "false"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_FINALIZED.String()),
			},
		},
	}
	totalEventsEmitted := 0
	for _, tc := range tests {
		// Create and fund the account
		testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, tc.eIBCdemandAddrBalance)
		eibcSupplyAddr := testAddresses[0]
		eibcDemandAddr := testAddresses[1]
		// Get balances
		eibcSupplyAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcSupplyAddr, sdk.DefaultBondDenom)
		eibcDemandAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcDemandAddr, sdk.DefaultBondDenom)
		// Set the rollapp packet
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		// Create new demand order
		demandOrder, err := types.NewDemandOrder(*rollappPacket, tc.demandOrderPrice, tc.demandOrderFee, tc.demandOrderDenom, eibcSupplyAddr.String())
		suite.Require().NoError(err)
		demandOrder.IsFullfilled = tc.demandOrderFulfillmentStatus
		err = suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
		suite.Require().NoError(err)
		// Update rollapp status if needed
		if rollappPacket.Status != tc.demandOrderUnderlyingPacketStatus {
			_, err = suite.App.DelayedAckKeeper.UpdateRollappPacketWithStatus(suite.Ctx, *rollappPacket, tc.demandOrderUnderlyingPacketStatus)
			suite.Require().NoError(err)
		}
		// Validate creation events emitted
		suite.AssertEventEmitted(suite.Ctx, tc.expectedPostCreationEventsType, tc.expectedPostCreationEventsCount+totalEventsEmitted)
		totalEventsEmitted += tc.expectedPostCreationEventsCount
		lastEvent, ok := suite.FindLastEventOfType(suite.Ctx.EventManager().Events(), tc.expectedPostCreationEventsType)
		suite.Require().True(ok)
		suite.AssertAttributes(lastEvent, tc.expectedPostCreationEventsAttributes)
		// try fulfill the demand order
		demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, tc.demandOrderUnderlyingPacketStatus, demandOrder.Id)
		suite.Require().NoError(err)
		_, err = suite.msgServer.FulfillOrder(suite.Ctx, &types.MsgFulfillOrder{
			FulfillerAddress: eibcDemandAddr.String(),
			OrderId:          demandOrder.Id,
		})
		if tc.expectedFulfillmentError != nil {
			suite.Require().Error(err)
			suite.Require().ErrorIs(err, tc.expectedFulfillmentError)
		} else {
			suite.Require().NoError(err)
		}
		// Check that the demand fulfillment
		demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, tc.demandOrderUnderlyingPacketStatus, demandOrder.Id)
		suite.Require().NoError(err)
		suite.Assert().Equal(tc.expectedDemandOrdefFulfillmentStatus, demandOrder.IsFullfilled)
		// Check balances updates in case of success
		if tc.expectedFulfillmentError == nil {
			afterFulfillmentSupplyAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcSupplyAddr, sdk.DefaultBondDenom)
			afterFulfillmentDemandAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcDemandAddr, sdk.DefaultBondDenom)
			priceInt, ok := math.NewIntFromString(tc.demandOrderPrice)
			suite.Require().True(ok)
			suite.Require().Equal(eibcSupplyAddrBalance.Add(sdk.NewCoin(sdk.DefaultBondDenom, priceInt)), afterFulfillmentSupplyAddrBalance)
			suite.Require().Equal(eibcDemandAddrBalance.Sub(sdk.NewCoin(sdk.DefaultBondDenom, priceInt)), afterFulfillmentDemandAddrBalance)
		}
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
