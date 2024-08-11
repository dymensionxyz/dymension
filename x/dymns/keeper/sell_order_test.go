package keeper_test

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_GetSetSellOrder() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("reject invalid SO", func() {
		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{})
		s.Require().Error(err)
	})

	s.Run("can set", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.FriendlyString(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(100),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			})
		}
	})

	s.Run("event should be fired on set", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.FriendlyString(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(100),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

				func(events sdk.Events) {
					for _, event := range events {
						if event.Type != dymnstypes.EventTypeSellOrder {
							continue
						}

						var actionName string
						for _, attr := range event.Attributes {
							if attr.Key == dymnstypes.AttributeKeySoActionName {
								actionName = attr.Value
							}
						}
						s.Require().NotEmpty(actionName, "event attr action name could not be found")
						s.Require().Equalf(
							actionName, dymnstypes.AttributeValueSoActionNameSet,
							"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameSet,
						)
						return
					}

					s.T().Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
				}(events)
			})
		}
	})

	s.Run("can not set Sell-Order with unknown type", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.AssetType_AT_UNKNOWN,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
		})
		s.Require().ErrorContains(err, "invalid SO type")
	})

	s.Run("non-exists returns nil", func() {
		s.RefreshContext()

		for _, assetType := range supportedAssetTypes {
			s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", assetType))
		}
	})

	s.Run("omit Sell Price if zero", func() {
		for _, sellPrice := range []*sdk.Coin{nil, uptr.To(s.coin(0)), {}} {
			for _, assetType := range supportedAssetTypes {
				s.Run(assetType.FriendlyString(), func() {
					s.RefreshContext()

					err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
						AssetId:   "asset",
						AssetType: assetType,
						ExpireAt:  1,
						MinPrice:  s.coin(100),
						SellPrice: sellPrice,
					})
					s.Require().NoError(err)

					s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", assetType).SellPrice)
				})
			}
		}
	})

	s.Run("omit Bid params if empty", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.TypeName,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: testAddr(1).bech32(),
				Price:  s.coin(200),
				Params: []string{},
			},
		})
		s.Require().NoError(err)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", dymnstypes.TypeName).HighestBid.Params)
	})

	s.Run("get returns correct inserted record, regardless type", func() {
		s.RefreshContext()

		var sellOrders []dymnstypes.SellOrder

		const seed int = 100

		for i := 0; i < seed; i++ {
			for _, assetType := range supportedAssetTypes {

				so := dymnstypes.SellOrder{
					AssetId:   fmt.Sprintf("dog%d", i), // same asset id for all types
					AssetType: assetType,
					ExpireAt:  1 + int64(i),
					MinPrice:  s.coin(int64(seed + i)),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				sellOrders = append(sellOrders, so)
			}
		}

		for _, so := range sellOrders {
			got := s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType)
			s.Require().NotNil(got)
			s.Require().Equal(so, *got)
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_DeleteSellOrder() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("can delete", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.FriendlyString(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))

				s.dymNsKeeper.DeleteSellOrder(s.ctx, so.AssetId, so.AssetType)

				s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))
			})
		}
	})

	s.Run("event should be fired upon deletion", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.FriendlyString(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.dymNsKeeper.DeleteSellOrder(s.ctx, so.AssetId, so.AssetType)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

				func(events sdk.Events) {
					for _, event := range events {
						if event.Type != dymnstypes.EventTypeSellOrder {
							continue
						}

						var actionName string
						for _, attr := range event.Attributes {
							if attr.Key == dymnstypes.AttributeKeySoActionName {
								actionName = attr.Value
							}
						}
						s.Require().NotEmpty(actionName, "event attr action name could not be found")
						s.Require().Equalf(
							actionName, dymnstypes.AttributeValueSoActionNameSet,
							"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameDelete,
						)
						return
					}

					s.T().Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
				}(events)
			})
		}
	})

	s.Run("delete remove the correct record regardless type", func() {
		s.RefreshContext()

		type testCase struct {
			so      dymnstypes.SellOrder
			deleted bool
		}

		var testCases []*testCase

		// build testcases
		const seed int = 100
		for i := 0; i < seed; i++ {
			for j, assetType := range supportedAssetTypes {
				so := dymnstypes.SellOrder{
					AssetId:   fmt.Sprintf("dog%03d%d", i, j),
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))

				testCases = append(testCases, &testCase{so: so, deleted: false})
			}
		}

		s.Require().Len(testCases, seed*len(supportedAssetTypes))
		s.Require().Len(s.dymNsKeeper.GetAllSellOrders(s.ctx), len(testCases))

		// test delete
		for i, tc := range testCases {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, tc.so.AssetId, tc.so.AssetType)
			tc.deleted = true
			s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, tc.so.AssetId, tc.so.AssetType))

			s.Require().Len(s.dymNsKeeper.GetAllSellOrders(s.ctx), len(testCases)-(i+1))

			for _, tc2 := range testCases {
				if tc2.deleted {
					s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, tc2.so.AssetId, tc2.so.AssetType))
				} else {
					s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, tc2.so.AssetId, tc2.so.AssetType))
				}
			}
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_MoveSellOrderToHistorical() {
	s.RefreshContext()

	ownerA := testAddr(1).bech32()
	bidderA := testAddr(2).bech32()

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

	so11 := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  s.coin(100),
		SellPrice: uptr.To(s.coin(300)),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
	s.Require().NoError(err)

	alias2 := "alias"
	so21 := dymnstypes.SellOrder{
		AssetId:   alias2,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  2,
		MinPrice:  s.coin(100),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so21)
	s.Require().NoError(err)

	alias3 := "salas"
	so3 := dymnstypes.SellOrder{
		AssetId:   alias3,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  3,
		MinPrice:  s.coin(222),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so3)
	s.Require().NoError(err)

	s.Run("should able to move", func() {
		err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so11.AssetId, so11.AssetType)
		s.Require().NoError(err)

		err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so21.AssetId, so21.AssetType)
		s.Require().NoError(err)
	})

	s.Run("moved SO should be removed from active", func() {
		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so11.AssetId, so11.AssetType))
		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so21.AssetId, so21.AssetType))
	})

	s.Run("has min expiry mapping", func() {
		minExpiry, found := s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, so11.AssetId, so11.AssetType)
		s.Require().True(found)
		s.Require().Equal(so11.ExpireAt, minExpiry)

		minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, so21.AssetId, so21.AssetType)
		s.Require().True(found)
		s.Require().Equal(so21.ExpireAt, minExpiry)
	})

	s.Run("should not move non-exists", func() {
		err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, "non-exists", dymnstypes.TypeName)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Sell-Order: Dym-Name: non-exists")

		err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, "non-exists", dymnstypes.TypeAlias)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Sell-Order: Alias: non-exists")
	})

	s.Run("should able to move a duplicated without error", func() {
		defer func() {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so11.AssetId, so11.AssetType)
			s.dymNsKeeper.DeleteSellOrder(s.ctx, so21.AssetId, so21.AssetType)
		}()

		for _, so := range []dymnstypes.SellOrder{so11, so21} {
			err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
			s.Require().NoError(err)

			err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so.AssetId, so.AssetType)
			s.Require().NoError(err)

			list := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, so.AssetId, so.AssetType)
			s.Require().Len(list, 1, "do not persist duplicated historical SO")
		}
	})

	s.Run("other records remaining as-is", func() {
		s.Require().Empty(s.dymNsKeeper.GetSellOrder(s.ctx, dymName2.Name, dymnstypes.TypeName))
		s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, alias3, dymnstypes.TypeAlias))
	})

	so4 := dymnstypes.SellOrder{
		AssetId:   dymName2.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  s.coin(100),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so4)
	s.Require().NoError(err)

	s.Run("should able to move", func() {
		err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so4.AssetId, so4.AssetType)
		s.Require().NoError(err)
	})

	s.Run("other records remaining as-is", func() {
		s.Require().Len(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName1.Name, dymnstypes.TypeName), 1)
		s.Require().Len(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName), 1)
		s.Require().Len(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias2, dymnstypes.TypeAlias), 1)
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias))
	})

	so12 := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  s.now.Unix() + 100,
		MinPrice:  s.coin(100),
		SellPrice: uptr.To(s.coin(300)),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so12)
	s.Require().NoError(err)

	so22 := dymnstypes.SellOrder{
		AssetId:   alias2,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  s.now.Unix() + 100,
		MinPrice:  s.coin(100),
		SellPrice: uptr.To(s.coin(300)),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so22)
	s.Require().NoError(err)

	s.Run("should not move yet finished SO", func() {
		err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so12.AssetId, so12.AssetType)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Sell-Order not yet expired")

		err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so22.AssetId, so22.AssetType)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Sell-Order not yet expired")
	})

	for _, so := range []dymnstypes.SellOrder{so12, so22} {
		so.HighestBid = &dymnstypes.SellOrderBid{
			Bidder: bidderA,
			Price:  s.coin(300),
		}
		if so.AssetType == dymnstypes.TypeAlias {
			so.HighestBid.Params = []string{"rollapp_1-1"}
		}

		err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Run("should able to move finished SO", func() {
			err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so.AssetId, so.AssetType)
			s.Require().NoError(err)

			list := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, so.AssetId, so.AssetType)
			s.Require().Len(list, 2, "should appended to historical")
		})
	}

	minExpiry, found := s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, so12.AssetId, so12.AssetType)
	s.Require().True(found)
	s.Require().Equal(so11.ExpireAt, minExpiry, "should keep the minimum")
	s.Require().NotEqual(so12.ExpireAt, minExpiry, "should keep the minimum")

	minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, so22.AssetId, so22.AssetType)
	s.Require().True(found)
	s.Require().Equal(so21.ExpireAt, minExpiry, "should keep the minimum")
	s.Require().NotEqual(so22.ExpireAt, minExpiry, "should keep the minimum")

	s.Run("other records remaining as-is", func() {
		s.Require().Len(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName), 1)
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias))
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAndDeleteHistoricalSellOrders() {
	s.RefreshContext()

	ownerA := testAddr(1).bech32()

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

	alias3 := "alias"
	alias4 := "salas"

	s.Run("getting non-exists should returns empty", func() {
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName1.Name, dymnstypes.TypeName))
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName))
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias))
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias4, dymnstypes.TypeAlias))
	})

	so11 := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  s.coin(100),
		SellPrice: uptr.To(s.coin(300)),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so11)
	s.Require().NoError(err)
	err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so11.AssetId, so11.AssetType)
	s.Require().NoError(err)

	so2 := dymnstypes.SellOrder{
		AssetId:   alias3,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  7,
		MinPrice:  s.coin(200),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so2)
	s.Require().NoError(err)
	err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so2.AssetId, so2.AssetType)
	s.Require().NoError(err)

	so3 := dymnstypes.SellOrder{
		AssetId:   dymName2.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  s.coin(100),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so3)
	s.Require().NoError(err)
	err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so3.AssetId, so3.AssetType)
	s.Require().NoError(err)

	so3.ExpireAt++
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so3)
	s.Require().NoError(err)
	err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so3.AssetId, so3.AssetType)
	s.Require().NoError(err)

	so4 := dymnstypes.SellOrder{
		AssetId:   alias4,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  5,
		MinPrice:  s.coin(500),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so4)
	s.Require().NoError(err)
	err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, so4.AssetId, so4.AssetType)
	s.Require().NoError(err)

	s.Run("fetch correctly", func() {
		list1 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName1.Name, dymnstypes.TypeName)
		s.Require().Len(list1, 1)
		list2 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName)
		s.Require().Len(list2, 2)
		list3 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias)
		s.Require().Len(list3, 1)
		list4 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias4, dymnstypes.TypeAlias)
		s.Require().Len(list4, 1)

		s.Require().Equal(so3.AssetId, list2[0].AssetId)
		s.Require().Equal(so3.AssetId, list2[1].AssetId)

		s.Require().Equal(int64(1), list2[0].ExpireAt)
		s.Require().Equal(int64(2), list2[1].ExpireAt)

		s.Require().Equal(int64(7), list3[0].ExpireAt)
		s.Require().Equal(int64(5), list4[0].ExpireAt)
	})

	s.Run("delete", func() {
		s.dymNsKeeper.DeleteHistoricalSellOrders(s.ctx, dymName1.Name, dymnstypes.TypeName)
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName1.Name, dymnstypes.TypeName))

		list2 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName)
		s.Require().Len(list2, 2)

		s.dymNsKeeper.DeleteHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName)
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, dymName2.Name, dymnstypes.TypeName))

		list3 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias)
		s.Require().Len(list3, 1)

		s.dymNsKeeper.DeleteHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias)
		s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias3, dymnstypes.TypeAlias))

		list4 := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias4, dymnstypes.TypeAlias)
		s.Require().Len(list4, 1)
	})
}

