package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	dacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestMsgFulfillOrder() {
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
			fulfillmentExpectedFee:               "0",
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
			if tc.demandOrderFulfillmentStatus {
				demandOrder.FulfillerAddress = eibcDemandAddr.String() // simulate fulfillment
			}
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
			suite.Assert().Equal(tc.expectedDemandOrdefFulfillmentStatus, demandOrder.IsFulfilled(), tc.name)
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

func (suite *KeeperTestSuite) TestMsgFulfillOrderAuthorized() {
	tests := []struct {
		name                              string
		orderPrice                        sdk.Coin
		orderFee                          sdk.Int
		orderRecipient                    string
		msg                               *types.MsgFulfillOrderAuthorized
		lpAccountBalance                  sdk.Coins
		operatorFeeAccountBalance         sdk.Coins
		expectError                       error
		expectOrderFulfilled              bool
		expectedLPAccountBalance          sdk.Coins
		expectedOperatorFeeAccountBalance sdk.Coins
	}{
		{
			name:           "Successful fulfillment",
			orderPrice:     sdk.NewInt64Coin("adym", 90),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 90)),
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:                  sdk.NewCoins(sdk.NewInt64Coin("adym", 200)),
			operatorFeeAccountBalance:         sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:                       nil,
			expectOrderFulfilled:              true,
			expectedLPAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 108)), // 200 - 90 (price) - 2 (operator fee)
			expectedOperatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 52)),  // 50 + 2 (operator fee)
		},
		{
			name:           "Failure due to mismatched rollapp ID",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           "rollapp2", // Mismatched Rollapp ID
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 100)),
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 200)),
			operatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:               types.ErrRollappIdMismatch,
			expectOrderFulfilled:      false,
			expectedLPAccountBalance:  sdk.NewCoins(sdk.NewInt64Coin("adym", 200)), // Unchanged
		},
		{
			name:           "Failure due to mismatched price",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 110)), // Mismatched Price
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 200)),
			operatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:               types.ErrPriceMismatch,
			expectOrderFulfilled:      false,
			expectedLPAccountBalance:  sdk.NewCoins(sdk.NewInt64Coin("adym", 200)), // Unchanged
		},
		{
			name:           "Failure due to mismatched expected fee",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 100)),
				ExpectedFee:         "15",                                        // Mismatched Expected Fee
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 200)),
			operatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:               types.ErrExpectedFeeNotMet,
			expectOrderFulfilled:      false,
			expectedLPAccountBalance:  sdk.NewCoins(sdk.NewInt64Coin("adym", 200)), // Unchanged
		},
		{
			name:           "Failure due to LP account not existing",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 100)),
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(), // Non-existent LP account
				SettlementValidated: false,
			},
			lpAccountBalance:          nil, // Account does not exist
			operatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:               types.ErrLPAccountDoesNotExist,
			expectOrderFulfilled:      false,
		},
		{
			name:           "Failure due to operator fee account not existing",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 100)),
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),                         // Non-existent operator account
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 200)),
			operatorFeeAccountBalance: nil, // Account does not exist
			expectError:               types.ErrOperatorFeeAccountDoesNotExist,
			expectOrderFulfilled:      false,
			expectedLPAccountBalance:  sdk.NewCoins(sdk.NewInt64Coin("adym", 200)), // Unchanged
		},
		{
			name:           "Failure due to insufficient funds in LP account",
			orderPrice:     sdk.NewInt64Coin("adym", 100),
			orderFee:       sdk.NewInt(10),
			orderRecipient: sample.AccAddress(),
			msg: &types.MsgFulfillOrderAuthorized{
				RollappId:           rollappPacket.RollappId,
				Price:               sdk.NewCoins(sdk.NewInt64Coin("adym", 100)),
				ExpectedFee:         "10",
				OperatorFeeShare:    sdk.DecProto{Dec: sdk.NewDecWithPrec(2, 1)}, // 0.2
				OperatorFeeAddress:  sample.AccAddress(),
				LpAddress:           sample.AccAddress(),
				SettlementValidated: false,
			},
			lpAccountBalance:          sdk.NewCoins(sdk.NewInt64Coin("adym", 90)), // Insufficient funds
			operatorFeeAccountBalance: sdk.NewCoins(sdk.NewInt64Coin("adym", 50)),
			expectError:               sdkerrors.ErrInsufficientFunds,
			expectOrderFulfilled:      false,
			expectedLPAccountBalance:  sdk.NewCoins(sdk.NewInt64Coin("adym", 90)), // Unchanged
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Set up initial state
			suite.SetupTest() // Reset the context and keepers before each test

			// Create accounts
			var lpAccount, operatorFeeAccount sdk.AccAddress

			// LP Account
			if tc.lpAccountBalance != nil {
				lpAccount = sdk.MustAccAddressFromBech32(tc.msg.LpAddress)
				err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, lpAccount, tc.lpAccountBalance)
				require.NoError(suite.T(), err, "Failed to fund LP account")
			}

			// Operator Account
			if tc.operatorFeeAccountBalance != nil {
				operatorFeeAccount = sdk.MustAccAddressFromBech32(tc.msg.OperatorFeeAddress)
				err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, operatorFeeAccount, tc.operatorFeeAccountBalance)
				require.NoError(suite.T(), err, "Failed to fund operator account")
			}

			suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
			demandOrder := types.NewDemandOrder(*rollappPacket, tc.orderPrice.Amount, tc.orderFee, tc.orderPrice.Denom, tc.orderRecipient)
			err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
			suite.Require().NoError(err)

			tc.msg.OrderId = demandOrder.Id

			// Execute the handler
			_, err = suite.msgServer.FulfillOrderAuthorized(sdk.WrapSDKContext(suite.Ctx), tc.msg)

			// Check for expected errors
			if tc.expectError != nil {
				suite.Require().Error(err, tc.name)
				suite.Require().ErrorContains(err, tc.expectError.Error(), tc.name)
			} else {
				suite.Require().NoError(err, tc.name)
			}

			// Check if the demand order is fulfilled
			gotOrder, _ := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_PENDING, demandOrder.Id)
			suite.Require().Equal(tc.expectOrderFulfilled, gotOrder.IsFulfilled(), tc.name)

			// Check account balances if no error expected
			if tc.expectError == nil {
				// LP Account Balance
				lpBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, lpAccount)
				suite.Require().Equal(tc.expectedLPAccountBalance, lpBalance, "LP account balance mismatch")

				// Operator Fee Account Balance (if applicable)
				operatorFeeBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, operatorFeeAccount)
				suite.Require().Equal(tc.expectedOperatorFeeAccountBalance, operatorFeeBalance, "Operator fee account balance mismatch")
			}
		})
	}
}

