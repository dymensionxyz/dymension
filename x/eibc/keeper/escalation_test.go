package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	dacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// Creation clamps maxFee to amount-bridgingFee so the price floor stays >= 0.
func (suite *KeeperTestSuite) TestCreateDemandOrderEscalationClamp() {
	dackParams := dacktypes.NewParams("hour", math.LegacyNewDecWithPrec(1, 2), 0) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)

	// amount 1000, 1% bridging => bridging fee 10. maxFee clamped to
	// amount-bridgingFee-1 = 989 so the price floor stays >= 1.
	pd := transfertypes.NewFungibleTokenPacketData(
		sdk.DefaultBondDenom, "1000", eibcSenderAddr.String(), eibcReceiverAddr.String(),
		`{"eibc":{"fee":"100","fee_max":"100000","fee_escalation_blocks":20}}`,
	)
	pkt := channeltypes.NewPacket(pd.GetBytes(), 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	rp := *rollappPacket
	rp.Packet = &pkt
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, rp)

	order, err := suite.App.EIBCKeeper.CreateDemandOrderOnRecv(suite.Ctx, pd, &rp)
	suite.Require().NoError(err)
	suite.Require().NotNil(order.FeeEscalation)
	suite.Require().Equal(math.NewInt(989), order.FeeEscalation.MaxFeeAmount)
	suite.Require().Equal(uint64(20), order.FeeEscalation.DurationBlocks)
	suite.Require().Equal(math.NewInt(100), order.GetFeeAmount())

	// At/after the window end the effective price must stay strictly positive and
	// EffectiveFeePercent must not panic (division by zero).
	suite.Require().True(order.EffectivePriceAmount(order.CreationHeight + 20).IsPositive())
	suite.Require().NotPanics(func() { _ = order.EffectiveFeePercent(order.CreationHeight + 20) })
	suite.Require().NotPanics(func() { _ = order.EffectiveFeePercent(order.CreationHeight + 9999) })
}

// MsgFulfillOrder on an escalating order: succeeds when expected_fee <= effectiveFee(H),
// recipient receives effectivePrice(H); fails when expected_fee > effectiveFee(H).
func (suite *KeeperTestSuite) TestMsgFulfillOrderEscalation() {
	// base fee=10, price=90; escalate to maxFee=110 over 10 blocks from creation 10.
	// at height 15 (elapsed 5): effFee=60, effPrice=40.
	mkOrder := func() *types.DemandOrder {
		o := types.NewDemandOrder(*rollappPacket, math.NewInt(90), math.NewInt(10), sdk.DefaultBondDenom,
			sample.AccAddress(), 10, nil, &types.FeeEscalation{MaxFeeAmount: math.NewInt(110), DurationBlocks: 10})
		return o
	}

	suite.Run("expected fee too high - fails", func() {
		suite.SetupTest()
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		o := mkOrder()
		suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, o))
		suite.Ctx = suite.Ctx.WithBlockHeight(15)
		fulfiller := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0]
		_, err := suite.msgServer.FulfillOrder(suite.Ctx, types.NewMsgFulfillOrder(fulfiller.String(), o.Id, "61"))
		suite.Require().ErrorIs(err, types.ErrExpectedFeeNotMet)
	})

	suite.Run("expected fee at/below effective - succeeds, recipient gets effective price", func() {
		suite.SetupTest()
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		o := mkOrder()
		recipient := o.GetRecipientBech32Address()
		suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, o))
		suite.Ctx = suite.Ctx.WithBlockHeight(15)
		fulfiller := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0]
		recipientBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, recipient, sdk.DefaultBondDenom).Amount

		_, err := suite.msgServer.FulfillOrder(suite.Ctx, types.NewMsgFulfillOrder(fulfiller.String(), o.Id, "60"))
		suite.Require().NoError(err)

		recipientAfter := suite.App.BankKeeper.GetBalance(suite.Ctx, recipient, sdk.DefaultBondDenom).Amount
		suite.Require().Equal(math.NewInt(40), recipientAfter.Sub(recipientBefore)) // effective price

		got, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_PENDING, o.Id)
		suite.Require().NoError(err)
		suite.Require().Nil(got.FeeEscalation)
		suite.Require().Equal(math.NewInt(60), got.GetFeeAmount())
		suite.Require().Equal(math.NewInt(40), got.PriceAmount())
	})
}

