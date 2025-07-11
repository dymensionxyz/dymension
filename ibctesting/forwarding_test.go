package ibctesting_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	comettypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
	ibccompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"
)

type eibcForwardSuite struct {
	eibcSuite
}

func TestForwardSuite(t *testing.T) {
	suite.Run(t, new(eibcForwardSuite))
}

func (s *eibcForwardSuite) SetupTest() {
	s.eibcSuite.SetupTest()
}

type mockTransferCompletionHook struct {
	called   bool
	checkBal bool
	s        *ibcTestingSuite
}

func (h *mockTransferCompletionHook) ValidateArg(hookData []byte) error {
	return nil
}

func (h *mockTransferCompletionHook) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {
	h.called = true
	if !h.checkBal {
		return nil
	}
	balances, err := h.s.hubApp().BankKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: fundsSource.String(),
		Denom:   budget.Denom,
	})
	h.s.Require().NoError(err)
	h.s.Require().Equal(balances, sdk.NewCoins(budget))
	return nil
}

func (s *eibcForwardSuite) TestFulfillHookIsCalled() {
	dummy := "dummy"
	h := mockTransferCompletionHook{
		s: &s.ibcTestingSuite,
	}
	s.hubApp().DelayedAckKeeper.SetCompletionHooks(
		map[string]delayedackkeeper.CompletionHookInstance{
			dummy: &h,
		},
	)
	s.T().Log("running test forward!")

	hookData := commontypes.CompletionHookCall{
		Name: dummy,
		Data: []byte{},
	}
	bz, err := proto.Marshal(&hookData)
	s.Require().NoError(err)
	s.eibcTransferFulfillment([]eibcTransferFulfillmentTC{
		{
			name:              "forwarding works",
			fulfillerStartBal: "300",
			eibcFee:           "150",
			transferAmt:       "200",
			fulfillHook:       bz,
		},
	})
	s.Require().True(h.called)
}

type FinalizeFwdTC struct {
	bridgeFee      int64 // percentage
	forwardChannel string
	ibcAmt         string
	expectOK       bool
}

var FinalizeFwdTCOK = FinalizeFwdTC{
	bridgeFee:      1,
	forwardChannel: "channel-0",
	ibcAmt:         "200",
	expectOK:       true,
}

func (s *eibcForwardSuite) TestFinalizeRolToRolOK() {
	tc := FinalizeFwdTCOK
	s.runFinalizeFwdTC(tc)
}

func (s *eibcForwardSuite) TestFinalizeRolToRolWrongChan() {
	tc := FinalizeFwdTCOK
	tc.forwardChannel = "channel-999"
	tc.expectOK = false
	s.runFinalizeFwdTC(tc)
}

func (s *eibcForwardSuite) runFinalizeFwdTC(tc FinalizeFwdTC) {
	p := s.dackK().GetParams(s.hubCtx())
	p.BridgingFee = math.LegacyNewDecWithPrec(tc.bridgeFee, 2) // 1%
	s.dackK().SetParams(s.hubCtx(), p)
	hookPayload := forwardtypes.NewHookForwardToIBC(
		tc.forwardChannel,
		"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp",
		uint64(time.Now().Add(time.Minute*5).UnixNano()), //nolint:gosec
	)
	err := hookPayload.ValidateBasic()
	s.Require().NoError(err)
	hook, err := forwardtypes.NewHookForwardToIBCCall(hookPayload)
	s.Require().NoError(err)
	hookBz, err := proto.Marshal(hook)
	s.Require().NoError(err)

	ibcRecipient := s.hubChain().SenderAccounts[0].SenderAccount.GetAddress()
	ibcRecipientBalBefore := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), ibcRecipient)

	s.rollappChain().NextBlock()
	rolH := uint64(s.rollappCtx().BlockHeight()) //nolint:gosec
	s.updateRollappState(rolH)

	eibcFee := "100" // arbitrary, shouldn't have an effect because we don't fulfil
	memo := delayedacktypes.CreateMemo(eibcFee, hookBz)
	packet := s.transferRollappToHub(s.path, s.rollappSender(), ibcRecipient.String(), tc.ibcAmt, memo, false)
	s.Require().True(s.rollappHasPacketCommitment(packet))

	rolH = uint64(s.rollappCtx().BlockHeight()) //nolint:gosec
	_, err = s.finalizeRollappState(1, rolH)
	s.Require().NoError(err)
	evts := s.finalizeRollappPacketsByAddress(ibcRecipient.String())

	_, err = ibctesting.ParseAckFromEvents(evts.ToABCIEvents())
	s.Require().NoError(err)

	ok, err := parseFwdErrFromEvents(evts.ToABCIEvents())
	s.Require().NoError(err)

	ibcRecipientBalAfter := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), ibcRecipient)
	if tc.expectOK {
		s.Require().True(ok)
		// no change, all the funds are used for forwarding!
		s.Require().Equal(ibcRecipientBalBefore, ibcRecipientBalAfter)
	} else {
		s.Require().False(ok)
		// recipient still has funds
		extra, _ := math.NewIntFromString(tc.ibcAmt)
		extra = extra.Sub(s.dackK().BridgingFeeFromAmt(s.hubCtx(), extra))
		ibcDenom := "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878" // found in debugger :/
		extraCoin := sdk.NewCoin(ibcDenom, extra)
		s.Require().Equal(ibcRecipientBalBefore.Add(extraCoin), ibcRecipientBalAfter)
	}
}

