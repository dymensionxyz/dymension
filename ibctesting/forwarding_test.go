package ibctesting_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	comettypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"
)

type forwardSuite struct {
	eibcSuite
}

func TestForwardSuite(t *testing.T) {
	suite.Run(t, new(forwardSuite))
}

func (s *forwardSuite) SetupTest() {
	s.eibcSuite.SetupTest()
}

type mockTransferCompletionHook struct {
	*forwardSuite
	called bool
}

func (h *mockTransferCompletionHook) ValidateArg(hookData []byte) error {
	return nil
}

func (h *mockTransferCompletionHook) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {
	h.called = true
	return nil
}

func (s *forwardSuite) TestFulfillHookIsCalled() {
	dummy := "dummy"
	h := mockTransferCompletionHook{
		forwardSuite: s,
	}
	s.utilSuite.hubApp().DelayedAckKeeper.SetCompletionHooks(
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
	forwardAmt     int64
	ibcAmt         string
	expectOK       bool
}

var FinalizeFwdTCOK = FinalizeFwdTC{
	bridgeFee:      1,
	forwardChannel: "channel-0",
	forwardAmt:     100,
	ibcAmt:         "200",
	expectOK:       true,
}

func (s *forwardSuite) TestFinalizeRolToRolOK() {
	tc := FinalizeFwdTCOK
	s.runFinalizeFwdTC(tc)
}
func (s *forwardSuite) TestFinalizeRolToRolWrongChan() {
	tc := FinalizeFwdTCOK
	tc.forwardChannel = "channel-999"
	tc.expectOK = false
	s.runFinalizeFwdTC(tc)
}

func (s *forwardSuite) runFinalizeFwdTC(tc FinalizeFwdTC) {
	p := s.dackK().GetParams(s.hubCtx())
	p.BridgingFee = math.LegacyNewDecWithPrec(tc.bridgeFee, 2) // 1%
	s.dackK().SetParams(s.hubCtx(), p)
	ibcDenom := "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878" // found in debugger :/
	hookPayload := forwardtypes.MakeHookForwardToIBC(
		tc.forwardChannel,
		sdk.NewCoin(ibcDenom, math.NewInt(tc.forwardAmt)),
		"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp",
		uint64(time.Now().Add(time.Minute*5).UnixNano()),
	)
	hook, err := forwardtypes.NewRollToIBCHook(hookPayload)
	s.Require().NoError(err)
	hookBz, err := proto.Marshal(hook)
	s.Require().NoError(err)

	ibcRecipient := s.hubChain().SenderAccounts[0].SenderAccount.GetAddress()
	ibcRecipientBalBefore := s.hubApp().BankKeeper.SpendableCoins(s.hubCtx(), ibcRecipient)

	s.rollappChain().NextBlock()
	rolH := uint64(s.rollappCtx().BlockHeight())
	s.updateRollappState(rolH)

	eibcFee := "100" // arbitrary, shouldn't have an effect because we don't fulfil
	memo := delayedacktypes.CreateMemo(eibcFee, hookBz)
	packet := s.transferRollappToHub(s.path, s.rollappSender(), ibcRecipient.String(), tc.ibcAmt, memo, false)
	s.Require().True(s.rollappHasPacketCommitment(packet))

	rolH = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, rolH)
	s.Require().NoError(err)
	evts := s.finalizePacketsByAddr(ibcRecipient.String())

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
		extraCoin := sdk.NewCoin(ibcDenom, extra)
		s.Require().Equal(ibcRecipientBalBefore.Add(extraCoin), ibcRecipientBalAfter)
	}

}

func parseFwdErrFromEvents(events []comettypes.Event) (bool, error) {
	for _, ev := range events {
		if ev.Type == forwardtypes.EvtTypeForward {
			for _, attr := range ev.Attributes {
				if attr.Key == forwardtypes.EvtAttrOK {
					ok, err := strconv.ParseBool(attr.Value)
					return ok, err
				}
			}
		}
	}
	return false, gerrc.ErrNotFound
}
