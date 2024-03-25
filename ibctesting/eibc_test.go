package ibctesting_test

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
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

	msgServer eibctypes.MsgServer
}

func TestEIBCTestSuite(t *testing.T) {
	suite.Run(t, new(EIBCTestSuite))
}

func (suite *EIBCTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
	eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
	suite.msgServer = eibckeeper.NewMsgServerImpl(eibcKeeper)
	// Change the delayedAck epoch to trigger every month to not
	// delete the rollapp packets and demand orders
	delayedAckKeeper := ConvertToApp(suite.hubChain).DelayedAckKeeper
	params := delayedAckKeeper.GetParams(suite.hubChain.GetContext())
	params.EpochIdentifier = "month"
	delayedAckKeeper.SetParams(suite.hubChain.GetContext(), params)
}

func (suite *EIBCTestSuite) TestEIBCDemandOrderCreation() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Register sequencer
	suite.RegisterSequencer()
	// adding state for the rollapp
	suite.UpdateRollappState(uint64(suite.rollappChain.GetContext().BlockHeight()))
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
		expectAck           bool
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
			_ = suite.TransferRollappToHub(path, IBCSenderAccount, tc.recipient, tc.amount, memo, tc.expectAck)
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

// TestEIBCDemandOrderCreation tests the creation of a demand order and its fulfillment logic.
// It starts by transferring the fulfiller the relevant IBC tokens which it will use to possibly fulfill the demand order.
func (suite *EIBCTestSuite) TestEIBCDemandOrderFulfillment() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Create the path once here so we'll be using the same channel all the time and hence same IBC denom
	// Register sequencer
	suite.RegisterSequencer()
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
			"fulfill demand order fail - insufficient balance",
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
			fulfiller := suite.hubChain.SenderAccounts[fulfillerAccountInitialIndex+idx].SenderAccount.GetAddress()
			// Update the rollapp state
			suite.rollappChain.NextBlock()
			currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
			rollappStateIndex = rollappStateIndex + 1
			suite.UpdateRollappState(currentRollappBlockHeight)
			// Transfer initial IBC funds to fulfiller account with ibc memo
			eibc := map[string]map[string]string{
				"eibc": {
					"fee": tc.EIBCTransferFee,
				},
			}
			eibcJson, _ := json.Marshal(eibc)
			memo := string(eibcJson)
			packet := suite.TransferRollappToHub(path, IBCSenderAccount, fulfiller.String(), tc.fulfillerInitialIBCDenomBalance, memo, false)
			// Finalize rollapp state - at this state no demand order was fulfilled
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			err := suite.FinalizeRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			suite.Require().NoError(err)
			// Check the fulfiller balance was updated fully with the IBC amount
			isUpdated := false
			fulfillerAccountBalanceAfterFinalization := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fulfiller)
			IBCDenom := suite.GetRollappToHubIBCDenomFromPacket(packet)
			requiredFulfillerBalance, ok := sdk.NewIntFromString(tc.fulfillerInitialIBCDenomBalance)
			suite.Require().True(ok)
			for _, coin := range fulfillerAccountBalanceAfterFinalization {
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
			// Get last demand order created by TrackingPacketKey. Last part of the key is the sequence
			lastDemandOrder := getLastDemandOrderByChannelAndSequence(demandOrders)
			// Validate demand order wasn't fulfilled but finalized
			suite.Require().False(lastDemandOrder.IsFullfilled)
			suite.Require().Equal(commontypes.Status_FINALIZED, lastDemandOrder.TrackingPacketStatus)
			// Send another EIBC packet but this time fulfill it with the fulfiller balance.
			// Increase the block height to make sure the next ibc packet won't be considered already finalized when sent
			suite.rollappChain.NextBlock()
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			rollappStateIndex = rollappStateIndex + 1
			suite.UpdateRollappState(currentRollappBlockHeight)
			_ = suite.TransferRollappToHub(path, IBCSenderAccount, IBCOriginalRecipient.String(), tc.IBCTransferAmount, memo, false)
			// Validate demand order created. Calling TransferRollappToHub also promotes the block time for
			// ibc purposes which causes the AfterEpochEnd of the rollapp packet deletion to fire (which also deletes the demand order)
			// hence we should only expect 1 demand order created
			demandOrders, err = eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Greater(len(demandOrders), totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			// Get the last demand order created
			lastDemandOrder = getLastDemandOrderByChannelAndSequence(demandOrders)
			// Try and fulfill the demand order
			preFulfillmentAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fulfiller)
			msgFulfillDemandOrder := &eibctypes.MsgFulfillOrder{
				FulfillerAddress: fulfiller.String(),
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

			// Validate balances of fulfiller and recipient
			fulfillerAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fulfiller)
			recipientAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), IBCOriginalRecipient)
			ibcTransferAmountInt, _ := strconv.ParseInt(tc.IBCTransferAmount, 10, 64)
			eibcTransferFeeInt, _ := strconv.ParseInt(tc.EIBCTransferFee, 10, 64)
			demandOrderPriceInt := ibcTransferAmountInt - eibcTransferFeeInt
			suite.Require().True(fulfillerAccountBalance.IsEqual(preFulfillmentAccountBalance.Sub(sdk.NewCoin(IBCDenom, sdk.NewInt(demandOrderPriceInt)))))
			suite.Require().True(recipientAccountBalance.IsEqual(initialIBCOriginalRecipientBalance.Add(sdk.NewCoin(IBCDenom, sdk.NewInt(demandOrderPriceInt)))))

			// Finalize rollapp and check fulfiller balance was updated with fee
			currentRollappBlockHeight = uint64(suite.rollappChain.GetContext().BlockHeight())
			err = suite.FinalizeRollappState(rollappStateIndex, uint64(currentRollappBlockHeight))
			suite.Require().NoError(err)
			fulfillerAccountBalanceAfterFinalization = eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fulfiller)
			suite.Require().True(fulfillerAccountBalanceAfterFinalization.IsEqual(preFulfillmentAccountBalance.Add(sdk.NewCoin(IBCDenom, sdk.NewInt(eibcTransferFeeInt)))))

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

