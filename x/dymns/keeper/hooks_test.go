package keeper_test

import (
	"time"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"

	sdkmath "cosmossdk.io/math"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_epochHooks_BeforeEpochStart() {
	s.Run("should do nothing", func() {
		originalGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetEpochHooks().BeforeEpochStart(
			s.ctx, "hour", 1,
		)
		s.Require().NoError(err)

		s.Require().Equal(originalGas, s.ctx.GasMeter().GasConsumed())
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_epochHooks_AfterEpochEnd() {
	s.Run("should do something even nothing to do", func() {
		s.RefreshContext()

		moduleParams := s.moduleParams()

		originalGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(
			s.ctx,
			moduleParams.Misc.EndEpochHookIdentifier, 1,
		)
		s.Require().NoError(err)

		// gas should be changed because it should at least reading the params to check epoch identifier
		s.Require().Less(originalGas, s.ctx.GasMeter().GasConsumed(), "should do something")
	})

	s.Run("process active mixed Dym-Name and alias Sell-Orders", func() {
		s.RefreshContext()

		dymNameOwner := testAddr(1).bech32()
		dymNameBuyer := testAddr(2).bech32()

		creator1_asOwner := testAddr(3).bech32()
		creator2_asBuyer := testAddr(4).bech32()

		dymName1 := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      dymNameOwner,
			Controller: dymNameOwner,
			ExpireAt:   s.now.Add(2 * 365 * 24 * time.Hour).Unix(),
		}
		err := s.dymNsKeeper.SetDymName(s.ctx, dymName1)
		s.Require().NoError(err)

		rollApp1_asSrc := *newRollApp("rollapp_1-1").WithOwner(creator1_asOwner).WithAlias("one")
		s.persistRollApp(rollApp1_asSrc)
		s.requireRollApp(rollApp1_asSrc.rollAppId).HasAlias("one")
		rollApp2_asDst := *newRollApp("rollapp_2-2").WithOwner(creator2_asBuyer)
		s.persistRollApp(rollApp2_asDst)
		s.requireRollApp(rollApp2_asDst.rollAppId).HasNoAlias()

		const dymNameOrderPrice = 100
		const aliasOrderPrice = 200

		s.mintToModuleAccount(dymNameOrderPrice + aliasOrderPrice + 1)

		dymNameSO := s.newDymNameSellOrder(dymName1.Name).
			WithMinPrice(dymNameOrderPrice).
			WithDymNameBid(dymNameBuyer, dymNameOrderPrice).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, dymNameSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameSO.AssetId,
					ExpireAt: dymNameSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		aliasSO := s.newAliasSellOrder(rollApp1_asSrc.alias).
			WithMinPrice(aliasOrderPrice).
			WithAliasBid(rollApp2_asDst.owner, aliasOrderPrice, rollApp2_asDst.rollAppId).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, aliasSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasSO.AssetId,
					ExpireAt: aliasSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))
		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))

		s.Equal(int64(1), s.moduleBalance())
		s.Equal(int64(dymNameOrderPrice), s.balance(dymNameOwner))
		s.Equal(int64(aliasOrderPrice), s.balance(rollApp1_asSrc.owner))

		laterDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name)
		if s.NotNil(laterDymName) {
			s.Equal(dymNameBuyer, laterDymName.Owner)
			s.Equal(dymNameBuyer, laterDymName.Controller)
		}

		s.requireRollApp(rollApp1_asSrc.rollAppId).HasNoAlias()
		s.requireRollApp(rollApp2_asDst.rollAppId).HasAlias("one")
	})

	s.Run("should not process Dym-Name SO if trading is disabled", func() {
		s.RefreshContext()

		dymNameOwner := testAddr(1).bech32()
		dymNameBuyer := testAddr(2).bech32()

		dymName1 := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      dymNameOwner,
			Controller: dymNameOwner,
			ExpireAt:   s.now.Add(2 * 365 * 24 * time.Hour).Unix(),
		}
		err := s.dymNsKeeper.SetDymName(s.ctx, dymName1)
		s.Require().NoError(err)

		const dymNameOrderPrice = 100

		s.mintToModuleAccount(dymNameOrderPrice + 1)

		dymNameSO := s.newDymNameSellOrder(dymName1.Name).
			WithMinPrice(dymNameOrderPrice).
			WithDymNameBid(dymNameBuyer, dymNameOrderPrice).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, dymNameSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameSO.AssetId,
					ExpireAt: dymNameSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingName = false
			return p
		})

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		// the SellOrder should still be there
		s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))

		// re-enable and test again to make sure it not processes just because trading was disabled
		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingName = true
			return p
		})

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))
	})

	s.Run("should not process Alias SO if trading is disabled", func() {
		s.RefreshContext()

		creator1_asOwner := testAddr(3).bech32()
		creator2_asBuyer := testAddr(4).bech32()

		rollApp1_asSrc := *newRollApp("rollapp_1-1").WithOwner(creator1_asOwner).WithAlias("one")
		s.persistRollApp(rollApp1_asSrc)
		s.requireRollApp(rollApp1_asSrc.rollAppId).HasAlias("one")
		rollApp2_asDst := *newRollApp("rollapp_2-2").WithOwner(creator2_asBuyer)
		s.persistRollApp(rollApp2_asDst)
		s.requireRollApp(rollApp2_asDst.rollAppId).HasNoAlias()

		const aliasOrderPrice = 200

		s.mintToModuleAccount(aliasOrderPrice + 1)

		aliasSO := s.newAliasSellOrder(rollApp1_asSrc.alias).
			WithMinPrice(aliasOrderPrice).
			WithAliasBid(rollApp2_asDst.owner, aliasOrderPrice, rollApp2_asDst.rollAppId).
			Expired().Build()
		err := s.dymNsKeeper.SetSellOrder(s.ctx, aliasSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasSO.AssetId,
					ExpireAt: aliasSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingAlias = false
			return p
		})

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		// the SellOrder should still be there
		s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))

		// re-enable and test again to make sure it not processes just because trading was disabled
		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingAlias = true
			return p
		})

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))
	})
}

