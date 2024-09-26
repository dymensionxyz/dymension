package keeper_test

import (
	"fmt"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_CancelSellOrder_DymName() {
	msgServer := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper)

	ownerA := testAddr(1).bech32()
	notOwnerA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}
	err := s.dymNsKeeper.SetDymName(s.ctx, dymName1)
	s.Require().NoError(err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}
	err = s.dymNsKeeper.SetDymName(s.ctx, dymName2)
	s.Require().NoError(err)

	dymNames := s.dymNsKeeper.GetAllNonExpiredDymNames(s.ctx)
	s.Require().Len(dymNames, 2)

	s.Run("do not process message that not pass basic validation", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId: "abc",
			Owner:   "0x1", // invalid owner
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())

		s.Require().Nil(resp)
	})

	s.Run("do not process message that refer to non-existing Dym-Name", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   "not-exists",
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "Dym-Name: not-exists: not found")
	})

	s.Run("do not process message that type is Unknown", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.AssetType_AT_UNKNOWN,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "invalid asset type")
	})

	s.Run("do not process that owner does not match", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeName,
			Owner:     notOwnerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "not the owner of the Dym-Name")
	})

	s.Run("do not process for Dym-Name that does not have any SO", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), fmt.Sprintf("Sell-Order: %s: not found", dymName1.Name))
	})

	s.Run("can not cancel expired", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "cannot cancel an expired order")
	})

	s.Run("can not cancel once bid placed", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  s.coin(300),
			},
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "cannot cancel once bid placed")
	})

	s.Run("can will remove the active SO expiration mapping record", func() {
		aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)

		so11 := dymnstypes.SellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)
		aSoe.Add(so11.AssetId, so11.ExpireAt)

		so12 := dymnstypes.SellOrder{
			AssetId:   dymName2.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so12)
		s.Require().NoError(err)
		aSoe.Add(so12.AssetId, so12.ExpireAt)

		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, aSoe, dymnstypes.TypeName)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName)
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so12.AssetId, dymnstypes.TypeName)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName), "SO should be removed from active")

		aSoe = s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)

		allNames := make(map[string]bool)
		for _, record := range aSoe.Records {
			allNames[record.AssetId] = true
		}
		s.Require().NotContains(allNames, so11.AssetId)
		s.Require().Contains(allNames, so12.AssetId)
	})

	s.Run("can cancel if satisfied conditions", func() {
		const previousRunGasConsumed = 100_000_000
		s.ctx.GasMeter().ConsumeGas(previousRunGasConsumed, "simulate previous run")

		moduleParams := s.dymNsKeeper.GetParams(s.ctx)
		moduleParams.Misc.EnableTradingName = false // allowed to cancel even if trading is disabled
		s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))

		so11 := dymnstypes.SellOrder{
			AssetId:   dymName1.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		so12 := dymnstypes.SellOrder{
			AssetId:   dymName2.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so12)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName)
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so12.AssetId, dymnstypes.TypeName)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeName,
			Owner:     ownerA,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeName), "SO should be removed from active")
		s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName2.Name, dymnstypes.TypeName), "other records remaining as-is")

		s.GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCloseSellOrder,
			"should consume params gas",
		)
		s.GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), previousRunGasConsumed+dymnstypes.OpGasCloseSellOrder,
			"gas consumption should be stacked with previous run",
		)
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_CancelSellOrder_Alias() {
	msgServer := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper)

	ownerA := testAddr(1).bech32()
	anotherA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	rollapp_1_ofOwner := *newRollApp("rollapp_1-1").WithOwner(ownerA).WithAlias("one")
	rollapp_2_ofOwner := *newRollApp("rollapp_2-1").WithOwner(ownerA).WithAlias("two")
	rollapp_3_ofAnother := *newRollApp("rollapp_3-1").WithOwner(anotherA).WithAlias("three")
	rollapp_4_ofBidder := *newRollApp("rollapp_4-2").WithOwner(bidderA)
	for _, ra := range []rollapp{rollapp_1_ofOwner, rollapp_2_ofOwner, rollapp_3_ofAnother, rollapp_4_ofBidder} {
		s.persistRollApp(ra)
	}

	s.Run("do not process message that not pass basic validation", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId: rollapp_1_ofOwner.alias,
			Owner:   "0x1", // invalid owner
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
		s.Require().Nil(resp)
	})

	s.Run("do not process message that refer to non-existing alias", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   "void",
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "alias is not in-used: void: not found")
	})

	s.Run("do not process for Alias that does not have any SO", func() {
		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), fmt.Sprintf("Sell-Order: %s: not found", rollapp_1_ofOwner.alias))
	})

	s.Run("do not process that owner does not match", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeAlias,
			Owner:     anotherA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "not the owner of the RollApp")
	})

	s.Run("can not cancel expired order", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "cannot cancel an expired order")
	})

	s.Run("can not cancel once bid placed", func() {
		so11 := dymnstypes.SellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  s.coin(300),
				Params: []string{rollapp_4_ofBidder.rollAppId},
			},
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Contains(err.Error(), "cannot cancel once bid placed")
	})

	s.Run("cancellation will remove the active SO expiration mapping record", func() {
		aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)

		so11 := dymnstypes.SellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)
		aSoe.Add(so11.AssetId, so11.ExpireAt)

		so12 := dymnstypes.SellOrder{
			AssetId:   rollapp_2_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so12)
		s.Require().NoError(err)
		aSoe.Add(so12.AssetId, so12.ExpireAt)

		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, aSoe, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias)
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so12.AssetId, dymnstypes.TypeAlias)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias), "SO should be removed from active")

		aSoe = s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)

		allAliases := make(map[string]bool)
		for _, record := range aSoe.Records {
			allAliases[record.AssetId] = true
		}
		s.Require().NotContains(allAliases, so11.AssetId)
		s.Require().Contains(allAliases, so12.AssetId)
	})

	s.Run("can cancel if satisfied conditions", func() {
		const previousRunGasConsumed = 100_000_000
		s.ctx.GasMeter().ConsumeGas(previousRunGasConsumed, "simulate previous run")

		moduleParams := s.dymNsKeeper.GetParams(s.ctx)
		moduleParams.Misc.EnableTradingAlias = false // allowed to cancel even if trading is disabled
		s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))

		so11 := dymnstypes.SellOrder{
			AssetId:   rollapp_1_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so11)
		s.Require().NoError(err)

		so12 := dymnstypes.SellOrder{
			AssetId:   rollapp_2_ofOwner.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so12)
		s.Require().NoError(err)

		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias)
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so12.AssetId, dymnstypes.TypeAlias)
		}()

		resp, err := msgServer.CancelSellOrder(sdk.WrapSDKContext(s.ctx), &dymnstypes.MsgCancelSellOrder{
			AssetId:   so11.AssetId,
			AssetType: dymnstypes.TypeAlias,
			Owner:     ownerA,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so11.AssetId, dymnstypes.TypeAlias), "SO should be removed from active")
		s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, so12.AssetId, dymnstypes.TypeAlias), "other records remaining as-is")

		s.GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCloseSellOrder,
			"should consume params gas",
		)
		s.GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), previousRunGasConsumed+dymnstypes.OpGasCloseSellOrder,
			"gas consumption should be stacked with previous run",
		)

		s.requireAlias(rollapp_1_ofOwner.alias).LinkedToRollApp(rollapp_1_ofOwner.rollAppId)
	})
}
