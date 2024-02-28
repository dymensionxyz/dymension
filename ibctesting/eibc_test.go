package ibctesting_test

import (
	"encoding/json"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type EIBCTestSuite struct {
	IBCTestUtilSuite

	msgServer   eibctypes.MsgServer
	ctx         sdk.Context
	queryClient eibctypes.QueryClient
}

func TestEIBCTestSuite(t *testing.T) {
	suite.Run(t, new(EIBCTestSuite))
}

func (suite *EIBCTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
	eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
	suite.msgServer = eibckeeper.NewMsgServerImpl(eibcKeeper)
}

func (suite *EIBCTestSuite) TestEIBCDemandOrderCreation() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Create path so we'll be using the same channel
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)
	// Setup globals for the test cases
	IBCSenderAccount := suite.rollappChain.SenderAccount.GetAddress().String()
	// Create cases
	cases := []struct {
		name                string
		amount              string
		fee                 string
		recipient           string
		demandOrdersCreated int
		isAckError          bool
		extraMemoData       map[string]map[string]string
	}{
		{
			"valid demand order",
			"1000000000",
			"150",
			suite.hubChain.SenderAccount.GetAddress().String(),
			1,
			false,
			map[string]map[string]string{},
		},
		{
			"invalid demand order - negative fee",
			"1000000000",
			"-150",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
			map[string]map[string]string{},
		},
		{
			"invalid demand order - fee > amount",
			"1000",
			"1001",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
			map[string]map[string]string{},
		},
		{
			"invalid demand order - fee is 0",
			"1",
			"0",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
			map[string]map[string]string{},
		},
		{
			"invalid demand order - fee > max uint64",
			"10000",
			"100000000000000000000000000000",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
			map[string]map[string]string{},
		},
		{
			"invalid demand order - PFM and EIBC are not supported together",
			"1000000000",
			"150",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
			map[string]map[string]string{"forward": {
				"receiver": suite.hubChain.SenderAccount.GetAddress().String(),
				"port":     "transfer",
				"channel":  "channel-0",
			}},
		},
	}
	totalDemandOrdersCreated := 0
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			// Send the EIBC Packet
			memoObj := map[string]map[string]string{
				"eibc": {
					"fee": tc.fee,
				},
			}
			if tc.extraMemoData != nil {
				for key, value := range tc.extraMemoData {
					memoObj[key] = value
				}
			}
			eibcJson, _ := json.Marshal(memoObj)
			memo := string(eibcJson)
			_ = suite.TransferRollappToHub(path, IBCSenderAccount, tc.recipient, tc.amount, memo, tc.isAckError)
			// Validate demand orders results
			eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
			demandOrders, err := eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Equal(tc.demandOrdersCreated, len(demandOrders)-totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			amountInt, ok := sdk.NewIntFromString(tc.amount)
			suite.Require().True(ok)
			feeInt, ok := sdk.NewIntFromString(tc.fee)
			suite.Require().True(ok)
			if tc.demandOrdersCreated > 0 {
				lastDemandOrder := demandOrders[len(demandOrders)-1]
				suite.Require().True(ok)
				suite.Require().Equal(tc.recipient, lastDemandOrder.Recipient)
				suite.Require().Equal(amountInt.Sub(feeInt), lastDemandOrder.Price[0].Amount)
				suite.Require().Equal(feeInt, lastDemandOrder.Fee[0].Amount)
			}

		})
	}
}