// MsgFulfillOrderAuthorized on an escalating order: succeeds when effectivePrice(H) <= msg.Price
// and effectiveFee(H) >= msg.ExpectedFee; operator fee split computed from effectiveFee(H).
func (suite *KeeperTestSuite) TestMsgFulfillOrderAuthorizedEscalation() {
	// base fee=10, price=90; maxFee=110 over 10 blocks from creation 10.
	// at height 15: effFee=60, effPrice=40. operator share 0.2 => operatorFee=12.
	denom := "adym"
	lpAddr := sample.AccAddress()
	opAddr := sample.AccAddress()

	suite.SetupTest()
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	o := types.NewDemandOrder(*rollappPacket, math.NewInt(90), math.NewInt(10), denom,
		sample.AccAddress(), 10, nil, &types.FeeEscalation{MaxFeeAmount: math.NewInt(110), DurationBlocks: 10})
	recipient := o.GetRecipientBech32Address()
	suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, o))

	lp := sdk.MustAccAddressFromBech32(lpAddr)
	op := sdk.MustAccAddressFromBech32(opAddr)
	suite.Require().NoError(bankutil.FundAccount(suite.Ctx, suite.App.BankKeeper, lp, sdk.NewCoins(sdk.NewInt64Coin(denom, 200))))
	suite.Require().NoError(bankutil.FundAccount(suite.Ctx, suite.App.BankKeeper, op, sdk.NewCoins(sdk.NewInt64Coin(denom, 50))))

	suite.Ctx = suite.Ctx.WithBlockHeight(15)
	msg := &types.MsgFulfillOrderAuthorized{
		OrderId:            o.Id,
		RollappId:          rollappPacket.RollappId,
		Price:              sdk.NewCoins(sdk.NewInt64Coin(denom, 50)), // LP willing to pay up to 50 >= 40
		Amount:             math.NewInt(100),
		ExpectedFee:        "55", // LP wants at least 55 <= 60
		OperatorFeeShare:   math.LegacyNewDecWithPrec(2, 1),
		OperatorFeeAddress: opAddr,
		LpAddress:          lpAddr,
	}
	_, err := suite.msgServer.FulfillOrderAuthorized(suite.Ctx, msg)
	suite.Require().NoError(err)

	// recipient receives effective price 40; operator fee = 60 * 0.2 = 12 from LP.
	suite.Require().Equal(math.NewInt(40), suite.App.BankKeeper.GetBalance(suite.Ctx, recipient, denom).Amount)
	suite.Require().Equal(math.NewInt(200-40-12), suite.App.BankKeeper.GetBalance(suite.Ctx, lp, denom).Amount)
	suite.Require().Equal(math.NewInt(50+12), suite.App.BankKeeper.GetBalance(suite.Ctx, op, denom).Amount)

	// fail: expected fee above effective fee at this height
	suite.Run("expected fee above effective - fails", func() {
		suite.SetupTest()
		suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
		o2 := types.NewDemandOrder(*rollappPacket, math.NewInt(90), math.NewInt(10), denom,
			sample.AccAddress(), 10, nil, &types.FeeEscalation{MaxFeeAmount: math.NewInt(110), DurationBlocks: 10})
		suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, o2))
		suite.Ctx = suite.Ctx.WithBlockHeight(15)
		msg2 := *msg
		msg2.OrderId = o2.Id
		msg2.ExpectedFee = "70" // > effFee 60
		_, err := suite.msgServer.FulfillOrderAuthorized(suite.Ctx, &msg2)
		suite.Require().Error(err)
	})
}