func (s *KeeperTestSuite) Test_epochHooks_AfterEpochEnd_processActiveDymNameSellOrders() {
	ownerAcc := testAddr(1)
	ownerA := ownerAcc.bech32()

	bidderAcc := testAddr(2)
	bidderA := bidderAcc.bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	originalDymNameA := dymNameA
	originalDymNameB := dymNameB
	originalDymNameC := dymNameC
	originalDymNameD := dymNameD

	coin100 := s.coin(100)
	coin200 := s.coin(200)

	soExpiredEpoch := s.now.Unix() - 1
	soNotExpiredEpoch := s.now.Add(time.Hour).Unix()

	const soExpired = true
	const soNotExpired = false
	genSo := func(
		dymName dymnstypes.DymName,
		expired bool, sellPrice *sdk.Coin, highestBid *dymnstypes.SellOrderBid,
	) dymnstypes.SellOrder {
		return dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt: func() int64 {
				if expired {
					return soExpiredEpoch
				}
				return soNotExpiredEpoch
			}(),
			MinPrice:   coin100,
			SellPrice:  sellPrice,
			HighestBid: highestBid,
		}
	}

	tests := []struct {
		name                  string
		dymNames              []dymnstypes.DymName
		sellOrders            []dymnstypes.SellOrder
		expiryByDymName       []dymnstypes.ActiveSellOrdersExpirationRecord
		preMintModuleBalance  int64
		customEpochIdentifier string
		beforeHookTestFunc    func(*KeeperTestSuite)
		wantErr               bool
		wantErrContains       string
		wantExpiryByDymName   []dymnstypes.ActiveSellOrdersExpirationRecord
		afterHookTestFunc     func(*KeeperTestSuite)
	}{
		{
			name:       "pass - simple process expired SO",
			dymNames:   []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, nil)},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymNameA.Name).
					noActiveSO().
					mustEquals(originalDymNameA)

				s.Require().EqualValues(200, s.moduleBalance())

				s.EqualValues(0, s.balance(dymNameA.Owner))

				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name:     "pass - simple process expired & completed SO",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  coin200,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymNameA.Name).
					noActiveSO().
					ownerChangedTo(bidderA).
					expiryEquals(originalDymNameA.ExpireAt)

				s.Require().EqualValues(0, s.moduleBalance()) // 200 should be transferred to previous owner

				s.Require().EqualValues(200, s.balance(dymNameA.Owner)) // previous owner should earn from bid

				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(bidderA).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(bidderAcc.fallback()).mappedDymNames(dymNameA.Name)
			},
		},
		{
			name:     "pass - simple process expired & completed SO, match by min price",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  coin100,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 250,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymNameA.Name).
					noActiveSO().
					ownerChangedTo(bidderA).
					expiryEquals(originalDymNameA.ExpireAt)

				s.Require().EqualValues(150, s.moduleBalance()) // 100 should be transferred to previous owner

				s.Require().EqualValues(100, s.balance(dymNameA.Owner)) // previous owner should earn from bid

				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(bidderA).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(bidderAcc.fallback()).mappedDymNames(dymNameA.Name)
			},
		},
		{
			name:     "pass - process multiple - mixed SOs",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidderA,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidderA,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 450,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// SO for Dym-Name A is expired without any bid/winner
				s.requireDymName(dymNameA.Name).
					noActiveSO().
					mustEquals(originalDymNameA)

				// SO for Dym-Name B not yet finished
				soB := s.dymNsKeeper.GetSellOrder(s.ctx, dymNameB.Name, dymnstypes.TypeName)
				s.Require().NotNil(soB)
				s.requireDymName(dymNameB.Name).
					mustEquals(originalDymNameB)

				// SO for Dym-Name C is completed with winner
				s.requireDymName(dymNameC.Name).
					noActiveSO().
					ownerChangedTo(bidderA).
					expiryEquals(originalDymNameC.ExpireAt)

				// SO for Dym-Name D is completed with winner
				s.requireDymName(dymNameD.Name).
					noActiveSO().
					ownerChangedTo(bidderA).
					expiryEquals(originalDymNameD.ExpireAt)

				s.Require().EqualValues(150, s.moduleBalance())

				s.Require().EqualValues(300, s.balance(ownerA)) // 200 from SO C, 100 from SO D

				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireConfiguredAddress(bidderA).mappedDymNames(dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).mappedDymNames(dymNameC.Name, dymNameD.Name)
			},
		},
		{
			name:     "pass - should do nothing if invalid epoch identifier",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidderA,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidderA,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance:  450,
			customEpochIdentifier: "another",
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymNameA.Name).mustEquals(originalDymNameA)
				s.requireDymName(dymNameB.Name).mustEquals(originalDymNameB)
				s.requireDymName(dymNameC.Name).mustEquals(originalDymNameC)
				s.requireDymName(dymNameD.Name).mustEquals(originalDymNameD)

				s.Require().EqualValues(450, s.moduleBalance())

				s.Require().EqualValues(0, s.balance(ownerA))

				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name:     "pass - should remove expiry reference to non-exists SO",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				// no SO for Dym-Name B
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// no SO for Dym-Name B but still have reference
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr:             false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to Dym-Name A because of processed
				// removed reference to Dym-Name B because SO not exists
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name:     "pass - update expiry if in-correct",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, nil, nil), // SO not expired
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// incorrect, SO not expired
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name, dymNameB.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 1, // not enough balance
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged

				s.requireConfiguredAddress(ownerA).mappedDymNames(dymNameA.Name)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
				s.requireFallbackAddress(ownerAcc.fallback()).mappedDymNames(dymNameA.Name)
				s.requireFallbackAddress(bidderAcc.fallback()).notMappedToAnyDymName()

				s.requireDymName(dymNameA.Name).mustHaveActiveSO()

				s.EqualValues(1, s.moduleBalance())
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs, one pass one fail",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin100,
				}),
				genSo(dymNameB, soExpired, nil, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 101, // just enough process the first SO
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(
					dymNameA.Name, dymNameB.Name,
				)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(
					dymNameB.Name,
				)
				s.requireConfiguredAddress(bidderA).mappedDymNames(
					dymNameA.Name,
				)

				s.requireDymName(dymNameA.Name).noActiveSO()
				s.requireDymName(dymNameB.Name).mustHaveActiveSO()

				s.EqualValues(1, s.moduleBalance())
				s.EqualValues(100, s.balance(ownerA))
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs, one fail one pass",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				}),
				genSo(dymNameB, soExpired, nil, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 101, // just enough process the second SO
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(
					dymNameA.Name, dymNameB.Name,
				)
				s.requireConfiguredAddress(bidderA).notMappedToAnyDymName()
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).mappedDymNames(
					dymNameA.Name,
				)
				s.requireConfiguredAddress(bidderA).mappedDymNames(
					dymNameB.Name,
				)

				s.requireDymName(dymNameA.Name).mustHaveActiveSO()
				s.requireDymName(dymNameB.Name).noActiveSO()

				s.EqualValues(1, s.moduleBalance())
				s.EqualValues(100, s.balance(ownerA))
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NotNil(tt.beforeHookTestFunc, "mis-configured test case")
			s.Require().NotNil(tt.afterHookTestFunc, "mis-configured test case")

			s.RefreshContext()

			if tt.preMintModuleBalance > 0 {
				s.mintToModuleAccount(tt.preMintModuleBalance)
			}

			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.expiryByDymName,
			}, dymnstypes.TypeName)
			s.Require().NoError(err)

			for _, dymName := range tt.dymNames {
				s.setDymNameWithFunctionsAfter(dymName)
			}

			for _, so := range tt.sellOrders {
				err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			}

			moduleParams := s.dymNsKeeper.GetParams(s.ctx)

			useEpochIdentifier := moduleParams.Misc.EndEpochHookIdentifier
			if tt.customEpochIdentifier != "" {
				useEpochIdentifier = tt.customEpochIdentifier
			}

			tt.beforeHookTestFunc(s)

			err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, useEpochIdentifier, 1)

			defer func() {
				if s.T().Failed() {
					return
				}

				tt.afterHookTestFunc(s)

				aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)
				if len(tt.wantExpiryByDymName) == 0 {
					s.Require().Empty(aSoe.Records)
				} else {
					s.Require().Equal(tt.wantExpiryByDymName, aSoe.Records)
				}
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				return
			}

			s.Require().NoError(err)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_epochHooks_AfterEpochEnd_processActiveAliasSellOrders() {
	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBidder := testAddr(2).bech32()

	rollApp_1_byOwner_asSrc := *newRollApp("rollapp_1-1").WithAlias("one").WithOwner(creator_1_asOwner)
	rollApp_2_byBuyer_asDst := *newRollApp("rollapp_2-2").WithOwner(creator_2_asBidder)
	rollApp_3_byOwner_asSrc := *newRollApp("rollapp_3-1").WithAlias("three").WithOwner(creator_1_asOwner)
	rollApp_4_byOwner_asSrc := *newRollApp("rollapp_4-1").WithAlias("four").WithOwner(creator_1_asOwner)
	rollApp_5_byOwner_asSrc := *newRollApp("rollapp_5-1").WithAlias("five").WithOwner(creator_1_asOwner)

	const aliasProhibitedTrading = "prohibited"

	const minPrice = 100
	const soExpiredEpoch = 1
	soNotExpiredEpoch := s.now.Add(time.Hour).Unix()

	tests := []struct {
		name                  string
		rollApps              []rollapp
		sellOrders            []dymnstypes.SellOrder
		expiryByAlias         []dymnstypes.ActiveSellOrdersExpirationRecord
		preMintModuleBalance  int64
		customEpochIdentifier string
		beforeHookTestFunc    func(s *KeeperTestSuite)
		wantErr               bool
		wantErrContains       string
		wantExpiryByAlias     []dymnstypes.ActiveSellOrdersExpirationRecord
		afterHookTestFunc     func(s *KeeperTestSuite)
	}{
		{
			name:     "pass - simple process expired SO without bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()

				// unchanged

				s.Equal(int64(200), s.moduleBalance())
				s.Zero(s.balance(rollApp_1_byOwner_asSrc.owner))

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
		},
		{
			name:     "pass - simple process expired & completed SO",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()

				s.Zero(s.moduleBalance())                                     // should be transferred to previous owner
				s.Equal(int64(200), s.balance(rollApp_1_byOwner_asSrc.owner)) // previous owner should earn from bid
			},
		},
		{
			name:     "pass - simple process expired & completed SO, match by min price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 250,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()

				s.Equal(int64(250-minPrice), s.moduleBalance())                    // should be transferred to previous owner
				s.Equal(int64(minPrice), s.balance(rollApp_1_byOwner_asSrc.owner)) // previous owner should earn from bid
			},
		},
		{
			name: "pass - refunds records that alias presents in params",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst,
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(aliasProhibitedTrading).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 500,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, aliasProhibitedTrading)
				s.NoError(err)

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(
					rollApp_1_byOwner_asSrc.alias, aliasProhibitedTrading,
				)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()

				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = append(p.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
						ChainId: "some-chain",
						Aliases: []string{aliasProhibitedTrading},
					})
					return p
				})
			},
			wantErr:           false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, aliasProhibitedTrading, dymnstypes.TypeAlias))

				// refunded
				s.Equal(int64(500-minPrice), s.moduleBalance())
				s.Equal(int64(minPrice), s.balance(rollApp_2_byBuyer_asDst.owner))
			},
		},
		{
			name: "pass - failed to refunds records that alias presents in params will keep the data as is",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst,
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(aliasProhibitedTrading).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 1,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, aliasProhibitedTrading)
				s.NoError(err)

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(
					rollApp_1_byOwner_asSrc.alias, aliasProhibitedTrading,
				)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()

				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = append(p.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
						ChainId: "some-chain",
						Aliases: []string{aliasProhibitedTrading},
					})
					return p
				})
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, aliasProhibitedTrading, dymnstypes.TypeAlias))

				s.Equal(int64(1), s.moduleBalance())
				s.Zero(s.balance(rollApp_2_byBuyer_asDst.owner))
			},
		},
		{
			name: "pass - process multiple - mixed SOs",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc, rollApp_4_byOwner_asSrc, rollApp_5_byOwner_asSrc,
			},
			sellOrders: []dymnstypes.SellOrder{
				// expired SO without bid
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
				// not yet finished
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soNotExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by matching sell-price
				s.newAliasSellOrder(rollApp_4_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price
				s.newAliasSellOrder(rollApp_5_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price, but prohibited trading because presents in module params
				s.newAliasSellOrder(aliasProhibitedTrading).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 450,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, aliasProhibitedTrading)
				s.NoError(err)

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(
					rollApp_1_byOwner_asSrc.alias, aliasProhibitedTrading,
				)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)

				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = append(p.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
						ChainId: "some-chain",
						Aliases: []string{aliasProhibitedTrading},
					})
					return p
				})
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// SO for alias 1 is expired without any bid/winner
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()

				// SO of the prohibited alias should be removed
				s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, aliasProhibitedTrading, dymnstypes.TypeAlias))

				// SO for alias 3 not yet finished
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp_3_byOwner_asSrc.alias, dymnstypes.TypeAlias))

				// SO for alias 4 is completed with winner
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireAlias(rollApp_4_byOwner_asSrc.alias).noActiveSO()

				// SO for alias 5 is completed with winner
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireAlias(rollApp_5_byOwner_asSrc.alias).noActiveSO()

				// aliases moved to RollApps of the winner
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).
					HasAlias(rollApp_4_byOwner_asSrc.alias, rollApp_5_byOwner_asSrc.alias)

				s.Equal(int64(50), s.moduleBalance())
				s.Equal(int64(300), s.balance(creator_1_asOwner))  // price from 2 completed SO
				s.Equal(int64(100), s.balance(creator_2_asBidder)) // refunded from prohibited trading SO
			},
		},
		{
			name: "pass - should do nothing if invalid epoch identifier",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc, rollApp_4_byOwner_asSrc, rollApp_5_byOwner_asSrc,
			},
			sellOrders: []dymnstypes.SellOrder{
				// expired SO without bid
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
				// not yet finished
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soNotExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by matching sell-price
				s.newAliasSellOrder(rollApp_4_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price
				s.newAliasSellOrder(rollApp_5_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			customEpochIdentifier: "another",
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// deep unchanged but order changed due to sorting
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)
			},
		},
		{
			name:     "pass - should remove expiry reference to non-exists SO",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(300).
					WithExpiry(soExpiredEpoch).
					Build(),
				// no SO for alias of rollapp 3
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					// no SO for alias of RollApp 3 but still have reference
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
			wantErr:           false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to alias of RollApp 1 because of processed
				// removed reference to alias of RollApp 2 because SO not exists
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
		},
		{
			name:     "pass - update expiry if in-correct",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					Build(),
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soNotExpiredEpoch). // SO not expired
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					// incorrect, SO not expired
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to alias of RollApp 1 because of processed
				// reference to alias of RollApp 3 was kept because not expired
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 1, // not enough balance
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).mustHaveActiveSO()
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs, one pass one fail",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: minPrice + 1, // just enough for the first SO
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()
				s.requireAlias(rollApp_3_byOwner_asSrc.alias).mustHaveActiveSO()

				s.EqualValues(1, s.moduleBalance())
				s.EqualValues(minPrice, s.balance(rollApp_1_byOwner_asSrc.owner))
			},
		},
		{
			name:     "pass - ignore processing SO when error occurs, one fail one pass",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice*2, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: minPrice + 1, // just enough for the second SO
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasNoAlias()

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).mustHaveActiveSO()
				s.requireAlias(rollApp_3_byOwner_asSrc.alias).noActiveSO()

				s.EqualValues(1, s.moduleBalance())
				s.EqualValues(minPrice, s.balance(rollApp_1_byOwner_asSrc.owner))
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			s.Require().NotNil(tt.beforeHookTestFunc, "mis-configured test case")
			s.Require().NotNil(tt.afterHookTestFunc, "mis-configured test case")

			if tt.preMintModuleBalance > 0 {
				s.mintToModuleAccount(tt.preMintModuleBalance)
			}

			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.expiryByAlias,
			}, dymnstypes.TypeAlias)
			s.Require().NoError(err)

			for _, rollApp := range tt.rollApps {
				s.persistRollApp(rollApp)
			}

			for _, so := range tt.sellOrders {
				err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			}

			useEpochIdentifier := s.moduleParams().Misc.EndEpochHookIdentifier
			if tt.customEpochIdentifier != "" {
				useEpochIdentifier = tt.customEpochIdentifier
			}

			tt.beforeHookTestFunc(s)

			err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, useEpochIdentifier, 1)

			defer func() {
				tt.afterHookTestFunc(s)

				aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)
				if len(tt.wantExpiryByAlias) == 0 {
					s.Empty(aSoe.Records)
				} else {
					s.Equal(tt.wantExpiryByAlias, aSoe.Records)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)

				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) Test_rollappHooks_RollappCreated() {
	const price1L = 9
	const price2L = 8
	const price3L = 7
	const price4L = 6
	const price5L = 5
	const price6L = 4
	const price7PL = 3

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.AliasPriceSteps = []sdkmath.Int{
			sdk.NewInt(price1L).Mul(priceMultiplier),
			sdk.NewInt(price2L).Mul(priceMultiplier),
			sdk.NewInt(price3L).Mul(priceMultiplier),
			sdk.NewInt(price4L).Mul(priceMultiplier),
			sdk.NewInt(price5L).Mul(priceMultiplier),
			sdk.NewInt(price6L).Mul(priceMultiplier),
			sdk.NewInt(price7PL).Mul(priceMultiplier),
		}
		return moduleParams
	})
	s.SaveCurrentContext()

	creatorAccAddr := sdk.AccAddress(testAddr(1).bytes())
	dymNameOwnerAcc := testAddr(2)
	anotherAcc := testAddr(3)

	tests := []struct {
		name                    string
		addRollApps             []string
		preRunSetup             func(s *KeeperTestSuite)
		originalCreatorBalance  int64
		originalModuleBalance   int64
		rollAppId               string
		alias                   string
		wantErr                 bool
		wantErrContains         string
		wantSuccess             bool
		wantLaterCreatorBalance int64
		postTest                func(s *KeeperTestSuite)
	}{
		{
			name:                    "pass - register without problem",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
		},
		{
			name:                   "pass - mapping RollApp ID <=> Alias should be set",
			addRollApps:            []string{"rollapp_1-1"},
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:                    "pass - if input alias is empty, do nothing",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  0,
			rollAppId:               "rollapp_1-1",
			alias:                   "",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 0,
			postTest: func(s *KeeperTestSuite) {
				// mapping should not be created

				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:                    "pass - Alias cost subtracted from creator and burned",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 1 char",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "a",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 2 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price2L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "ab",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 3 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price3L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dog",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 4 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price4L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "pool",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 5 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "angel",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 6 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price6L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "academy",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7+ chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dymension",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "fail - RollApp not exists",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not a RollApp chain-id",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:                    "fail - mapping should not be created",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				_, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "nad_0-0")
				s.False(found)

				_, found = s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
			},
		},
		{
			name:                    "fail - reject bad alias",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L,
			rollAppId:               "rollapp_1-1",
			alias:                   "@@@",
			wantErr:                 true,
			wantErrContains:         "invalid alias format",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:        "pass - can register if alias is not used",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "bridge",
							Aliases: []string{"b"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "bridge")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as alias of a chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "dym",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "dym")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is a RollApp-ID",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.pureSetRollApp(rollapptypes.Rollapp{
					RollappId: "rollapp",
					Owner:     creatorAccAddr.String(),
				})

				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "rollapp",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "rollapp")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias used by another RollApp",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "alias")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_2-2", rollAppId)
			},
		},
		{
			name:                    "fail - reject if creator does not have enough balance to pay the fee",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  1,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "insufficient funds",
			wantSuccess:             false,
			wantLaterCreatorBalance: 1,
		},
		{
			name:                   "pass - can resolve address using alias",
			addRollApps:            []string{"rollapp_1-1"},
			preRunSetup:            nil,
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(s *KeeperTestSuite) {
				dymName := dymnstypes.DymName{
					Name:       "my-name",
					Owner:      dymNameOwnerAcc.bech32(),
					Controller: dymNameOwnerAcc.bech32(),
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Value:   dymNameOwnerAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Path:    "sub",
							Value:   anotherAcc.bech32(),
						},
					},
				}
				s.setDymNameWithFunctionsAfter(dymName)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@rollapp_1-1")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@alias")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub.my-name@alias")
				s.Require().NoError(err)
				s.Equal(anotherAcc.bech32(), outputAddr)

				outputs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, anotherAcc.bech32(), "rollapp_1-1")
				s.Require().NoError(err)
				s.Require().NotEmpty(outputs)
				s.Equal("sub.my-name@alias", outputs[0].String())
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NotEqual(tt.wantSuccess, tt.wantErr, "mis-configured test case")

			s.RefreshContext()

			if tt.originalCreatorBalance > 0 {
				s.mintToAccount2(creatorAccAddr.String(), sdk.NewInt(tt.originalCreatorBalance).Mul(priceMultiplier))
			}

			if tt.originalModuleBalance > 0 {
				s.mintToModuleAccount2(sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier))
			}

			for _, rollAppId := range tt.addRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Owner:     creatorAccAddr.String(),
				})
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, tt.rollAppId, tt.alias, creatorAccAddr)

			defer func() {
				if s.T().Failed() {
					return
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(
					sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier),
					laterModuleBalance,
					"module balance should not be changed regardless of success because of burn",
				)

				if tt.postTest != nil {
					tt.postTest(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)

			laterCreatorBalance := s.balance2(creatorAccAddr.String())
			s.Equal(
				sdk.NewInt(tt.wantLaterCreatorBalance).Mul(priceMultiplier),
				laterCreatorBalance,
				"creator balance mismatch",
			)

			// event should be fired
			func() {
				if tt.alias == "" {
					return
				}

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

				for _, event := range events {
					if event.Type == dymnstypes.EventTypeSell {
						return
					}
				}

				s.T().Errorf("event %s not found", dymnstypes.EventTypeSell)
			}()
		})
	}

	s.Run("if alias is empty, do nothing", func() {
		originalTxGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, "rollapp_1-1", "", creatorAccAddr)
		s.Require().NoError(err)

		s.Equal(originalTxGas, s.ctx.GasMeter().GasConsumed(), "should not consume gas")
	})
}