const (
	ForwardEvtTypeForward = "dymensionxyz.dymension.forward.EventForward"
	ForwardEvtAttrOK      = "ok"
)

func parseFwdErrFromEvents(events []comettypes.Event) (bool, error) {
	for _, ev := range events {
		if ev.Type == ForwardEvtTypeForward {
			for _, attr := range ev.Attributes {
				if attr.Key == ForwardEvtAttrOK {
					ok, err := strconv.ParseBool(attr.Value)
					return ok, err
				}
			}
		}
	}
	return false, gerrc.ErrNotFound
}

type osmosisForwardSuite struct {
	ibcTestingSuite
	path *ibctesting.Path
}

func TestOsmosisForwardSuite(t *testing.T) {
	suite.Run(t, new(osmosisForwardSuite))
}

func (s *osmosisForwardSuite) TestForward() {
	cosmosEndpoint := s.path.EndpointB

	hubIBCKeeper := s.hubChain().App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amount, ok := math.NewIntFromString("10000000000000000000") // 10DYM
	s.Require().True(ok)

	coinToSendToB := sdk.NewCoin("foo", amount)
	apptesting.FundAccount(s.hubApp(), s.cosmosCtx(), s.cosmosChain().SenderAccount.GetAddress(), sdk.NewCoins(coinToSendToB))

	dummy := "dummy"
	h := mockTransferCompletionHook{
		s: &s.ibcTestingSuite,
	}
	s.hubApp().DelayedAckKeeper.SetCompletionHooks(
		map[string]delayedackkeeper.CompletionHookInstance{
			dummy: &h,
		},
	)
	s.T().Log("running test forward!")

	hookData := commontypes.CompletionHookCall{
		Name: dummy,
		Data: []byte{},
	}
	bz, err := proto.Marshal(&hookData)
	s.Require().NoError(err)

	memo, err := ibccompletiontypes.MakeMemo(bz)
	s.Require().NoError(err)
	msg := types.NewMsgTransfer(
		cosmosEndpoint.ChannelConfig.PortID,
		cosmosEndpoint.ChannelID,
		coinToSendToB,
		s.cosmosChain().SenderAccount.GetAddress().String(),
		s.hubChain().SenderAccount.GetAddress().String(),
		timeoutHeight,
		0,
		memo,
	)
	res, err := s.cosmosChain().SendMsgs(msg)
	s.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = s.path.RelayPacket(packet)
	s.Require().NoError(err) // relay committed

	found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(s.hubCtx(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	s.Require().True(found)

	s.Require().True(h.called)
}

func (s *osmosisForwardSuite) SetupTest() {
	s.ibcTestingSuite.SetupTest()
	s.hubApp().LightClientKeeper.SetEnabled(false)

	s.hubApp().BankKeeper.SetDenomMetaData(s.hubCtx(), banktypes.Metadata{
		Base: sdk.DefaultBondDenom,
	})

	s.path = s.newTransferPath(s.hubChain(), s.cosmosChain())
	s.coordinator.Setup(s.path)
}