// An order whose base fee% is below an LP's minFee is rejected at creation height but
// accepted (and fulfilled) at a later height once EffectiveFeePercent crosses minFee;
// lp.Spent increases by the effective price.
func (suite *KeeperTestSuite) TestOnDemandLPEscalation() {
	denom := sdk.DefaultBondDenom
	k := suite.App.EIBCKeeper

	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, math.NewInt(10_000))
	orderAddr := addrs[0]
	lpFundsAddr := addrs[1]

	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// base fee=10, price=90, creation 10; maxFee=60 over 10 blocks.
	// fee% rises 10/90=0.11 -> ... ; crosses 0.5 at elapsed 5 (fee=35, price=65, 0.538).
	o := types.NewDemandOrder(*rollappPacket, math.NewInt(90), math.NewInt(10), denom, orderAddr.String(), 10, nil,
		&types.FeeEscalation{MaxFeeAmount: math.NewInt(60), DurationBlocks: 10})
	suite.Require().NoError(k.SetDemandOrder(suite.Ctx, o))

	lpID, err := k.LPs.Create(suite.Ctx, &types.OnDemandLP{
		FundsAddr:  lpFundsAddr.String(),
		Rollapp:    rollappPacket.RollappId,
		Denom:      denom,
		MaxPrice:   math.NewInt(1000),
		MinFee:     math.LegacyMustNewDecFromStr("0.5"),
		SpendLimit: math.NewInt(1000),
	})
	suite.Require().NoError(err)

	// At elapsed 4 (height 14): fee%=30/70=0.428 < 0.5 -> not compatible.
	lps, err := k.LPs.GetOrderCompatibleLPs(suite.Ctx.WithBlockHeight(14), *o)
	suite.Require().NoError(err)
	suite.Require().Empty(lps)

	// At elapsed 5 (height 15): fee%=0.538 >= 0.5 -> compatible and fulfillable.
	suite.Ctx = suite.Ctx.WithBlockHeight(15)
	orderBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, orderAddr, denom).Amount
	lpBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, lpFundsAddr, denom).Amount

	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, o.Id, 0))

	// recipient receives effective price 65, LP pays 65.
	suite.Require().Equal(math.NewInt(65), suite.App.BankKeeper.GetBalance(suite.Ctx, orderAddr, denom).Amount.Sub(orderBefore))
	suite.Require().Equal(math.NewInt(65), lpBefore.Sub(suite.App.BankKeeper.GetBalance(suite.Ctx, lpFundsAddr, denom).Amount))

	rec, err := k.LPs.Get(suite.Ctx, lpID)
	suite.Require().NoError(err)
	suite.Require().Equal(math.NewInt(65), rec.Spent) // effective price
}

// Manual update clears escalation and makes the static fee/price authoritative.
func (suite *KeeperTestSuite) TestMsgUpdateDemandOrderClearsEscalation() {
	dackParams := dacktypes.NewParams("hour", math.LegacyNewDecWithPrec(1, 2), 0) // 1%
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, dackParams)
	denom := sdk.DefaultBondDenom

	recipient := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(100_000))[0]
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)

	// amount 1000, 1% bridging fee => initial price = 1000 - 100 - 10 = 890.
	o := types.NewDemandOrder(*rollappPacket, math.NewInt(890), math.NewInt(100), denom, recipient.String(), 1, nil,
		&types.FeeEscalation{MaxFeeAmount: math.NewInt(500), DurationBlocks: 10})
	suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, o))

	_, err := suite.msgServer.UpdateDemandOrder(suite.Ctx, types.NewMsgUpdateDemandOrder(recipient.String(), o.Id, "400"))
	suite.Require().NoError(err)

	got, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_PENDING, o.Id)
	suite.Require().NoError(err)
	suite.Require().Nil(got.FeeEscalation)
	suite.Require().Equal(math.NewInt(400), got.GetFeeAmount())
	suite.Require().Equal(math.NewInt(590), got.PriceAmount()) // 1000 - 400 - 10
}
