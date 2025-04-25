package ibctesting_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
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

func (s *forwardSuite) TestFulfillRolToRol() {
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

	hook := forwardtypes.MakeHookForwardToIBC(
		"transfer",
		sdk.NewCoin("uatom", math.NewInt(100)),
		"cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgp",
		uint64(time.Now().Add(time.Minute*5).UnixNano()),
	)

	hookData := commontypes.CompletionHookCall{
		Name: forwardtypes.HookNameRollToIBC,
		Data: []byte{},
	}
	bz, err := proto.Marshal(&hookData)
	s.Require().NoError(err)
	s.eibcTransferFulfillment([]eibcTransferFulfillmentTC{
		{
			name:              "rol to rol works",
			fulfillerStartBal: "300",
			eibcFee:           "150",
			transferAmt:       "200",
			fulfillHook:       bz,
		},
	})
	s.Require().True(h.called)
}
