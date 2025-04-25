package ibctesting_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
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

func (s *forwardSuite) TestFinalizeRolToRol() {

	type TC struct {
		bridgeFee      int // percentage
		forwardChannel string
		forwardAmt     int64
		forwardDst     string
		ibcAmt         string
		expectOK       bool
	}

	tc := TC{
		bridgeFee:      1,
		forwardChannel: "channel-0",
		forwardAmt:     100,
		forwardDst:     "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp",
		ibcAmt:         "200",
		expectOK:       true,
	}

	p := s.dackK().GetParams(s.hubCtx())
	p.BridgingFee = math.LegacyNewDecWithPrec(1, 2) // 1%
	s.dackK().SetParams(s.hubCtx(), p)
	ibcDenom := "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878" // found in debugger :/
	hookPayload := forwardtypes.MakeHookForwardToIBC(
		tc.forwardChannel,
		sdk.NewCoin(ibcDenom, math.NewInt(tc.forwardAmt)),
		tc.forwardDst,
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

	memo := delayedacktypes.CreateMemo("100", hookBz)
	packet := s.transferRollappToHub(s.path, s.rollappSender(), ibcRecipient.String(), tc.ibcAmt, memo, false)
	s.Require().True(s.rollappHasPacketCommitment(packet))

	// Finalize rollapp and check fulfiller balance was updated with fee
	rolH = uint64(s.rollappCtx().BlockHeight())
	_, err = s.finalizeRollappState(1, rolH)
	s.Require().NoError(err)
	evts := s.finalizePacketsByAddr(ibcRecipient.String())

	_, err = ibctesting.ParseAckFromEvents(evts.ToABCIEvents())
	s.Require().NoError(err)

	_ = ibcRecipientBalBefore
}