func (s *KeeperTestSuite) TestKeeper_GetSetActiveSellOrdersExpiration() {
	s.RefreshContext()

	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("get", func() {
		for _, assetType := range supportedAssetTypes {
			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Empty(aSoe.Records, "default list must be empty")
			s.Require().NotNil(aSoe.Records, "list must be initialized")
		}
	})

	s.Run("set", func() {
		for _, assetType := range supportedAssetTypes {
			aSoe := &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "hello",
						ExpireAt: 123,
					},
					{
						AssetId:  "world",
						ExpireAt: 456,
					},
				},
			}
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, aSoe, assetType)
			s.Require().NoError(err)

			aSoe = s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Len(aSoe.Records, 2)
			s.Require().Equal("hello", aSoe.Records[0].AssetId)
			s.Require().Equal(int64(123), aSoe.Records[0].ExpireAt)
			s.Require().Equal("world", aSoe.Records[1].AssetId)
			s.Require().Equal(int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	s.Run("must automatically sort when set", func() {
		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "bbb",
						ExpireAt: 456,
					},
					{
						AssetId:  "aaa",
						ExpireAt: 123,
					},
				},
			}, assetType)
			s.Require().NoError(err)

			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Len(aSoe.Records, 2)

			s.Require().Equal("aaa", aSoe.Records[0].AssetId)
			s.Require().Equal(int64(123), aSoe.Records[0].ExpireAt)
			s.Require().Equal("bbb", aSoe.Records[1].AssetId)
			s.Require().Equal(int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	s.Run("can not set if set is not valid", func() {
		for _, assetType := range supportedAssetTypes {
			// not unique
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "dup",
						ExpireAt: 456,
					},
					{
						AssetId:  "dup",
						ExpireAt: 123,
					},
				},
			}, assetType)
			s.Require().Error(err)

			// zero expiry
			err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "alice",
						ExpireAt: -1,
					},
					{
						AssetId:  "bob",
						ExpireAt: 0,
					},
				},
			}, assetType)
			s.Require().Error(err)
		}
	})

	s.Run("each asset type persists separately", func() {
		err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 1,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 2,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		listName := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)
		s.Require().Len(listName.Records, 1)
		s.Require().Equal("asset", listName.Records[0].AssetId)
		s.Require().Equal(int64(1), listName.Records[0].ExpireAt)

		listAlias := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)
		s.Require().Len(listAlias.Records, 1)
		s.Require().Equal("asset", listAlias.Records[0].AssetId)
		s.Require().Equal(int64(2), listAlias.Records[0].ExpireAt)
	})
}

