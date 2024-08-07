package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_CancelSellOrder_DymName(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	// force enable trading
	moduleParams := dk.GetParams(ctx)
	moduleParams.Misc.EnableTradingName = true
	moduleParams.Misc.EnableTradingAlias = true
	require.NoError(t, dk.SetParams(ctx, moduleParams))

	msgServer := dymnskeeper.NewMsgServerImpl(dk)

	ownerA := testAddr(1).bech32()
	notOwnerA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	dymNames := dk.GetAllNonExpiredDymNames(ctx)
	require.Len(t, dymNames, 2)

	t.Run("do not process message that not pass basic validation", func(t *testing.T) {
		requireErrorFContains(t, func() error {
			resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
				GoodsId: "abc",
				Owner:   "0x1", // invalid owner
			})

			require.Nil(t, resp)

			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	t.Run("do not process message that refer to non-existing Dym-Name", func(t *testing.T) {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   "not-exists",
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "Dym-Name: not-exists: not found")
	})

	t.Run("do not process message that type is Unknown", func(t *testing.T) {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   "goods",
			OrderType: dymnstypes.OrderType_OT_UNKNOWN,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "invalid order type")
	})

	t.Run("do not process that owner does not match", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:   dymName1.Name,
			Type:      dymnstypes.NameOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.NameOrder,
			Owner:     notOwnerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "not the owner of the Dym-Name")
	})

	t.Run("do not process for Dym-Name that does not have any SO", func(t *testing.T) {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   dymName1.Name,
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), fmt.Sprintf("Sell-Order: %s: not found", dymName1.Name))
	})

	t.Run("can not cancel expired", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:   dymName1.Name,
			Type:      dymnstypes.NameOrder,
			ExpireAt:  1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "cannot cancel an expired order")
	})

	t.Run("can not cancel once bid placed", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:  dymName1.Name,
			Type:     dymnstypes.NameOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  dymnsutils.TestCoin(300),
			},
		}
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "cannot cancel once bid placed")
	})

	t.Run("can will remove the active SO expiration mapping record", func(t *testing.T) {
		aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.NameOrder)

		so11 := dymnstypes.SellOrder{
			GoodsId:  dymName1.Name,
			Type:     dymnstypes.NameOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)
		aSoe.Add(so11.GoodsId, so11.ExpireAt)

		so12 := dymnstypes.SellOrder{
			GoodsId:  dymName2.Name,
			Type:     dymnstypes.NameOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so12)
		require.NoError(t, err)
		aSoe.Add(so12.GoodsId, so12.ExpireAt)

		err = dk.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.NameOrder)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder)
			dk.DeleteSellOrder(ctx, so12.GoodsId, dymnstypes.NameOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Nil(t, dk.GetSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder), "SO should be removed from active")

		aSoe = dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.NameOrder)

		allNames := make(map[string]bool)
		for _, record := range aSoe.Records {
			allNames[record.GoodsId] = true
		}
		require.NotContains(t, allNames, so11.GoodsId)
		require.Contains(t, allNames, so12.GoodsId)
	})

	t.Run("can cancel if satisfied conditions", func(t *testing.T) {
		moduleParams := dk.GetParams(ctx)
		moduleParams.Misc.EnableTradingName = false // allowed to cancel even if trading is disabled
		require.NoError(t, dk.SetParams(ctx, moduleParams))

		so11 := dymnstypes.SellOrder{
			GoodsId:  dymName1.Name,
			Type:     dymnstypes.NameOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		so12 := dymnstypes.SellOrder{
			GoodsId:  dymName2.Name,
			Type:     dymnstypes.NameOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so12)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder)
			dk.DeleteSellOrder(ctx, so12.GoodsId, dymnstypes.NameOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.NameOrder,
			Owner:     ownerA,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Nil(t, dk.GetSellOrder(ctx, so11.GoodsId, dymnstypes.NameOrder), "SO should be removed from active")
		require.NotNil(t, dk.GetSellOrder(ctx, dymName2.Name, dymnstypes.NameOrder), "other records remaining as-is")

		list := dk.GetHistoricalSellOrders(ctx, so11.GoodsId, dymnstypes.NameOrder)
		require.Empty(t, list, "no historical record should be added")

		require.GreaterOrEqual(t,
			ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCloseSellOrder,
			"should consume params gas",
		)
	})
}