// TestHubToRollappEarlyFulfillment : when a packet hub->rollapp times out, or gets an error ack, than eIBC can be used to recover quickly.
func (suite *EIBCTestSuite) TestTimeoutEIBCDemandOrderFulfillment() {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)
	// Setup endpoints
	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB
	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()
	// Create rollapp and update its initial state
	suite.CreateRollapp()
	suite.RegisterSequencer()
	suite.UpdateRollappState(uint64(suite.rollappChain.GetContext().BlockHeight()))

	type TC struct {
		name     string
		malleate func(channeltypes.Packet)
	}

	for _, tc := range []TC{
		{
			name: "",
			malleate: func(packet channeltypes.Packet) {
				// TestTimeoutEIBCDemandOrderFulfillment tests the following:
				// 1. Send a packet from hub to rollapp and timeout the packet.
				// 2. Validate a new demand order is created.
				// 3. Fulfill the demand order and validate the original sender and fulfiller balances are updated.
				// 4. Finalize the rollapp state and validate the demand order fulfiller balance is updated with the amount.

				// Timeout the packet. Shouldn't release funds until rollapp height is finalized
				err := hubEndpoint.TimeoutPacket(packet)
				suite.Require().NoError(err)
			},
		},
		{
			name: "",

			malleate: func(packet channeltypes.Packet) {
				// TestAckErrEIBCDemandOrderFulfillment tests the following:
				// 1. Send a packet from hub to rollapp and cause an errored ack from the packet.
				// 2. Validate a new demand order is created.
				// 3. Fulfill the demand order and validate the original sender and fulfiller balances are updated.
				// 4. Finalize the rollapp state and validate the demand order fulfiller balance is updated with the amount.

				// return an err ack
				ack := channeltypes.NewErrorAcknowledgement(errors.New("foobar"))
				err := rollappEndpoint.WriteAcknowledgement(ack, packet)
				suite.Require().NoError(err)
				err = hubEndpoint.AcknowledgePacket(packet, ack.Acknowledgement())
				suite.Require().NoError(err)
			},
		},
	} {
		suite.Run(tc.name, func() {
			// Set the timeout height
			timeoutHeight := clienttypes.GetSelfHeight(suite.rollappChain.GetContext())
			amount, ok := sdk.NewIntFromString("1000000000000000000") // 1DYM
			suite.Require().True(ok)
			coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)
			// Setup accounts
			senderAccount := hubEndpoint.Chain.SenderAccount.GetAddress()
			receiverAccount := rollappEndpoint.Chain.SenderAccount.GetAddress()
			fulfillerAccount := suite.hubChain.SenderAccounts[1].SenderAccount.GetAddress()
			// Get initial balances
			bankKeeper := ConvertToApp(suite.hubChain).BankKeeper
			senderInitialBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
			fulfillerInitialBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), fulfillerAccount, sdk.DefaultBondDenom)
			receiverInitialBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), receiverAccount, sdk.DefaultBondDenom)
			// Send from hubChain to rollappChain
			msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
			res, err := hubEndpoint.Chain.SendMsgs(msg)
			suite.Require().NoError(err)
			packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
			suite.Require().NoError(err)
			found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(hubEndpoint.Chain.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			suite.Require().True(found)
			// Check balance decreased
			postSendBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
			suite.Require().Equal(senderInitialBalance.Amount.Sub(coinToSendToB.Amount), postSendBalance.Amount)
			// Validate no demand orders exist
			eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
			demandOrders, err := eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Equal(len(demandOrders), 0)
			// Update the client to create timeout
			err = hubEndpoint.UpdateClient()
			suite.Require().NoError(err)

			tc.malleate(packet)

			// Validate funds are still not returned to the sender
			postTimeoutBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
			suite.Require().Equal(postSendBalance.Amount, postTimeoutBalance.Amount)
			// Validate demand order created
			demandOrders, err = eibcKeeper.ListAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().NoError(err)
			suite.Require().Greater(len(demandOrders), 0)
			// Get the last demand order created t
			lastDemandOrder := getLastDemandOrderByChannelAndSequence(demandOrders)
			// Validate the demand order price and denom
			timeoutFee := eibcKeeper.GetParams(suite.hubChain.GetContext()).TimeoutFee
			amountDec, err := sdk.NewDecFromStr(coinToSendToB.Amount.String())
			suite.Require().NoError(err)
			expectedPrice := amountDec.Mul(sdk.NewDec(1).Sub(timeoutFee)).TruncateInt()
			suite.Require().Equal(expectedPrice, lastDemandOrder.Price[0].Amount)
			suite.Require().Equal(coinToSendToB.Denom, lastDemandOrder.Price[0].Denom)
			// Fulfill the demand order
			msgFulfillDemandOrder := &eibctypes.MsgFulfillOrder{
				FulfillerAddress: fulfillerAccount.String(),
				OrderId:          lastDemandOrder.Id,
			}
			_, err = suite.msgServer.FulfillOrder(suite.hubChain.GetContext(), msgFulfillDemandOrder)
			suite.Require().NoError(err)
			// Validate balances of fulfiller and sender are updated while the original recipient is not
			fulfillerAccountBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), fulfillerAccount, sdk.DefaultBondDenom)
			senderAccountBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), senderAccount, sdk.DefaultBondDenom)
			receiverAccountBalance := bankKeeper.GetBalance(suite.hubChain.GetContext(), receiverAccount, sdk.DefaultBondDenom)
			suite.Require().True(fulfillerAccountBalance.IsEqual(fulfillerInitialBalance.Sub(lastDemandOrder.Price[0])))
			suite.Require().True(senderAccountBalance.IsEqual(senderInitialBalance.Sub(lastDemandOrder.Fee[0])))
			suite.Require().True(receiverAccountBalance.IsEqual(receiverInitialBalance))
			// Finalize the rollapp state
			currentRollappBlockHeight := uint64(suite.rollappChain.GetContext().BlockHeight())
			err = suite.FinalizeRollappState(1, currentRollappBlockHeight)
			suite.Require().NoError(err)
			// Funds are passed to the fulfiller
			fulfillerAccountBalanceAfterTimeout := bankKeeper.GetBalance(suite.hubChain.GetContext(), fulfillerAccount, sdk.DefaultBondDenom)
			suite.Require().True(fulfillerAccountBalanceAfterTimeout.IsEqual(fulfillerInitialBalance.Add(lastDemandOrder.Fee[0])))
		})
	}
}

/* -------------------------------------------------------------------------- */
/*                                    Utils                                   */
/* -------------------------------------------------------------------------- */

// TransferRollappToHub sends a transfer packet from rollapp to hub and returns the packet
func (suite *EIBCTestSuite) TransferRollappToHub(
	path *ibctesting.Path,
	sender string,
	receiver string,
	amount string,
	memo string,
	expectAck bool,
) channeltypes.Packet {
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

	// If there is an error than an ack is returned immediately. Relay will always try to extract an ack, and will
	// return an error if there isn't one.
	if expectAck {
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

// Each demand order tracks the underlying packet key which can than indicate the order by the channel and seuqence
func getLastDemandOrderByChannelAndSequence(demandOrders []*eibctypes.DemandOrder) *eibctypes.DemandOrder {
	sort.Slice(demandOrders, func(i, j int) bool {
		iKeyParts := strings.Split((demandOrders)[i].TrackingPacketKey, "/")
		jKeyParts := strings.Split((demandOrders)[j].TrackingPacketKey, "/")
		return iKeyParts[len(iKeyParts)-1] < jKeyParts[len(jKeyParts)-1]
	})
	return demandOrders[len(demandOrders)-1]
}