// TestFulfillOrderEvent tests the event upon fulfilling a demand order
func (suite *KeeperTestSuite) TestFulfillOrderEvent() {
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
		name                               string
		fulfillmentShouldFail              bool
		expectedPostFulfillmentEventsCount int
		expectedPostFulfillmentEvent       proto.Message
	}{
		{
			name:                               "Test demand order fulfillment - success",
			expectedPostFulfillmentEventsCount: 1,
			expectedPostFulfillmentEvent: &types.EventDemandOrderFulfilled{
				OrderId:      types.BuildDemandIDFromPacketKey(string(rollappPacketKey)),
				Price:        "200" + sdk.DefaultBondDenom,
				Fee:          "50" + sdk.DefaultBondDenom,
				IsFulfilled:  true,
				PacketStatus: commontypes.Status_PENDING.String(),
				Fulfiller:    eibcDemandAddr.String(),
				PacketType:   commontypes.RollappPacket_ON_RECV.String(),
			},
		},
		{
			name:                               "Failed fulfillment - no event",
			fulfillmentShouldFail:              true,
			expectedPostFulfillmentEventsCount: 0,
		},
	}

	for _, tc := range tests {
		suite.Ctx = suite.Ctx.WithEventManager(sdk.NewEventManager())
		expectedFee := "50"
		if tc.fulfillmentShouldFail {
			expectedFee = "30" // wrong expected fee to fail the fulfillment msg
		}
		msg := types.NewMsgFulfillOrder(eibcDemandAddr.String(), demandOrder.Id, expectedFee)
		_, err = suite.msgServer.FulfillOrder(suite.Ctx, msg)
		if tc.fulfillmentShouldFail {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}
		eventName := proto.MessageName(tc.expectedPostFulfillmentEvent)
		suite.AssertEventEmitted(suite.Ctx, eventName, tc.expectedPostFulfillmentEventsCount)
		if tc.expectedPostFulfillmentEventsCount > 0 {
			lastEvent, ok := suite.FindLastEventOfType(suite.Ctx.EventManager().Events(), eventName)
			suite.Require().True(ok)
			event, _ := uevent.TypedEventToEvent(tc.expectedPostFulfillmentEvent)
			suite.AssertAttributes(lastEvent, getEventAttributes(event))
		}
	}
}

