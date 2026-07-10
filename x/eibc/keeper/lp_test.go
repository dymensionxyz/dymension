package keeper_test

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	agentkeeper "github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	agenttypes "github.com/dymensionxyz/dymension/v3/x/agent/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// create some lps and for a given order, find the compatible lps
func (suite *KeeperTestSuite) TestLPFindCompatible() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "1", // wrong rollup
		Denom:      "aaa",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("0"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "1", // wrong rollup
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("0"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "aaa", // wrong denom
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("0"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1), // max price too low
		MinFee:     math.LegacyMustNewDecFromStr("0"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(5),
		MinFee:     math.LegacyMustNewDecFromStr("0.8"), // min fee too high
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	expect, err := k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(5), // valid
		MinFee:     math.LegacyMustNewDecFromStr("0.2"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	expect1, err := k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(6), // also valid, but not first
		MinFee:     math.LegacyMustNewDecFromStr("0.2"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "3", // wrong rollup
		Denom:      "aaa",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("0.2"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)

	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "3", // wrong rollup
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("0.2"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	o := types.DemandOrder{
		Price:     sdk.NewCoins(sdk.NewCoin("bbb", math.NewInt(5))),
		Fee:       sdk.NewCoins(sdk.NewCoin("bbb", math.NewInt(3))),
		RollappId: "2",
	}
	lps, err := k.LPs.GetOrderCompatibleLPs(ctx, o)
	suite.Require().NoError(err)
	suite.Equal(2, len(lps))
	suite.Equal(expect, lps[0].Id)
	suite.Equal(expect1, lps[1].Id)
}

// sanity check because another test was failing so wanted to double check
func (suite *KeeperTestSuite) TestDebug() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	denom := sdk.DefaultBondDenom
	rol := "foo_224126-1"
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    rol, // wrong rollup
		Denom:      denom,
		MaxPrice:   math.NewInt(1),
		MinFee:     math.LegacyMustNewDecFromStr("1"),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	compat, err := k.LPs.GetOrderCompatibleLPs(ctx, types.DemandOrder{
		RollappId: rol,
		Price:     sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(1))),
		Fee:       sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(1))),
	})
	suite.Require().NoError(err)
	suite.Require().Len(compat, 1)
}

// test the order age compatibility
// not practical due to test in other test as get explosion of combinations
func (suite *KeeperTestSuite) TestLPCompatibilityHeightAge() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:           "1", // wrong rollup
		Denom:             "aaa",
		SpendLimit:        math.NewInt(100),
		MaxPrice:          math.NewInt(100),
		OrderMinAgeBlocks: 50, // practical value of 5 mins with 6 secs per block, although doesn't matter for test
	})
	suite.Require().NoError(err)
	o := types.DemandOrder{
		RollappId:      "1",
		Price:          sdk.NewCoins(sdk.NewCoin("aaa", math.NewInt(7))),
		Fee:            sdk.NewCoins(sdk.NewCoin("aaa", math.NewInt(7))),
		CreationHeight: 20,
	}
	for i := 68; i < 72; i++ {
		lps, err := k.LPs.GetOrderCompatibleLPs(ctx.WithBlockHeight(int64(i)), o)
		suite.Require().NoError(err)
		if i < 70 {
			suite.Empty(lps)
		} else {
			suite.NotEmpty(lps)
		}
	}
}

// orderWithSeq builds and stores a distinct outstanding demand order (unique
// packet sequence => unique id) and returns its id.
func (suite *KeeperTestSuite) orderWithSeq(seq uint64, recipient string, price, fee math.Int) string {
	pkt := *rollappPacket
	innerPkt := *pkt.Packet
	innerPkt.Sequence = seq
	pkt.Packet = &innerPkt
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, pkt)
	order := types.NewDemandOrder(pkt, price, fee, sdk.DefaultBondDenom, recipient, 1, nil, nil)
	suite.Require().NoError(suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, order))
	return order.Id
}