func (s *KeeperTestSuite) TestKeeper_GetSetMinExpiryHistoricalSellOrder() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	for _, assetType := range supportedAssetTypes {
		s.Run(assetType.FriendlyString(), func() {
			s.RefreshContext()

			s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "hello", assetType, 123)
			s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "world", assetType, 456)

			minExpiry, found := s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "hello", assetType)
			s.Require().True(found)
			s.Require().Equal(int64(123), minExpiry)

			minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "world", assetType)
			s.Require().True(found)
			s.Require().Equal(int64(456), minExpiry)

			minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "non-exists", assetType)
			s.Require().False(found)
			s.Require().Zero(minExpiry)

			s.Run("set zero means delete", func() {
				s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "hello", assetType, 0)

				minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "hello", assetType)
				s.Require().False(found)
				s.Require().Zero(minExpiry)
			})
		})
	}

	s.Run("each asset type persists separately", func() {
		s.RefreshContext()

		s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "asset", dymnstypes.TypeName, 1)
		s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "asset", dymnstypes.TypeAlias, 2)
		s.dymNsKeeper.SetMinExpiryHistoricalSellOrder(s.ctx, "pool", dymnstypes.TypeAlias, 2)

		minExpiry, found := s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "asset", dymnstypes.TypeName)
		s.Require().True(found)
		s.Require().Equal(int64(1), minExpiry)

		minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "asset", dymnstypes.TypeAlias)
		s.Require().True(found)
		s.Require().Equal(int64(2), minExpiry)

		minExpiry, found = s.dymNsKeeper.GetMinExpiryHistoricalSellOrder(s.ctx, "pool", dymnstypes.TypeName)
		s.Require().False(found)
		s.Require().Zero(minExpiry)
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAllSellOrders() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.RefreshContext()

	var sellOrders []dymnstypes.SellOrder

	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000))
	s.Require().NoError(err)

	seed := 200 + int(n.Int64())

	for i := 0; i < seed; i++ {
		for j, assetType := range supportedAssetTypes {
			so := dymnstypes.SellOrder{
				AssetId:   fmt.Sprintf("dog%03d%d", i, j),
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  s.coin(1),
			}
			err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
			s.Require().NoError(err)

			sellOrders = append(sellOrders, so)
		}
	}

	s.Require().Len(sellOrders, seed*len(supportedAssetTypes))

	allSellOrders := s.dymNsKeeper.GetAllSellOrders(s.ctx)

	s.Require().Len(allSellOrders, len(sellOrders), "should returns all inserted records")

	for _, so := range sellOrders {
		s.Require().Contains(allSellOrders, so)
	}
}
