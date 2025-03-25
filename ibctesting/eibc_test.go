package ibctesting_test

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type eibcSuite struct {
	utilSuite
	path *ibctesting.Path
}

func (s *eibcSuite) msgServer() eibctypes.MsgServer {
	return eibckeeper.NewMsgServerImpl(s.hubApp().EIBCKeeper)
}

func TestEIBCTestSuite(t *testing.T) {
	suite.Run(t, new(eibcSuite))
}

func (s *eibcSuite) SetupTest() {
	s.utilSuite.SetupTest()
	s.hubApp().LightClientKeeper.SetEnabled(false)

	s.hubApp().BankKeeper.SetDenomMetaData(s.hubCtx(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})
	// Change the delayedAck epoch to trigger every month to not
	// delete the rollapp packets and demand orders
	delayedAckKeeper := s.hubApp().DelayedAckKeeper
	params := delayedAckKeeper.GetParams(s.hubCtx())
	params.EpochIdentifier = "month"
	params.BridgingFee = math.LegacyZeroDec()
	delayedAckKeeper.SetParams(s.hubCtx(), params)
	// Create path so we'll be using the same channel
	path := s.newTransferPath(s.hubChain(), s.rollappChain())
	s.coordinator.Setup(path)
	// Create rollapp only once
	s.createRollappWithFinishedGenesis(path.EndpointA.ChannelID)
	// Register sequencer
	s.registerSequencer()
	s.path = path
	s.setRollappLightClientID(s.rollappChain().ChainID, path.EndpointA.ClientID)
}

func (s *eibcSuite) TestEIBCDemandOrderCreation() {
	// adding state for the rollapp
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))
	// Setup globals for the test cases
	IBCSenderAccount := s.rollappChain().SenderAccount.GetAddress().String()
	// Create cases
	cases := []struct {
		name                string
		amount              string
		fee                 string
		demandOrdersCreated int
		expectAck           bool
		extraMemoData       map[string]map[string]string
		skipEIBCmemo        bool
	}{
		{
			"valid demand order",
			"1000000000",
			"150",
			1,
			false,
			map[string]map[string]string{},
			false,
		},
		{
			"valid demand order - fee is 0",
			"1000000000",
			"0",
			1,
			false,
			map[string]map[string]string{},
			false,
		},
		{
			"valid demand order - auto created",
			"1000000000",
			"0",
			1,
			false,
			map[string]map[string]string{},
			true,
		},
		{
			"invalid demand order - negative fee",
			"1000000000",
			"-150",
			0,
			true,
			map[string]map[string]string{},
			false,
		},
		{
			"invalid demand order - fee > amount",
			"1000",
			"1001",
			0,
			true,
			map[string]map[string]string{},
			false,
		},
		{
			"invalid demand order - fee > max uint64",
			"10000",
			"100000000000000000000000000000",
			0,
			true,
			map[string]map[string]string{},
			false,
		},
		{
			"invalid demand order - PFM and EIBC are not supported together",
			"1000000000",
			"150",
			0,
			true,
			map[string]map[string]string{"forward": {
				"receiver": s.hubChain().SenderAccount.GetAddress().String(),
				"port":     "transfer",
				"channel":  "channel-0",
			}},
			false,
		},
	}
	totalDemandOrdersCreated := 0
	for _, tc := range cases {
		s.Run(tc.name, func() {
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

			if tc.skipEIBCmemo {
				memo = ""
			}

			recipient := apptesting.CreateRandomAccounts(1)[0]
			_ = s.transferRollappToHub(s.path, IBCSenderAccount, recipient.String(), tc.amount, memo, tc.expectAck)
			// Validate demand orders results
			eibcKeeper := s.hubApp().EIBCKeeper
			demandOrders, err := eibcKeeper.ListAllDemandOrders(s.hubCtx())
			s.Require().NoError(err)
			s.Require().Equal(tc.demandOrdersCreated, len(demandOrders)-totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)

			amountInt, ok := math.NewIntFromString(tc.amount)
			s.Require().True(ok)
			feeInt, ok := math.NewIntFromString(tc.fee)
			s.Require().True(ok)
			if tc.demandOrdersCreated > 0 {
				var demandOrder *eibctypes.DemandOrder
				for _, order := range demandOrders {
					if order.Recipient == recipient.String() {
						demandOrder = order
						break
					}
				}
				s.Require().NotNil(demandOrder)
				s.Require().Equal(recipient.String(), demandOrder.Recipient)
				s.Require().Equal(amountInt.Sub(feeInt), demandOrder.Price[0].Amount)
				s.Require().Equal(feeInt, demandOrder.Fee.AmountOf(demandOrder.Price[0].Denom))
			}
		})
	}
}

