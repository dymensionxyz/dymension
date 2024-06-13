package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestMsgFulfillOrderEvent() {
	// Create and fund the account
	testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, sdk.NewInt(1000))
	eibcSupplyAddr := testAddresses[0]
	eibcDemandAddr := testAddresses[1]
	// Set the rollapp packet
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// Create new demand order
	demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(200), math.NewIntFromUint64(50), sdk.DefaultBondDenom, eibcSupplyAddr.String())
	err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)

	tests := []struct {
		name                                    string
		fulfillmentShouldFail                   bool
		expectedPostFulfillmentEventsCount      int
		expectedPostFulfillmentEventsType       string
		expectedPostFulfillmentEventsAttributes []sdk.Attribute
	}{
		{
			name:                               "Test demand order fulfillment - success",
			expectedPostFulfillmentEventsType:  eibcEventType,
			expectedPostFulfillmentEventsCount: 1,
			expectedPostFulfillmentEventsAttributes: []sdk.Attribute{
				sdk.NewAttribute(types.AttributeKeyId, types.BuildDemandIDFromPacketKey(string(rollappPacketKey))),
				sdk.NewAttribute(types.AttributeKeyPrice, "200"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyFee, "50"+sdk.DefaultBondDenom),
				sdk.NewAttribute(types.AttributeKeyIsFulfilled, "true"),
				sdk.NewAttribute(types.AttributeKeyPacketStatus, commontypes.Status_PENDING.String()),
			},
		},
		{
			name:                               "Failed fulfillment - ",
			fulfillmentShouldFail:              true,
			expectedPostFulfillmentEventsCount: 0,
		},
	}

	for _, tc := range tests {
		suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
		expectedFee := "50"
		if tc.fulfillmentShouldFail {
			expectedFee = "30"
		}
		msg := types.NewMsgFulfillOrder(eibcDemandAddr.String(), demandOrder.Id, expectedFee)
		_, err = suite.msgServer.FulfillOrder(suite.Ctx, msg)
		if tc.fulfillmentShouldFail {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}
		suite.AssertEventEmitted(suite.Ctx, tc.expectedPostFulfillmentEventsType, tc.expectedPostFulfillmentEventsCount)
		if tc.expectedPostFulfillmentEventsCount > 0 {
			lastEvent, ok := suite.FindLastEventOfType(suite.Ctx.EventManager().Events(), tc.expectedPostFulfillmentEventsType)
			suite.Require().True(ok)
			suite.AssertAttributes(lastEvent, tc.expectedPostFulfillmentEventsAttributes)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgFulfillOrderWithoutEvents() {
	tests := []struct {
		name                                 string
		demandOrderPacketKey                 string
		demandOrderPrice                     uint64
		demandOrderFee                       uint64
		demandOrderFulfillmentStatus         bool
		demandOrderUnderlyingPacketStatus    commontypes.Status
		demandOrderDenom                     string
		fulfillmentExpectedFee               string
		expectedFulfillmentError             error
		eIBCdemandAddrBalance                math.Int
		expectedDemandOrdefFulfillmentStatus bool
	}{
		{
			name:                                 "Test demand order fulfillment - success",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			eIBCdemandAddrBalance:                math.NewInt(1000),
			expectedDemandOrdefFulfillmentStatus: true,
		},
		{
			name:                                 "order with zero fee - success",
			demandOrderPrice:                     150,
			demandOrderFee:                       0,
			eIBCdemandAddrBalance:                math.NewInt(1000),
			expectedDemandOrdefFulfillmentStatus: true,
		},
		{
			name:                                 "Test demand order fulfillment - wrong expected fee",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			fulfillmentExpectedFee:               "30",
			expectedFulfillmentError:             types.ErrExpectedFeeNotMet,
			eIBCdemandAddrBalance:                math.NewInt(1000),
			expectedDemandOrdefFulfillmentStatus: false,
		},
		{
			name:                                 "Test demand order fulfillment - insufficient balance same denom",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			expectedFulfillmentError:             sdkerrors.ErrInsufficientFunds,
			eIBCdemandAddrBalance:                math.NewInt(130),
			expectedDemandOrdefFulfillmentStatus: false,
		},
		{
			name:                                 "Test demand order fulfillment - insufficient balance different denom",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			demandOrderDenom:                     "adym",
			expectedFulfillmentError:             sdkerrors.ErrInsufficientFunds,
			eIBCdemandAddrBalance:                math.NewInt(130),
			expectedDemandOrdefFulfillmentStatus: false,
		},
		{
			name:                                 "Test demand order fulfillment - already fulfilled",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			demandOrderFulfillmentStatus:         true,
			expectedFulfillmentError:             types.ErrDemandAlreadyFulfilled,
			eIBCdemandAddrBalance:                math.NewInt(300),
			expectedDemandOrdefFulfillmentStatus: true,
		},
		{
			name:                                 "Test demand order fulfillment - status not pending",
			demandOrderPrice:                     150,
			demandOrderFee:                       50,
			demandOrderFulfillmentStatus:         false,
			demandOrderUnderlyingPacketStatus:    commontypes.Status_FINALIZED,
			expectedFulfillmentError:             types.ErrDemandOrderDoesNotExist,
			eIBCdemandAddrBalance:                math.NewInt(300),
			expectedDemandOrdefFulfillmentStatus: false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
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
			if tc.demandOrderDenom == "" {
				tc.demandOrderDenom = sdk.DefaultBondDenom
			}
			demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(tc.demandOrderPrice), math.NewIntFromUint64(tc.demandOrderFee), tc.demandOrderDenom, eibcSupplyAddr.String())
			demandOrder.IsFulfilled = tc.demandOrderFulfillmentStatus
			err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
			suite.Require().NoError(err)
			// Update rollapp status if needed
			if rollappPacket.Status != tc.demandOrderUnderlyingPacketStatus {
				_, err = suite.App.DelayedAckKeeper.UpdateRollappPacketWithStatus(suite.Ctx, *rollappPacket, tc.demandOrderUnderlyingPacketStatus)
				suite.Require().NoError(err, tc.name)
			}

			// try to fulfill the demand order
			demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, tc.demandOrderUnderlyingPacketStatus, demandOrder.Id)
			suite.Require().NoError(err)

			if tc.fulfillmentExpectedFee == "" && len(demandOrder.Fee) > 0 {
				tc.fulfillmentExpectedFee = demandOrder.Fee[0].Amount.String()
			}
			msg := types.NewMsgFulfillOrder(eibcDemandAddr.String(), demandOrder.Id, tc.fulfillmentExpectedFee)
			_, err = suite.msgServer.FulfillOrder(suite.Ctx, msg)
			if tc.expectedFulfillmentError != nil {
				suite.Require().ErrorIs(err, tc.expectedFulfillmentError, tc.name)
			} else {
				suite.Require().NoError(err, tc.name)
			}
			// Check that the demand fulfillment
			demandOrder, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, tc.demandOrderUnderlyingPacketStatus, demandOrder.Id)
			suite.Require().NoError(err)
			suite.Assert().Equal(tc.expectedDemandOrdefFulfillmentStatus, demandOrder.IsFulfilled, tc.name)
			// Check balances updates in case of success
			if tc.expectedFulfillmentError == nil {
				afterFulfillmentSupplyAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcSupplyAddr, sdk.DefaultBondDenom)
				afterFulfillmentDemandAddrBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, eibcDemandAddr, sdk.DefaultBondDenom)
				suite.Require().Equal(eibcSupplyAddrBalance.Add(sdk.NewCoin(sdk.DefaultBondDenom, math.NewIntFromUint64(tc.demandOrderPrice))), afterFulfillmentSupplyAddrBalance)
				suite.Require().Equal(eibcDemandAddrBalance.Sub(sdk.NewCoin(sdk.DefaultBondDenom, math.NewIntFromUint64(tc.demandOrderPrice))), afterFulfillmentDemandAddrBalance)
			}
		})
	}
}