// LP with a velocity cap: cumulative fills up to the cap within a bucket
// succeed; the next fill that would exceed it is rejected; advancing into a new
// bucket resets capacity.
func (suite *KeeperTestSuite) TestLPVelocityCap() {
	k := suite.App.EIBCKeeper
	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, math.NewInt(10_000_000))
	orderAddr, lpAddr := addrs[0], addrs[1]

	suite.Ctx = suite.Ctx.WithBlockHeight(5) // bucket 0 (window of 10)
	_, err := k.LPs.Create(suite.Ctx, &types.OnDemandLP{
		FundsAddr:       lpAddr.String(),
		Rollapp:         rollappPacket.RollappId,
		Denom:           sdk.DefaultBondDenom,
		MaxPrice:        math.NewInt(100),
		MinFee:          math.LegacyZeroDec(),
		SpendLimit:      math.NewInt(1000),
		RateLimitAmount: math.NewInt(100),
		RateLimitBlocks: 10,
	})
	suite.Require().NoError(err)

	// fill 60 within the cap
	id1 := suite.orderWithSeq(1, orderAddr.String(), math.NewInt(60), math.NewInt(10))
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id1, 0))

	// next fill of 60 would bring window to 120 > 100: rejected (only lp filtered out)
	id2 := suite.orderWithSeq(2, orderAddr.String(), math.NewInt(60), math.NewInt(10))
	err = k.FulfillByOnDemandLP(suite.Ctx, id2, 0)
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrNotFound), "expected no compatible lp, got %v", err)

	// a fill that stays within remaining 40 still succeeds in the same bucket
	id3 := suite.orderWithSeq(3, orderAddr.String(), math.NewInt(40), math.NewInt(10))
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id3, 0))

	// advance into the next bucket (height 12 => bucket 10): capacity resets
	suite.Ctx = suite.Ctx.WithBlockHeight(12)
	id4 := suite.orderWithSeq(4, orderAddr.String(), math.NewInt(90), math.NewInt(10))
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id4, 0))
}

// LP with a validity window is matched below the expiry height and not at/after.
func (suite *KeeperTestSuite) TestLPValidityWindowFulfill() {
	k := suite.App.EIBCKeeper
	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, math.NewInt(10_000_000))
	orderAddr, lpAddr := addrs[0], addrs[1]

	suite.Ctx = suite.Ctx.WithBlockHeight(50)
	_, err := k.LPs.Create(suite.Ctx, &types.OnDemandLP{
		FundsAddr:        lpAddr.String(),
		Rollapp:          rollappPacket.RollappId,
		Denom:            sdk.DefaultBondDenom,
		MaxPrice:         math.NewInt(100),
		MinFee:           math.LegacyZeroDec(),
		SpendLimit:       math.NewInt(1000),
		ValidUntilHeight: 100,
	})
	suite.Require().NoError(err)

	// below expiry: fulfilled
	suite.Ctx = suite.Ctx.WithBlockHeight(99)
	id1 := suite.orderWithSeq(1, orderAddr.String(), math.NewInt(40), math.NewInt(10))
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id1, 0))

	// at expiry (exclusive): not matched
	suite.Ctx = suite.Ctx.WithBlockHeight(100)
	id2 := suite.orderWithSeq(2, orderAddr.String(), math.NewInt(40), math.NewInt(10))
	err = k.FulfillByOnDemandLP(suite.Ctx, id2, 0)
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrNotFound), "expected no compatible lp, got %v", err)
}

// CreateOnDemandLP rejects an LP whose validity window is already in the past.
func (suite *KeeperTestSuite) TestCreateLPAlreadyExpired() {
	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(10_000_000))
	suite.Ctx = suite.Ctx.WithBlockHeight(100)
	msg := &types.MsgCreateOnDemandLP{
		Lp: &types.OnDemandLP{
			FundsAddr:        addrs[0].String(),
			Rollapp:          rollappPacket.RollappId,
			Denom:            sdk.DefaultBondDenom,
			MaxPrice:         math.NewInt(100),
			MinFee:           math.LegacyZeroDec(),
			SpendLimit:       math.NewInt(1000),
			ValidUntilHeight: 100,
		},
		Signer: addrs[0].String(),
	}
	_, err := suite.msgServer.CreateOnDemandLP(suite.Ctx, msg)
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrInvalidArgument), "got %v", err)
}