// TestEIBCDemandOrderFulfillment tests the creation of a demand order and its fulfillment logic.
// It starts by transferring the fulfiller the relevant IBC tokens which it will use to possibly fulfill the demand order.
func (s *eibcSuite) TestEIBCDemandOrderFulfillment() {
	s.eibcTransferFulfillment([]eibcTransferFulfillmentTC{
		{
			"fulfill demand order successfully",
			"200",
			"150",
			"300",
			false,
			nil,
		},
		{
			"fulfill demand order fail - insufficient balance",
			"200",
			"40",
			"49",
			true,
			nil,
		},
	})
}

type eibcTransferFulfillmentTC struct {
	name              string
	transferAmt       string
	eibcTransferFee   string
	fulfillerStartBal string
	expectFulfillFail bool
	fulfillHook       []byte
}

func createMemo(eibcFee string, fulfillHook []byte) string {

	m := map[string]any{
		"eibc": map[string]string{
			"fee": eibcFee,
		},
	}
	if len(fulfillHook) > 0 {
		m["fulfill_hook"] = fulfillHook
	}

	eibcJson, _ := json.Marshal(m)

	return string(eibcJson)
}

func (s *eibcSuite) eibcTransferFulfillment(cases []eibcTransferFulfillmentTC) {
	numOrdersTotal := 0
	eibcK := s.hubApp().EIBCKeeper
	delayedackK := s.hubApp().DelayedAckKeeper
	transferSrcAcc := s.rollappChain().SenderAccount.GetAddress().String()
	rolStateIx := uint64(0) // FIXME: why is this shared...

	for idx, tc := range cases {
		s.Run(tc.name, func() {
			ibcRecipientIx := idx
			fulfillerIx := idx + 1 // offset
			rolStateIx++

			ibcRecipient := s.hubChain().SenderAccounts[ibcRecipientIx].SenderAccount.GetAddress()
			ibcRecipientBalBefore := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), ibcRecipient)
			fulfiller := s.hubChain().SenderAccounts[fulfillerIx].SenderAccount.GetAddress()

			s.rollappChain().NextBlock()
			rolH := uint64(s.rollappCtx().BlockHeight())
			s.updateRollappState(rolH)

			memo := createMemo(tc.eibcTransferFee, tc.fulfillHook)
			var IBCDenom string

			// transfer to fulfiller so he has money
			{

				packet := s.transferRollappToHub(s.path, transferSrcAcc, fulfiller.String(), tc.fulfillerStartBal, memo, false)
				// just finalize immediately, no eibc
				rolH = uint64(s.rollappCtx().BlockHeight())
				_, err := s.finalizeRollappState(rolStateIx, rolH)
				s.Require().NoError(err)
				s.finalizePacketsByAddr(fulfiller.String())

				// check it worked (balance + demand order)
				pass := false
				bal := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), fulfiller)
				IBCDenom = s.getRollappToHubIBCDenomFromPacket(packet)
				expect, ok := math.NewIntFromString(tc.fulfillerStartBal)
				s.Require().True(ok)
				for _, coin := range bal {
					if coin.Denom == IBCDenom && coin.Amount.Equal(expect) {
						pass = true
						break
					}
				}

				// TODO: why is demand order checked, why pass eibc in memo at all?
				s.Require().True(pass)
				orders, err := eibcK.ListAllDemandOrders(s.hubCtx())
				s.Require().NoError(err)
				s.Require().Greater(len(orders), numOrdersTotal)
				numOrdersTotal = len(orders)
				// Get last demand order created by TrackingPacketKey. Last part of the key is the sequence
				lastOrder := getLastDemandOrderByChannelAndSequence(orders)
				// Check wasn't fulfilled but finalized
				s.Require().False(lastOrder.IsFulfilled())
				s.Require().Equal(commontypes.Status_FINALIZED, lastOrder.TrackingPacketStatus)
			}

			// Send another EIBC packet but this time fulfill it with the fulfiller balance.
			s.rollappChain().NextBlock()
			// increase the block height to make sure the next ibc packet won't be considered already finalized when sent
			rolH = uint64(s.rollappCtx().BlockHeight())
			rolStateIx = rolStateIx + 1
			s.updateRollappState(rolH)
			packet := s.transferRollappToHub(s.path, transferSrcAcc, ibcRecipient.String(), tc.transferAmt, memo, false)
			s.Require().True(s.rollappHasPacketCommitment(packet))

			// Validate demand order created. Calling TransferRollappToHub also promotes the block time for
			// ibc purposes which causes the AfterEpochEnd of the rollapp packet deletion to fire (which also deletes the demand order)
			// hence we should only expect 1 demand order created
			orders, err := eibcK.ListAllDemandOrders(s.hubCtx())

			s.Require().NoError(err)
			s.Require().Greater(len(orders), numOrdersTotal)
			numOrdersTotal = len(orders)
			lastOrder := getLastDemandOrderByChannelAndSequence(orders)

			// Try and fulfill the demand order
			fulfillerBalBefore := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), fulfiller)
			mFulfill := &eibctypes.MsgFulfillOrder{
				FulfillerAddress: fulfiller.String(),
				OrderId:          lastOrder.Id,
				ExpectedFee:      tc.eibcTransferFee,
			}
			_, err = s.msgServer().FulfillOrder(s.hubCtx(), mFulfill)
			if tc.expectFulfillFail {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Check packet recipient is updated
			rolPacket, err := delayedackK.GetRollappPacket(s.hubCtx(), lastOrder.TrackingPacketKey)
			s.Require().NoError(err)
			var transferData transfertypes.FungibleTokenPacketData
			err = transfertypes.ModuleCdc.UnmarshalJSON(rolPacket.Packet.GetData(), &transferData)
			s.Require().NoError(err)
			s.Require().Equal(mFulfill.FulfillerAddress, transferData.Receiver)

			// Check balances
			fulfillerBal := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), fulfiller)
			recipientBal := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), ibcRecipient)
			transferAmt, _ := strconv.ParseInt(tc.transferAmt, 10, 64)
			fee, _ := strconv.ParseInt(tc.eibcTransferFee, 10, 64)
			feeCoin := sdk.NewCoin(IBCDenom, math.NewInt(fee))
			price := transferAmt - fee
			priceCoin := sdk.NewCoin(IBCDenom, math.NewInt(price))

			s.Require().True(fulfillerBal.Equal(fulfillerBalBefore.Sub(priceCoin)))
			s.Require().True(recipientBal.Equal(ibcRecipientBalBefore.Add(priceCoin)))

			// Finalize rollapp and check fulfiller balance was updated with fee
			rolH = uint64(s.rollappCtx().BlockHeight())
			_, err = s.finalizeRollappState(rolStateIx, rolH)
			s.Require().NoError(err)
			evts := s.finalizePacketsByAddr(fulfiller.String())

			ack, err := ibctesting.ParseAckFromEvents(evts.ToABCIEvents())
			s.Require().NoError(err)

			fulfillerBalAfter := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), fulfiller)
			s.Require().True(fulfillerBalAfter.Equal(fulfillerBalBefore.Add(feeCoin)))

			// Check order fulfilled and packet status
			finalizedOrders, err := eibcK.ListDemandOrdersByStatus(s.hubCtx(), commontypes.Status_FINALIZED, 0)
			s.Require().NoError(err)
			var finalizedO *eibctypes.DemandOrder
			for _, o := range finalizedOrders {
				if o.Id == lastOrder.Id {
					finalizedO = o
					break
				}
			}
			s.Require().NotNil(finalizedO)
			s.Require().True(finalizedO.IsFulfilled())
			s.Require().Equal(commontypes.Status_FINALIZED, finalizedO.TrackingPacketStatus)

			s.path.EndpointA.Chain.NextBlock()
			_ = s.path.EndpointB.UpdateClient()
			err = s.path.EndpointB.AcknowledgePacket(packet, ack)
			s.Require().NoError(err)
		})
	}
}