func (s *KeeperTestSuite) Test_rollappHooks_OnRollAppIdChanged() {
	const previousRollAppId = "rollapp_1-1"
	const newRollAppId = "rollapp_1-2"

	const name1 = "name1"
	const name2 = "name2"

	user1Acc := testAddr(1)
	user2Acc := testAddr(2)
	user3Acc := testAddr(3)
	user4Acc := testAddr(4)
	user5Acc := testAddr(5)

	genRollApp := func(rollAppId string, aliases ...string) *rollapp {
		ra := newRollApp(rollAppId)
		for _, alias := range aliases {
			_ = ra.WithAlias(alias)
		}
		return ra
	}

	tests := []struct {
		name      string
		setupFunc func(s *KeeperTestSuite)
		testFunc  func(s *KeeperTestSuite)
	}{
		{
			name: "pass - normal migration, with alias, with Dym-Name",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "alias"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("alias")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("alias")

				//

				_, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)
			},
		},
		{
			name: "pass - normal migration, with multiple aliases, with multiple Dym-Names",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "one", "two", "three"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("one", "two", "three")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				s.setDymNameWithFunctionsAfter(
					newDN(name2, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						cfgN(previousRollAppId, "sub2", user2Acc.bech32()).
						cfgN(previousRollAppId, "sub3", user3Acc.bech32()).
						cfgN("", "", user4Acc.bech32()).
						cfgN("", "sub5", user5Acc.bech32()).
						build(),
				)

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 3) // 1 from name1, 2 from name2

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub2."+name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub3."+name2+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user3Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user3Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user4Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), previousRollAppId)
				s.NoError(err)
				if s.Len(outputDymNameAddrs, 1) {
					s.Equal(name2+"@one", outputDymNameAddrs[0].String()) // result of fallback lookup
				}

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub5."+name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user5Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs) // no fallback lookup because it's a sub-name
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("one", "two", "three")

				//

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 3) // 1 from name1, 2 from name2

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub2."+name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub3."+name2+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user3Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user3Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user4Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user4Acc.bech32(), newRollAppId)
				s.NoError(err)
				if s.Len(outputDymNameAddrs, 1) {
					s.Equal(name2+"@one", outputDymNameAddrs[0].String()) // result of fallback lookup
				}

				//

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub5."+name2+"@"+s.chainId)
				s.NoError(err)
				s.Equal(user5Acc.bech32(), outputAddr)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), s.chainId)
				s.NoError(err)
				s.Len(outputDymNameAddrs, 1)

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user5Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs) // no fallback lookup because it's a sub-name
			},
		},
		{
			name: "fail - when migrate alias failed, should not change anything",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "alias"))
				s.persistRollApp(*genRollApp(newRollAppId))

				s.requireRollApp(previousRollAppId).HasAlias("alias")
				s.requireRollApp(newRollAppId).HasNoAlias()

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)

				// delete the rollapp to make the migration fail
				s.rollAppKeeper.RemoveRollapp(s.ctx, previousRollAppId)

				dymnskeeper.ClearCaches()

				s.Require().False(s.dymNsKeeper.IsRollAppId(s.ctx, previousRollAppId))
			},
			testFunc: func(s *KeeperTestSuite) {
				// unchanged

				s.requireRollApp(newRollAppId).HasNoAlias()

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)

				outputDymNameAddrs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), previousRollAppId)
				s.NoError(err)
				s.NotEmpty(outputDymNameAddrs)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.ErrorContains(err, "not found")

				outputDymNameAddrs, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, user2Acc.bech32(), newRollAppId)
				s.NoError(err)
				s.Empty(outputDymNameAddrs)
			},
		},
		{
			name: "pass - when the new RollApp has existing aliases, merge them",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "one", "two", "three"))
				s.persistRollApp(*genRollApp(newRollAppId, "four", "five", "six"))

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				//

				s.requireRollApp(previousRollAppId).HasAlias("one", "two", "three")
				s.requireRollApp(newRollAppId).HasAlias("four", "five", "six")

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("one", "two", "three", "four", "five", "six")

				// others

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
		},
		{
			name: "pass - when previous RollApp has no alias",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId))
				s.persistRollApp(*genRollApp(newRollAppId, "new"))

				s.setDymNameWithFunctionsAfter(
					newDN(name1, user1Acc.bech32()).
						cfgN(previousRollAppId, "", user2Acc.bech32()).
						build(),
				)

				//

				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("new")

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+previousRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAlias("new")

				// others

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name1+"@"+newRollAppId)
				s.NoError(err)
				s.Equal(user2Acc.bech32(), outputAddr)
			},
		},
		{
			name: "pass - when the new RollApp has existing aliases, priority previous default alias",
			setupFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*genRollApp(previousRollAppId, "old", "journey"))
				s.persistRollApp(*genRollApp(newRollAppId, "new"))

				s.requireRollApp(previousRollAppId).HasAlias("old", "journey")
				s.requireRollApp(newRollAppId).HasAlias("new")
			},
			testFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(previousRollAppId).HasNoAlias()
				s.requireRollApp(newRollAppId).HasAliasesWithOrder("old", "new", "journey")
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.setupFunc != nil {
				tt.setupFunc(s)
			}

			err := s.dymNsKeeper.GetFutureRollAppHooks().OnRollAppIdChanged(s.ctx, previousRollAppId, newRollAppId)

			if tt.testFunc != nil {
				tt.testFunc(s)
			}

			s.Require().NoError(err)
		})
	}
}