//goland:noinspection GoSnakeCaseUsage
func Test_msgServer_CancelSellOrder_Alias(t *testing.T) {
	now := time.Now().UTC()

	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	// force enable trading
	moduleParams := dk.GetParams(ctx)
	moduleParams.Misc.EnableTradingName = true
	moduleParams.Misc.EnableTradingAlias = true
	require.NoError(t, dk.SetParams(ctx, moduleParams))

	msgServer := dymnskeeper.NewMsgServerImpl(dk)

	ownerA := testAddr(1).bech32()
	anotherA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	type rollapp struct {
		rollAppId string
		creator   string
		alias     string
	}

	rollapp_1_ofOwner := rollapp{
		rollAppId: "rollapp_1-1",
		creator:   ownerA,
		alias:     "one",
	}
	rollapp_2_ofOwner := rollapp{
		rollAppId: "rollapp_2-1",
		creator:   ownerA,
		alias:     "two",
	}
	rollapp_3_ofAnother := rollapp{
		rollAppId: "rollapp_3-1",
		creator:   anotherA,
		alias:     "three",
	}
	rollapp_4_ofBidder := rollapp{
		rollAppId: "rollapp_4-2",
		creator:   bidderA,
		alias:     "",
	}
	for _, ra := range []rollapp{rollapp_1_ofOwner, rollapp_2_ofOwner, rollapp_3_ofAnother, rollapp_4_ofBidder} {
		rk.SetRollapp(ctx, rollapptypes.Rollapp{
			RollappId: ra.rollAppId,
			Owner:     ra.creator,
		})
		if ra.alias != "" {
			err := dk.SetAliasForRollAppId(ctx, ra.rollAppId, ra.alias)
			require.NoError(t, err)
			requireAliasLinkedToRollApp(ra.alias, ra.rollAppId, t, ctx, dk)
		} else {
			requireRollAppHasNoAlias(ra.rollAppId, t, ctx, dk)
		}
	}

	t.Run("do not process message that not pass basic validation", func(t *testing.T) {
		requireErrorFContains(t, func() error {
			resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
				GoodsId: rollapp_1_ofOwner.alias,
				Owner:   "0x1", // invalid owner
			})

			require.Nil(t, resp)

			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	t.Run("do not process message that refer to non-existing alias", func(t *testing.T) {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   "void",
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "alias is not in-used: void: not found")
	})

	t.Run("do not process for Alias that does not have any SO", func(t *testing.T) {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   rollapp_1_ofOwner.alias,
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), fmt.Sprintf("Sell-Order: %s: not found", rollapp_1_ofOwner.alias))
	})

	t.Run("do not process that owner does not match", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:   rollapp_1_ofOwner.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err := dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.AliasOrder,
			Owner:     anotherA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "not the owner of the RollApp")
	})

	t.Run("can not cancel expired order", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:   rollapp_1_ofOwner.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err := dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "cannot cancel an expired order")
	})

	t.Run("can not cancel once bid placed", func(t *testing.T) {
		so11 := dymnstypes.SellOrder{
			GoodsId:  rollapp_1_ofOwner.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  dymnsutils.TestCoin(300),
				Params: []string{rollapp_4_ofBidder.rollAppId},
			},
		}
		err := dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "cannot cancel once bid placed")
	})

	t.Run("cancellation will remove the active SO expiration mapping record", func(t *testing.T) {
		aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.AliasOrder)

		so11 := dymnstypes.SellOrder{
			GoodsId:  rollapp_1_ofOwner.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err := dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)
		aSoe.Add(so11.GoodsId, so11.ExpireAt)

		so12 := dymnstypes.SellOrder{
			GoodsId:  rollapp_2_ofOwner.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so12)
		require.NoError(t, err)
		aSoe.Add(so12.GoodsId, so12.ExpireAt)

		err = dk.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.AliasOrder)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder)
			dk.DeleteSellOrder(ctx, so12.GoodsId, dymnstypes.AliasOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Nil(t, dk.GetSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder), "SO should be removed from active")

		aSoe = dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.AliasOrder)

		allAliases := make(map[string]bool)
		for _, record := range aSoe.Records {
			allAliases[record.GoodsId] = true
		}
		require.NotContains(t, allAliases, so11.GoodsId)
		require.Contains(t, allAliases, so12.GoodsId)
	})

	t.Run("can cancel if satisfied conditions", func(t *testing.T) {
		moduleParams := dk.GetParams(ctx)
		moduleParams.Misc.EnableTradingAlias = false // allowed to cancel even if trading is disabled
		require.NoError(t, dk.SetParams(ctx, moduleParams))

		so11 := dymnstypes.SellOrder{
			GoodsId:  rollapp_1_ofOwner.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err := dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		so12 := dymnstypes.SellOrder{
			GoodsId:  rollapp_2_ofOwner.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so12)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder)
			dk.DeleteSellOrder(ctx, so12.GoodsId, dymnstypes.AliasOrder)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(ctx), &dymnstypes.MsgCancelSellOrder{
			GoodsId:   so11.GoodsId,
			OrderType: dymnstypes.AliasOrder,
			Owner:     ownerA,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Nil(t, dk.GetSellOrder(ctx, so11.GoodsId, dymnstypes.AliasOrder), "SO should be removed from active")
		require.NotNil(t, dk.GetSellOrder(ctx, so12.GoodsId, dymnstypes.AliasOrder), "other records remaining as-is")

		list := dk.GetHistoricalSellOrders(ctx, so11.GoodsId, dymnstypes.AliasOrder)
		require.Empty(t, list, "no historical record should be added")

		require.GreaterOrEqual(t,
			ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCloseSellOrder,
			"should consume params gas",
		)

		requireAliasLinkedToRollApp(rollapp_1_ofOwner.alias, rollapp_1_ofOwner.rollAppId, t, ctx, dk)
	})
}