func (s *eibcSuite) rollappHasPacketCommitment(packet channeltypes.Packet) bool {
	return s.hubChain().App.GetIBCKeeper().ChannelKeeper.HasPacketCommitment(s.rollappCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
}

// TestTimeoutEIBCDemandOrderFulfillment: when a packet hub->rollapp times out, or gets an error ack, than eIBC can be used to recover quickly.
func (s *eibcSuite) TestTimeoutEIBCDemandOrderFulfillment() {
	// Setup endpoints
	hubEndpoint := s.path.EndpointA
	rollappEndpoint := s.path.EndpointB
	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()
	// Create rollapp and update its initial state
	s.updateRollappState(uint64(s.rollappCtx().BlockHeight()))

	type TC struct {
		name     string
		malleate func(channeltypes.Packet)
		fee      func(params eibctypes.Params) math.LegacyDec
	}

	nOrdersCreated := 0

	for _, tc := range []TC{
		{
			name: "timeout",
			malleate: func(packet channeltypes.Packet) {
				// TestTimeoutEIBCDemandOrderFulfillment tests the following:
				// 1. Send a packet from hub to rollapp and timeout the packet.
				// 2. Validate a new demand order is created.
				// 3. Fulfill the demand order and validate the original sender and fulfiller balances are updated.
				// 4. Finalize the rollapp state and validate the demand order fulfiller balance is updated with the amount.

				// Timeout the packet. Shouldn't release funds until rollapp height is finalized
				err := hubEndpoint.TimeoutPacket(packet)
				s.Require().NoError(err)
			},
			fee: func(params eibctypes.Params) math.LegacyDec {
				return params.TimeoutFee
			},
		},
		{
			name: "err acknowledgement",
			malleate: func(packet channeltypes.Packet) {
				// TestAckErrEIBCDemandOrderFulfillment tests the following:
				// 1. Send a packet from hub to rollapp and cause an errored ack from the packet.
				// 2. Validate a new demand order is created.
				// 3. Fulfill the demand order and validate the original sender and fulfiller balances are updated.
				// 4. Finalize the rollapp state and validate the demand order fulfiller balance is updated with the amount.

				// return an err ack
				ack := uevent.NewErrorAcknowledgement(s.hubCtx(), errors.New("foobar"))
				err := rollappEndpoint.WriteAcknowledgement(ack, packet)
				s.Require().NoError(err)
				err = hubEndpoint.AcknowledgePacket(packet, ack.Acknowledgement())
				s.Require().NoError(err)
			},
			fee: func(params eibctypes.Params) math.LegacyDec {
				return params.ErrackFee
			},
		},
	} {
		s.Run(tc.name, func() {
			// Set the timeout height
			timeoutHeight := clienttypes.GetSelfHeight(s.rollappCtx())
			amount, ok := math.NewIntFromString("1000000000000000000") // 1DYM
			s.Require().True(ok)
			coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)
			// Setup accounts
			senderAccount := s.hubChain().SenderAccount.GetAddress()
			receiverAccount := s.rollappChain().SenderAccount.GetAddress()
			fulfillerAccount := s.hubChain().SenderAccounts[1].SenderAccount.GetAddress()
			// Get initial balances
			bankKeeper := s.hubApp().BankKeeper
			senderInitialBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
			fulfillerInitialBalance := bankKeeper.GetBalance(s.hubCtx(), fulfillerAccount, sdk.DefaultBondDenom)
			receiverInitialBalance := bankKeeper.GetBalance(s.hubCtx(), receiverAccount, sdk.DefaultBondDenom)
			// Send from hubChain to rollappChain
			msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSendToB, senderAccount.String(), receiverAccount.String(), timeoutHeight, disabledTimeoutTimestamp, "")
			res, err := s.hubChain().SendMsgs(msg)
			s.Require().NoError(err)
			packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
			s.Require().NoError(err)
			found := hubIBCKeeper.ChannelKeeper.HasPacketCommitment(s.hubCtx(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
			s.Require().True(found)
			// Check balance decreased
			postSendBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
			s.Require().Equal(senderInitialBalance.Amount.Sub(coinToSendToB.Amount), postSendBalance.Amount)
			// Validate no demand orders exist
			eibcKeeper := s.hubApp().EIBCKeeper
			demandOrders, err := eibcKeeper.ListAllDemandOrders(s.hubCtx())
			s.Require().NoError(err)
			s.Require().Equal(nOrdersCreated, len(demandOrders))
			// Update the client to create timeout
			err = hubEndpoint.UpdateClient()
			s.Require().NoError(err)

			tc.malleate(packet)

			// Validate funds are still not returned to the sender
			postTimeoutBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
			s.Require().Equal(postSendBalance.Amount, postTimeoutBalance.Amount)
			// Validate demand order created
			demandOrders, err = eibcKeeper.ListAllDemandOrders(s.hubCtx())
			s.Require().NoError(err)
			nOrdersCreated++
			s.Require().Equal(nOrdersCreated, len(demandOrders))
			// Get the last demand order created t
			lastDemandOrder := getLastDemandOrderByChannelAndSequence(demandOrders)
			// Validate the demand order price and denom
			fee := tc.fee(eibcKeeper.GetParams(s.hubCtx()))
			amountDec, err := math.LegacyNewDecFromStr(coinToSendToB.Amount.String())
			s.Require().NoError(err)
			expectedPrice := amountDec.Mul(math.LegacyNewDec(1).Sub(fee)).TruncateInt()
			s.Require().Equal(expectedPrice, lastDemandOrder.Price[0].Amount)
			s.Require().Equal(coinToSendToB.Denom, lastDemandOrder.Price[0].Denom)
			// Fulfill the demand order
			msgFulfillDemandOrder := eibctypes.NewMsgFulfillOrder(fulfillerAccount.String(), lastDemandOrder.Id, lastDemandOrder.Fee[0].Amount.String())
			_, err = s.msgServer().FulfillOrder(s.hubCtx(), msgFulfillDemandOrder)
			s.Require().NoError(err)
			// Validate balances of fulfiller and sender are updated while the original recipient is not
			fulfillerAccountBalance := bankKeeper.GetBalance(s.hubCtx(), fulfillerAccount, sdk.DefaultBondDenom)
			senderAccountBalance := bankKeeper.GetBalance(s.hubCtx(), senderAccount, sdk.DefaultBondDenom)
			receiverAccountBalance := bankKeeper.GetBalance(s.hubCtx(), receiverAccount, sdk.DefaultBondDenom)
			s.Require().True(fulfillerAccountBalance.IsEqual(fulfillerInitialBalance.Sub(lastDemandOrder.Price[0])))
			s.Require().True(senderAccountBalance.IsEqual(senderInitialBalance.Sub(lastDemandOrder.Fee[0])))
			s.Require().True(receiverAccountBalance.IsEqual(receiverInitialBalance))
			// Finalize the rollapp state
			currentRollappBlockHeight := uint64(s.rollappCtx().BlockHeight())
			_, err = s.finalizeRollappState(1, currentRollappBlockHeight)
			s.Require().NoError(err)
			// manually finalize packets through x/delayedack
			s.finalizePacketsByAddr(fulfillerAccount.String())
			// Funds are passed to the fulfiller
			fulfillerAccountBalanceAfterTimeout := bankKeeper.GetBalance(s.hubCtx(), fulfillerAccount, sdk.DefaultBondDenom)
			s.Require().True(fulfillerAccountBalanceAfterTimeout.IsEqual(fulfillerInitialBalance.Add(lastDemandOrder.Fee[0])))
		})
	}
}

/* -------------------------------------------------------------------------- */
/*                                    Utils                                   */
/* -------------------------------------------------------------------------- */

// transferRollappToHub sends a transfer packet from rollapp to hub and returns the packet
func (s *eibcSuite) transferRollappToHub(
	path *ibctesting.Path,
	sender string,
	receiver string,
	amountS string,
	memo string,
	expectAck bool,
) channeltypes.Packet {
	rolEnd := path.EndpointB

	hubIBCK := s.hubChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amt, ok := math.NewIntFromString(amountS)
	s.Require().True(ok)
	c := sdk.NewCoin(sdk.DefaultBondDenom, amt)

	msg := types.NewMsgTransfer(rolEnd.ChannelConfig.PortID, rolEnd.ChannelID,
		c, sender, receiver, timeoutHeight, 0, memo)
	res, err := s.rollappChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = path.RelayPacket(packet)

	expectErr := !expectAck // relay will throw an error if there is no ack
	if expectErr {
		s.Require().Error(err)
		found := hubIBCK.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		s.Require().False(found)
	} else {
		s.Require().NoError(err)
		found := hubIBCK.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		s.Require().True(found)
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