func getEventAttributes(event sdk.Event) []sdk.Attribute {
	attrs := make([]sdk.Attribute, len(event.Attributes))
	if event.Attributes == nil {
		return attrs
	}
	for i, a := range event.Attributes {
		attrs[i] = sdk.Attribute{
			Key:   a.Key,
			Value: a.Value,
		}
	}
	return attrs
}

func (suite *KeeperTestSuite) TestMsgUpdateDemandOrder() {
	// Create and fund the account
	testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, sdk.NewInt(100_000))
	eibcSupplyAddr := testAddresses[0]

	dackParams := dacktypes.NewParams("hour", sdk.NewDecWithPrec(1, 2), 0) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)
	denom := suite.App.StakingKeeper.BondDenom(suite.Ctx)

	// Set a rollapp packet with 1000 amount
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// Set the initial price and fee for total amount 1000 and 1% bridge fee
	initialFee := sdk.NewInt(100)
	initialPrice := sdk.NewInt(890) // 1000 - 100 fee - 10 bridging fee

	testCases := []struct {
		name          string
		newFee        sdk.Int
		submittedBy   string
		expectError   bool
		expectedPrice sdk.Int
	}{
		{
			name:          "happy case",
			newFee:        sdk.NewInt(400),
			submittedBy:   eibcSupplyAddr.String(),
			expectError:   false,
			expectedPrice: sdk.NewInt(590),
		},
		{
			name:          "happy case - zero eibc fee",
			newFee:        sdk.NewInt(0),
			submittedBy:   eibcSupplyAddr.String(),
			expectError:   false,
			expectedPrice: sdk.NewInt(990),
		},
		{
			name:        "wrong owner",
			newFee:      sdk.NewInt(400),
			submittedBy: testAddresses[1].String(),
			expectError: true,
		},
		{
			name:        "too high fee",
			newFee:      sdk.NewInt(1001),
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
		msg := types.NewMsgUpdateDemandOrder(tc.submittedBy, demandOrder.Id, tc.newFee.String())
		_, err = suite.msgServer.UpdateDemandOrder(suite.Ctx, msg)
		if tc.expectError {
			suite.Require().Error(err, tc.name)
			continue
		}
		suite.Require().NoError(err, tc.name)
		// check if the demand order is updated
		updatedDemandOrder, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, rollappPacket.Status, demandOrder.Id)
		suite.Require().NoError(err, tc.name)
		suite.Assert().Equal(updatedDemandOrder.Fee.AmountOf(denom), tc.newFee, tc.name)
		suite.Assert().Equal(updatedDemandOrder.Price.AmountOf(denom), tc.expectedPrice, tc.name)
	}
}

func (suite *KeeperTestSuite) TestUpdateDemandOrderOnAckOrTimeout() {
	// Create and fund the account
	testAddresses := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, sdk.NewInt(100_000))
	eibcSupplyAddr := testAddresses[0]

	dackParams := dacktypes.NewParams("hour", sdk.NewDecWithPrec(1, 2), 0) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)

	denom := suite.App.StakingKeeper.BondDenom(suite.Ctx)

	onAckRollappPkt := commontypes.RollappPacket{
		RollappId: "testRollappId",
		Status:    commontypes.Status_PENDING,
		Type:      commontypes.RollappPacket_ON_ACK,
		Packet:    &packet,
	}
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, onAckRollappPkt)

	// Set the initial price and fee for total amount 1000
	initialFee := sdk.NewInt(100)
	initialPrice := sdk.NewInt(900)
	demandOrder := types.NewDemandOrder(onAckRollappPkt, initialPrice, initialFee, denom, eibcSupplyAddr.String())
	err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)

	// try to update the demand order
	newFee := sdk.NewInt(400)
	expectedNewPrice := sdk.NewInt(600)
	msg := types.NewMsgUpdateDemandOrder(eibcSupplyAddr.String(), demandOrder.Id, newFee.String())
	_, err = suite.msgServer.UpdateDemandOrder(suite.Ctx, msg)
	suite.Require().NoError(err)
	// check if the demand order is updated
	updatedDemandOrder, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, rollappPacket.Status, demandOrder.Id)
	suite.Require().NoError(err)
	suite.Assert().Equal(updatedDemandOrder.Fee.AmountOf(denom), newFee)
	suite.Assert().Equal(updatedDemandOrder.Price.AmountOf(denom), expectedNewPrice)
}