// TestEIBCDemandOrderCreation tests the creation of a demand order and its fullfillment logic.
// It starts by trannsferring the supplier the relevant IBC tokens which it will use to possibly fulfill the demand order.
func (suite *EIBCTestSuite) TestEIBCDemandOrderFulfillment() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Create the path once here so we'll be using the same channel all the time and hence same IBC denom
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)
	// Setup globals for the test
	totalDemandOrdersCreated := 0
	eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
	delayedAckKeeper := ConvertToApp(suite.hubChain).DelayedAckKeeper
	IBCSenderAccount := suite.rollappChain.SenderAccount.GetAddress().String()
	rollappStateIndex := uint64(0)
	IBCrecipientAccountInitialIndex := 0
	fulfillerAccountInitialIndex := 1
	// Create cases
	cases := []struct {
		name                            string
		IBCTransferAmount               string
		EIBCTransferFee                 string
		fulfillerInitialIBCDenomBalance string
		isFulfilledSuccess              bool
	}{
		{
			"fulfill demand order successfully",
			"200",
			"150",
			"300",
			true,
		},
		{
			"fulfill demand order fail - insufficent balance",
			"200",
			"40",
			"49",
			false,
		},
	}
	for idx, tc := range cases {
		suite.Run(tc.name, func() {
			// Get the initial state of the accounts
			IBCOriginalRecipient := suite.hubChain.SenderAccounts[IBCrecipientAccountInitialIndex+idx].SenderAccount.GetAddress()
			initialIBCOriginalRecipientBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), IBCOriginalRecipient)
			fullfillerAccount := suite.hubChain.SenderAccounts[fulfillerAccountInitialIndex+idx].SenderAccount.GetAddress()
			// Update the rollapp state
			suite.rollappChain.NextBlock()
			currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
			rollappStateIndex = rollappStateIndex + 1
			suite.UpdateRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			// Transfer initial IBC funder to fulfiller account
			eibc := map[string]map[string]string{
				"eibc": {
					"fee": tc.EIBCTransferFee,
				},
			}
			eibcJson, _ := json.Marshal(eibc)
			memo := string(eibcJson)
			packet := suite.TransferRollappToHub(path, IBCSenderAccount, fullfillerAccount.String(), tc.fulfillerInitialIBCDenomBalance, memo, false)
			// Finalize rollapp state
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			suite.FinalizeRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			// Check the fulfiller balance was updated fully with the IBC amount
			isUpdated := false
			fullfillerAccountBalanceAfterFinalization := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount)
			IBCDenom := suite.GetRollappToHubIBCDenomFromPacket(packet)
			requiredFulfillerBalance, ok := sdk.NewIntFromString(tc.fulfillerInitialIBCDenomBalance)
			suite.Require().True(ok)
			for _, coin := range fullfillerAccountBalanceAfterFinalization {
				if coin.Denom == IBCDenom && coin.Amount.Equal(requiredFulfillerBalance) {
					isUpdated = true
					break
				}
			}
			suite.Require().True(isUpdated)
			// Validate eibc demand order created
			demandOrders, err := eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Greater(len(demandOrders), totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			lastDemandOrder := demandOrders[len(demandOrders)-1]
			// Validate demand order wasn't fulfilled but finalized
			suite.Require().False(lastDemandOrder.IsFullfilled)
			suite.Require().Equal(commontypes.Status_FINALIZED, lastDemandOrder.TrackingPacketStatus)

			// Send another EIBC packet but this time fulfill it with the fulfiller balance.
			// Increase the block height to make sure the next ibc packet won't be considered already finalized when sent
			suite.rollappChain.NextBlock()
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			rollappStateIndex = rollappStateIndex + 1
			suite.UpdateRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			packet = suite.TransferRollappToHub(path, IBCSenderAccount, IBCOriginalRecipient.String(), tc.IBCTransferAmount, memo, false)

			// Validate demand order created
			demandOrders, err = eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Greater(len(demandOrders), totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			// Get the last demand order created
			for _, order := range demandOrders {
				if order.Id != lastDemandOrder.Id {
					lastDemandOrder = order
					break
				}
			}
			// Try and fulfill the demand order
			preFulfillmentAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount)
			msgFulfillDemandOrder := &eibctypes.MsgFulfillOrder{
				FulfillerAddress: fullfillerAccount.String(),
				OrderId:          lastDemandOrder.Id,
			}
			// Validate demand order status based on fulfillment success
			_, err = suite.msgServer.FulfillOrder(suite.hubChain.GetContext(), msgFulfillDemandOrder)
			if !tc.isFulfilledSuccess {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			// Validate eibc packet recipient has been updated
			rollappPacket, err := delayedAckKeeper.GetRollappPacket(suite.hubChain.GetContext(), lastDemandOrder.TrackingPacketKey)
			suite.Require().NoError(err)
			var data transfertypes.FungibleTokenPacketData
			err = transfertypes.ModuleCdc.UnmarshalJSON(rollappPacket.Packet.GetData(), &data)
			suite.Require().NoError(err)
			suite.Require().Equal(msgFulfillDemandOrder.FulfillerAddress, data.Receiver)

			// Validate balances of fullfiller and recipient
			fullfillerAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount)
			recipientAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), IBCOriginalRecipient)
			ibcTransferAmountInt, _ := strconv.ParseInt(tc.IBCTransferAmount, 10, 64)
			eibcTransferFeeInt, _ := strconv.ParseInt(tc.EIBCTransferFee, 10, 64)
			demandOrderPriceInt := ibcTransferAmountInt - eibcTransferFeeInt
			suite.Require().True(fullfillerAccountBalance.IsEqual(preFulfillmentAccountBalance.Sub(sdk.NewCoin(IBCDenom, sdk.NewInt(demandOrderPriceInt)))))
			suite.Require().True(recipientAccountBalance.IsEqual(initialIBCOriginalRecipientBalance.Add(sdk.NewCoin(IBCDenom, sdk.NewInt(demandOrderPriceInt)))))

			// Finalize rollapp and check fulfiller balance was updated with fee
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			suite.FinalizeRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			fullfillerAccountBalanceAfterFinalization = eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount)
			suite.Require().True(fullfillerAccountBalanceAfterFinalization.IsEqual(preFulfillmentAccountBalance.Add(sdk.NewCoin(IBCDenom, sdk.NewInt(eibcTransferFeeInt)))))

			// Validate demand order fulfilled and packet status updated
			finalizedDemandOrders, err := eibcKeeper.ListDemandOrdersByStatus(suite.hubChain.GetContext(), commontypes.Status_FINALIZED)
			suite.Require().NoError(err)
			var finalizedDemandOrder *eibctypes.DemandOrder
			for _, order := range finalizedDemandOrders {
				if order.Id == lastDemandOrder.Id {
					finalizedDemandOrder = order
					break
				}
			}
			suite.Require().NotNil(finalizedDemandOrder)
			suite.Require().True(finalizedDemandOrder.IsFullfilled)
			suite.Require().Equal(commontypes.Status_FINALIZED, finalizedDemandOrder.TrackingPacketStatus)

		})
	}
}

/* -------------------------------------------------------------------------- */
/*                                    Utils                                   */
/* -------------------------------------------------------------------------- */

func (suite *EIBCTestSuite) TransferRollappToHub(path *ibctesting.Path, sender string, receiver string, amount string, memo string, isAckError bool) channeltypes.Packet {
	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB

	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amountInt, ok := sdk.NewIntFromString(amount)
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amountInt)

	msg := types.NewMsgTransfer(rollappEndpoint.ChannelConfig.PortID, rollappEndpoint.ChannelID,
		coinToSendToB, sender, receiver, timeoutHeight, 0, memo)
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = path.RelayPacket(packet)

	// If ack error that an ack is retuned immediately hence found. The reason we get err in the relay packet is
	// beacuse no ack can be parsed from events
	if isAckError {
		suite.Require().NoError(err)
		found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		suite.Require().True(found)
	} else {
		suite.Require().Error(err)
		found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		suite.Require().False(found)
	}
	return packet

}
