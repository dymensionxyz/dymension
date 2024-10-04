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
			s.Run(assetType.PrettyName(), func() {
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
			s.Run(assetType.PrettyName(), func() {
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
				s.Run(assetType.PrettyName(), func() {
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
			s.Run(assetType.PrettyName(), func() {
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
			s.Run(assetType.PrettyName(), func() {
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