// LP bound to an active agent fulfills exactly like an unbound one. After
// DeactivateAgent (fleet kill-switch), the bound LP is skipped at match time
// while a co-existing unbound LP still fulfills.
func (suite *KeeperTestSuite) TestLPAgentBinding() {
	k := suite.App.EIBCKeeper
	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 3, math.NewInt(10_000_000))
	orderAddr, lpAddr, ownerAddr := addrs[0], addrs[1], addrs[2]

	suite.Require().NoError(suite.App.AgentKeeper.SetAgent(suite.Ctx, agenttypes.Agent{
		Id:     "agent1",
		Owner:  ownerAddr.String(),
		Active: true,
	}))

	lp := types.OnDemandLP{
		FundsAddr:  lpAddr.String(),
		Rollapp:    rollappPacket.RollappId,
		Denom:      sdk.DefaultBondDenom,
		MaxPrice:   math.NewInt(100),
		MinFee:     math.LegacyZeroDec(),
		SpendLimit: math.NewInt(1000),
		AgentId:    "agent1",
	}
	boundID, err := k.LPs.Create(suite.Ctx, &lp)
	suite.Require().NoError(err)

	// bound to an active agent: matched and fulfilled like an unbound LP
	id1 := suite.orderWithSeq(1, orderAddr.String(), math.NewInt(40), math.NewInt(10))
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id1, 0))

	// deactivate the agent: every bound LP is disabled with a single message
	agentMsgServer := agentkeeper.NewMsgServerImpl(*suite.App.AgentKeeper)
	_, err = agentMsgServer.DeactivateAgent(suite.Ctx, agenttypes.NewMsgDeactivateAgent(ownerAddr.String(), "agent1"))
	suite.Require().NoError(err)

	id2 := suite.orderWithSeq(2, orderAddr.String(), math.NewInt(40), math.NewInt(10))
	o, err := k.GetDemandOrder(suite.Ctx, commontypes.Status_PENDING, id2)
	suite.Require().NoError(err)
	compat, err := k.LPs.GetOrderCompatibleLPs(suite.Ctx, *o)
	suite.Require().NoError(err)
	suite.Require().Empty(compat)
	err = k.FulfillByOnDemandLP(suite.Ctx, id2, 0)
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrNotFound), "expected no compatible lp, got %v", err)

	// an unbound LP is unaffected by the deactivation
	unbound := lp
	unbound.AgentId = ""
	unboundID, err := k.LPs.Create(suite.Ctx, &unbound)
	suite.Require().NoError(err)
	compat, err = k.LPs.GetOrderCompatibleLPs(suite.Ctx, *o)
	suite.Require().NoError(err)
	suite.Require().Len(compat, 1)
	suite.Require().Equal(unboundID, compat[0].Id)
	suite.Require().NotEqual(boundID, compat[0].Id)
	suite.Require().NoError(k.FulfillByOnDemandLP(suite.Ctx, id2, 0))
}

// CreateOnDemandLP requires a referenced agent to exist, but not to be active.
func (suite *KeeperTestSuite) TestCreateLPAgentBinding() {
	addrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 2, math.NewInt(10_000_000))
	lpAddr, ownerAddr := addrs[0], addrs[1]

	suite.Require().NoError(suite.App.AgentKeeper.SetAgent(suite.Ctx, agenttypes.Agent{
		Id:     "agent1",
		Owner:  ownerAddr.String(),
		Active: false,
	}))

	msg := func(agentID string) *types.MsgCreateOnDemandLP {
		return &types.MsgCreateOnDemandLP{
			Lp: &types.OnDemandLP{
				FundsAddr:  lpAddr.String(),
				Rollapp:    rollappPacket.RollappId,
				Denom:      sdk.DefaultBondDenom,
				MaxPrice:   math.NewInt(100),
				MinFee:     math.LegacyZeroDec(),
				SpendLimit: math.NewInt(1000),
				AgentId:    agentID,
			},
			Signer: lpAddr.String(),
		}
	}

	_, err := suite.msgServer.CreateOnDemandLP(suite.Ctx, msg("nonexistent"))
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrNotFound), "got %v", err)

	// existing but inactive agent: creation is allowed, only fulfillment requires liveness
	_, err = suite.msgServer.CreateOnDemandLP(suite.Ctx, msg("agent1"))
	suite.Require().NoError(err)

	// unbound
	_, err = suite.msgServer.CreateOnDemandLP(suite.Ctx, msg(""))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestLPQueriesByAddr() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	addrs := []string{
		"dym1ra6le06p8lle3q6gnsmwz769t2kqld9pmden5k",
		"dym10j59k4whfvtu5flc3lypsjmcyx3fn57ygw78du",
	}
	for i := range 6 {
		_, err = k.LPs.Create(ctx, &types.OnDemandLP{
			Rollapp:   "1",
			Denom:     "bbb",
			FundsAddr: addrs[i%2],
		})
		suite.Require().NoError(err)
	}
	for _, addr := range addrs {
		lps, err := k.LPs.GetByAddr(ctx, sdk.MustAccAddressFromBech32(addr))
		suite.Require().NoError(err)
		suite.Equal(3, len(lps))
		for _, lp := range lps {
			suite.Equal(addr, lp.Lp.FundsAddr)
		}
	}
}
