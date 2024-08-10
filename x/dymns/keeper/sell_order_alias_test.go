package keeper_test

import (
	"fmt"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) TestKeeper_CompleteAliasSellOrder() {
	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()

	rollApp_1_asSrc := *newRollApp("rollapp_1-1").WithOwner(creator_1_asOwner).WithAlias("alias")
	rollApp_2_asSrc := *newRollApp("rollapp_2-1").WithOwner(creator_1_asOwner)
	rollApp_3_asDst := *newRollApp("rollapp_3-2").WithOwner(creator_2_asBuyer)
	rollApp_4_asDst_byOwner := *newRollApp("rollapp_4-1").WithOwner(creator_1_asOwner).WithAlias("exists")

	s.Run("alias not found", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_2_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		so := dymnstypes.SellOrder{
			AssetId:   "void",
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(200),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_3_asDst.owner,
				Price:  dymnsutils.TestCoin(200),
				Params: []string{rollApp_3_asDst.rollAppId},
			},
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId), "alias not owned by any RollApp: void: not found")
	})

	s.Run("destination Roll-App not found", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)

		const sellPrice = 200

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(sellPrice),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_3_asDst.owner,
				Price:  dymnsutils.TestCoin(sellPrice),
				Params: []string{"nah_0-0"},
			},
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId), "destination Roll-App does not exists")
	})

	s.Run("SO not found", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		s.Require().ErrorContains(
			s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, rollApp_1_asSrc.alias),
			fmt.Sprintf("Sell-Order: %s: not found", rollApp_1_asSrc.alias),
		)
	})

	s.Run("SO not yet completed, no bidder", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(
			s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId),
			"Sell-Order has not finished yet",
		)
	})

	s.Run("SO has bidder but not yet completed", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_3_asDst.owner,
				Price:  dymnsutils.TestCoin(200), // lower than sell price
				Params: []string{rollApp_3_asDst.rollAppId},
			},
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(
			s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId),
			"Sell-Order has not finished yet",
		)
	})

	s.Run("SO expired without bidder", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() - 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId), "no bid placed")
	})

	s.Run("SO without sell price, with bid, finished by expiry", func() {
		s.SetupTest()

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_3_asDst)

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_3_asDst.owner,
				Price:  dymnsutils.TestCoin(200),
				Params: []string{rollApp_3_asDst.rollAppId},
			},
		}
		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId), "Sell-Order has not finished yet")
	})

	const ownerOriginalBalance int64 = 1000
	const buyerOriginalBalance int64 = 500
	tests := []struct {
		name                  string
		expiredSO             bool
		sellPrice             int64
		bid                   int64
		wantErr               bool
		wantErrContains       string
		wantOwnerBalanceLater int64
	}{
		{
			name:                  "pass - completed, expired, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "pass - completed, expired, under sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "pass - completed, expired, equals sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "pass - completed by sell-price met, not expired",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "fail - expired without bid, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - expired without bid, with sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - not expired but bid under sell price",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - not expired has bid, no sell price",
			expiredSO:             false,
			sellPrice:             0,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// setup execution context
			s.SetupTest()

			s.mintToAccount(creator_1_asOwner, ownerOriginalBalance)
			s.mintToAccount(creator_2_asBuyer, buyerOriginalBalance)

			s.persistRollApp(rollApp_1_asSrc)
			s.persistRollApp(rollApp_3_asDst)

			so := dymnstypes.SellOrder{
				AssetId:   rollApp_1_asSrc.alias,
				AssetType: dymnstypes.TypeAlias,
				MinPrice:  dymnsutils.TestCoin(100),
			}

			if tt.expiredSO {
				so.ExpireAt = s.now.Unix() - 1
			} else {
				so.ExpireAt = s.now.Unix() + 1
			}

			s.Require().GreaterOrEqual(tt.sellPrice, int64(0), "bad setup")
			so.SellPrice = dymnsutils.TestCoinP(tt.sellPrice)

			s.Require().GreaterOrEqual(tt.bid, int64(0), "bad setup")
			if tt.bid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: rollApp_3_asDst.owner,
					Price:  dymnsutils.TestCoin(tt.bid),
					Params: []string{rollApp_3_asDst.rollAppId},
				}

				// mint coin to module account because we charged buyer before update SO
				s.mintToModuleAccount2(so.HighestBid.Price.Amount)
			}
			err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
			s.Require().NoError(err)

			rollApp1, found := s.rollAppKeeper.GetRollapp(s.ctx, rollApp_1_asSrc.rollAppId)
			s.Require().True(found)
			rollApp2, found := s.rollAppKeeper.GetRollapp(s.ctx, rollApp_3_asDst.rollAppId)
			s.Require().True(found)

			// test

			errCompleteSellOrder := s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId)

			laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, dymnstypes.TypeAlias)

			historicalSo := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, so.AssetId, dymnstypes.TypeAlias)
			s.Require().Empty(historicalSo, "historical should be empty as not supported for asset type Alias")

			laterOwnerBalance := s.balance(creator_1_asOwner)
			laterBuyerBalance := s.balance(creator_2_asBuyer)

			laterAliasOfRollApp1, _ := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp_1_asSrc.rollAppId)
			laterAliasOfRollApp2, _ := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp_3_asDst.rollAppId)

			laterAliasLinkedToRollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, rollApp_1_asSrc.alias)
			s.Require().True(found)

			laterRollApp1, found := s.rollAppKeeper.GetRollapp(s.ctx, rollApp_1_asSrc.rollAppId)
			s.Require().True(found, "rollapp should be kept")
			s.Require().Equal(rollApp1, laterRollApp1, "rollapp should not be changed")
			laterRollApp2, found := s.rollAppKeeper.GetRollapp(s.ctx, rollApp_3_asDst.rollAppId)
			s.Require().True(found, "rollapp should be kept")
			s.Require().Equal(rollApp2, laterRollApp2, "rollapp should not be changed")

			if tt.wantErr {
				s.Require().Error(errCompleteSellOrder, "action should be failed")
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Contains(errCompleteSellOrder.Error(), tt.wantErrContains)

				s.Require().NotNil(laterSo, "SO should not be deleted")

				s.Require().Equal(rollApp_1_asSrc.alias, laterAliasOfRollApp1, "alias should be kept")
				s.Require().Equal(rollApp_1_asSrc.rollAppId, laterAliasLinkedToRollAppId, "alias should be kept")
				s.Require().Empty(laterAliasOfRollApp2, "should not be linked to RollApp 2")

				s.Require().Equal(ownerOriginalBalance, laterOwnerBalance, "owner balance should not be changed")
				s.Require().Equal(tt.wantOwnerBalanceLater, laterOwnerBalance, "owner balance mis-match")
				s.Require().Equal(buyerOriginalBalance, laterBuyerBalance, "buyer balance should not be changed")
				return
			}

			s.Require().NoError(errCompleteSellOrder, "action should be successful")

			s.Require().Nil(laterSo, "SO should be deleted")
			s.Require().Empty(historicalSo, "historical should be empty as not supported for asset type Alias")

			s.Require().Empty(laterAliasOfRollApp1, "should not be linked to RollApp 1 anymore")
			s.Require().Equal(rollApp_3_asDst.rollAppId, laterAliasLinkedToRollAppId, "alias should be linked to RollApp 2")

			s.Require().Equal(tt.wantOwnerBalanceLater, laterOwnerBalance, "owner balance mis-match")
			s.Require().Equal(buyerOriginalBalance, laterBuyerBalance, "buyer balance should not be changed")
		})
	}

	s.Run("if buyer is owner, can still process", func() {
		s.SetupTest()

		const ownerOriginalBalance = 100
		const moduleAccountOriginalBalance = 1000
		const offerValue = 300

		s.mintToModuleAccount(moduleAccountOriginalBalance)
		s.mintToAccount(creator_1_asOwner, ownerOriginalBalance)

		s.persistRollApp(rollApp_1_asSrc)
		s.persistRollApp(rollApp_4_asDst_byOwner)

		so := dymnstypes.SellOrder{
			AssetId:   rollApp_1_asSrc.alias,
			AssetType: dymnstypes.TypeAlias,
			ExpireAt:  s.now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(offerValue),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_4_asDst_byOwner.owner,
				Price:  dymnsutils.TestCoin(offerValue),
				Params: []string{rollApp_4_asDst_byOwner.rollAppId},
			},
		}

		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		err = s.dymNsKeeper.CompleteAliasSellOrder(s.ctx, so.AssetId)
		s.Require().NoError(err)

		// Alias should be transferred as normal
		laterAliasOfRollApp1, _ := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp_1_asSrc.rollAppId)
		laterAliasOfRollApp3, _ := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, rollApp_4_asDst_byOwner.rollAppId)
		laterAliasLinkedToRollAppId, _ := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, so.AssetId)

		s.Require().Empty(laterAliasOfRollApp1, "should not be linked to RollApp 1 anymore")
		s.Require().Equal(rollApp_4_asDst_byOwner.alias, laterAliasOfRollApp3, "alias should be linked to RollApp 3")
		s.Require().Equal(rollApp_4_asDst_byOwner.rollAppId, laterAliasLinkedToRollAppId, "alias should be linked to RollApp 3")

		// ensure all existing alias are linked to the correct RollApp
		for _, alias := range []string{rollApp_4_asDst_byOwner.alias, so.AssetId} {
			s.requireAlias(alias).LinkedToRollApp(rollApp_4_asDst_byOwner.rollAppId)
		}

		// owner receives the offer amount because owner also the buyer
		laterOwnerBalance := s.balance(creator_1_asOwner)
		s.Require().Equal(int64(offerValue+ownerOriginalBalance), laterOwnerBalance)

		// SO records should be processed as normal
		laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, dymnstypes.TypeAlias)
		s.Require().Nil(laterSo, "SO should be deleted")

		historicalSo := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, so.AssetId, dymnstypes.TypeAlias)
		s.Require().Empty(historicalSo, "historical should be empty as not supported for asset type Alias")
	})
}
