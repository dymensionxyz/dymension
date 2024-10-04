package keeper_test

import (
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_CompleteSellOrder_DymName() {
	ownerAcc := testAddr(1)
	buyerAcc := testAddr(2)
	ownerA := ownerAcc.bech32()
	buyerA := buyerAcc.bech32()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
			AssetType: dymnstypes.TypeName,
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	s.Run("reject if message asset type is Unknown", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
			AssetId:     "asset",
			AssetType:   dymnstypes.AssetType_AT_UNKNOWN,
			Participant: ownerA,
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	dymName := dymnstypes.DymName{
		Name:       "my-name",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	const ownerOriginalBalance int64 = 1000
	const buyerOriginalBalance int64 = 500
	const moduleOriginalBalance int64 = 100
	const minPrice int64 = 100

	expiredSO := dymnstypes.SellOrder{
		AssetId:   dymName.Name,
		AssetType: dymnstypes.TypeName,
		MinPrice:  s.coin(minPrice),
		ExpireAt:  s.now.Add(-time.Second).Unix(),
		HighestBid: &dymnstypes.SellOrderBid{
			Bidder: buyerA,
			Price:  s.coin(minPrice),
			Params: nil,
		},
	}

	{
		// prepare test data

		s.mintToAccount(ownerA, ownerOriginalBalance)
		s.mintToAccount(buyerA, buyerOriginalBalance)
		s.mintToModuleAccount(moduleOriginalBalance)

		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		err = s.dymNsKeeper.SetSellOrder(s.ctx, expiredSO)
		s.Require().NoError(err)

		s.SaveCurrentContext()
	}

	tests := []struct {
		name            string
		participant     string
		preRunFunc      func(s *KeeperTestSuite)
		wantErr         bool
		wantErrContains string
		postRunFunc     func(s *KeeperTestSuite)
	}{
		{
			name:        "pass - can complete sell order, by owner, expired, with bid",
			participant: ownerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).
					noActiveSO().
					ownerChangedTo(buyerA)
			},
		},
		{
			name:        "pass - can complete sell order, by buyer, expired, with bid",
			participant: buyerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).
					noActiveSO().
					ownerChangedTo(buyerA)
			},
		},
		{
			name:        "pass - balance changed on success",
			participant: ownerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.Equal(
					ownerOriginalBalance+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(ownerA),
					"owner should receive the bid amount",
				)
				s.Equal(
					buyerOriginalBalance,
					s.balance(buyerA),
					"buyer should not be charged",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"module account should be charged",
				)
			},
		},
		{
			name:        "pass - after completed, ownership changed to buyer",
			participant: buyerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).
					ownerChangedTo(buyerA)
			},
		},
		{
			name:        "pass - after completed, reverse resolution should be updated",
			participant: buyerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireConfiguredAddress(ownerA).notMappedToAnyDymName()
				s.requireConfiguredAddress(buyerA).mappedDymNames(dymName.Name)
				s.requireFallbackAddress(ownerAcc.fallback()).notMappedToAnyDymName()
				s.requireFallbackAddress(buyerAcc.fallback()).mappedDymNames(dymName.Name)
			},
		},
		{
			name:        "pass - after completed, SO will be deleted",
			participant: buyerA,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).noActiveSO()
			},
		},
		{
			name:        "pass - will refund when trading was disabled, requested by owner",
			participant: ownerA,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingName = false
					return p
				})
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).noActiveSO().ownerIs(ownerA)

				s.Equal(
					ownerOriginalBalance,
					s.balance(ownerA),
					"owner should not receive the bid amount",
				)
				s.Equal(
					buyerOriginalBalance+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(buyerA),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "pass - will refund when trading was disabled, requested by buyer",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingName = false
					return p
				})
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).noActiveSO().ownerIs(ownerA)

				s.Equal(
					ownerOriginalBalance,
					s.balance(ownerA),
					"owner should not receive the bid amount",
				)
				s.Equal(
					buyerOriginalBalance+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(buyerA),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "pass - will refund when Dym-Name was expired, requested by buyer",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				existingDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName.Name)
				s.Require().NotNil(existingDymName)

				existingDymName.ExpireAt = s.now.Add(-time.Second).Unix()
				err := s.dymNsKeeper.SetDymName(s.ctx, *existingDymName)
				s.Require().NoError(err)
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireDymName(dymName.Name).noActiveSO().ownerIs(ownerA)

				s.Equal(
					ownerOriginalBalance,
					s.balance(ownerA),
					"owner should not receive the bid amount",
				)
				s.Equal(
					buyerOriginalBalance+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(buyerA),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "fail - when failed to refund, keep as is",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				// increase the highest bid amount, module balance will not be enough to refund
				{
					existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
					s.Require().NotNil(existingSO)

					existingSO.HighestBid.Price.Amount = s.coin(moduleOriginalBalance + 1).Amount
					err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
					s.Require().NoError(err)
				}

				// disable trading to toggle refund logic
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingName = false
					return p
				})
			},
			wantErr:         true,
			wantErrContains: "insufficient",
			postRunFunc: func(s *KeeperTestSuite) {
				s.Equal(ownerOriginalBalance, s.balance(ownerA))
				s.Equal(buyerOriginalBalance, s.balance(buyerA))
				s.Equal(moduleOriginalBalance, s.balance(dymNsModuleAccAddr.String()))

				s.requireDymName(dymName.Name).
					mustHaveActiveSO().
					ownerIs(ownerA)
			},
		},
		{
			name:            "fail - non-participant cannot complete",
			participant:     testAddr(0).bech32(),
			wantErr:         true,
			wantErrContains: gerrc.ErrPermissionDenied.Error(),
		},
		{
			name:        "fail - can not complete when no bid placed",
			participant: ownerA,
			preRunFunc: func(s *KeeperTestSuite) {
				existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
				s.Require().NotNil(existingSO)

				existingSO.HighestBid = nil
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "no bid placed on the Sell-Order",
		},
		{
			name:        "fail - can not complete when SO not yet expired",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
				s.Require().NotNil(existingSO)

				existingSO.ExpireAt = s.now.Add(time.Second).Unix()
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "Sell-Order not yet completed",
		},
		{
			name:        "fail - can not complete when Dym-Name does not exists",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.DeleteDymName(s.ctx, dymName.Name)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrNotFound.Error(),
		},
		{
			name:        "fail - can not complete when SO does not exists",
			participant: buyerA,
			preRunFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.DeleteSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrNotFound.Error(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			resp, errCompleteSO := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
				AssetId:     dymName.Name,
				AssetType:   dymnstypes.TypeName,
				Participant: tt.participant,
			})

			defer func() {
				if tt.postRunFunc != nil {
					tt.postRunFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(errCompleteSO, tt.wantErrContains)
				s.Nil(resp)

				s.Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCompleteSellOrder,
					"should not consume params gas on failed operation",
				)
				return
			}

			s.Require().NoError(errCompleteSO)
			s.NotNil(resp)

			s.GreaterOrEqual(
				s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCompleteSellOrder,
				"should consume params gas",
			)
		})
	}

	s.Run("independently charge gas", func() {
		s.RefreshContext()

		s.ctx.GasMeter().ConsumeGas(100_000_000, "simulate previous run")

		s.setDymNameWithFunctionsAfter(dymName)
		err := s.dymNsKeeper.SetSellOrder(s.ctx, expiredSO)
		s.Require().NoError(err)
		s.mintToModuleAccount(minPrice)

		_, err = dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
			AssetId:     dymName.Name,
			AssetType:   dymnstypes.TypeName,
			Participant: buyerA,
		})
		s.Require().NoError(err)
		s.GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), 100_000_000+dymnstypes.OpGasCompleteSellOrder,
			"should consume params gas",
		)
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_CompleteSellOrder_Alias() {
	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
			AssetType: dymnstypes.TypeAlias,
		})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()

	rollApp_1_byOwner_asSrc := *newRollApp("rollapp_1-1").WithAlias("alias").WithOwner(creator_1_asOwner)
	rollApp_2_byBuyer_asDst := *newRollApp("rollapp_2-2").WithOwner(creator_2_asBuyer)

	const originalBalanceCreator1 int64 = 1000
	const originalBalanceCreator2 int64 = 500
	const moduleOriginalBalance int64 = 100
	const minPrice int64 = 100

	expiredSO := s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
		WithMinPrice(minPrice).
		WithSellPrice(300).
		WithAliasBid(creator_2_asBuyer, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
		WithExpiry(s.now.Add(-time.Second).Unix()).
		Build()

	{
		// prepare test data

		s.mintToAccount(creator_1_asOwner, originalBalanceCreator1)
		s.mintToAccount(creator_2_asBuyer, originalBalanceCreator2)
		s.mintToModuleAccount(moduleOriginalBalance)

		s.persistRollApp(rollApp_1_byOwner_asSrc)
		s.persistRollApp(rollApp_2_byBuyer_asDst)

		err := s.dymNsKeeper.SetSellOrder(s.ctx, expiredSO)
		s.Require().NoError(err)

		s.SaveCurrentContext()
	}

	tests := []struct {
		name            string
		participant     string
		preRunFunc      func(s *KeeperTestSuite)
		wantErr         bool
		wantErrContains string
		postRunFunc     func(s *KeeperTestSuite)
	}{
		{
			name:        "pass - can complete sell order, by owner, expired, with bid",
			participant: creator_1_asOwner,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					noActiveSO().
					LinkedToRollApp(rollApp_2_byBuyer_asDst.rollAppId)
			},
		},
		{
			name:        "pass - can complete sell order, by buyer, expired, with bid",
			participant: creator_2_asBuyer,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					noActiveSO().
					LinkedToRollApp(rollApp_2_byBuyer_asDst.rollAppId)
			},
		},
		{
			name:        "pass - balance changed on success",
			participant: creator_1_asOwner,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.Equal(
					originalBalanceCreator1+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(creator_1_asOwner),
					"owner should receive the bid amount",
				)
				s.Equal(
					originalBalanceCreator2,
					s.balance(creator_2_asBuyer),
					"buyer should not be charged",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"module account should be charged",
				)
			},
		},
		{
			name:        "pass - after completed, alias linked to new RollApp",
			participant: creator_2_asBuyer,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					LinkedToRollApp(rollApp_2_byBuyer_asDst.rollAppId)
			},
		},
		{
			name:        "pass - after completed, reverse resolution should be updated",
			participant: creator_2_asBuyer,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).
					HasNoAlias()

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					LinkedToRollApp(rollApp_2_byBuyer_asDst.rollAppId)
			},
		},
		{
			name:        "pass - after completed, SO will be deleted",
			participant: creator_2_asBuyer,
			preRunFunc:  nil,
			wantErr:     false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).noActiveSO()
			},
		},
		{
			name:        "pass - will refund when trading was disabled, requested by owner",
			participant: creator_1_asOwner,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingAlias = false
					return p
				})
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					noActiveSO().
					LinkedToRollApp(rollApp_1_byOwner_asSrc.rollAppId)

				s.Equal(
					originalBalanceCreator1,
					s.balance(creator_1_asOwner),
					"owner should not receive the bid amount",
				)
				s.Equal(
					originalBalanceCreator2+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(creator_2_asBuyer),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "pass - will refund when trading was disabled, requested by buyer",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingAlias = false
					return p
				})
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					noActiveSO().
					LinkedToRollApp(rollApp_1_byOwner_asSrc.rollAppId)

				s.Equal(
					originalBalanceCreator1,
					s.balance(creator_1_asOwner),
					"owner should not receive the bid amount",
				)
				s.Equal(
					originalBalanceCreator2+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(creator_2_asBuyer),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "pass - will refund when alias was prohibited trading, requested by buyer",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				// put the alias into module params, so it will be prohibited to trade
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "pseudo",
							Aliases: []string{expiredSO.AssetId},
						},
					}
					return p
				})
			},
			wantErr: false,
			postRunFunc: func(s *KeeperTestSuite) {
				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					noActiveSO().
					LinkedToRollApp(rollApp_1_byOwner_asSrc.rollAppId)

				s.Equal(
					originalBalanceCreator1,
					s.balance(creator_1_asOwner),
					"owner should not receive the bid amount",
				)
				s.Equal(
					originalBalanceCreator2+expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(creator_2_asBuyer),
					"buyer should get the refund amount",
				)
				s.Equal(
					moduleOriginalBalance-expiredSO.HighestBid.Price.Amount.Int64(),
					s.balance(dymNsModuleAccAddr.String()),
					"refund amount should be subtracted from module account",
				)
			},
		},
		{
			name:        "fail - when failed to refund, keep as is",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				// increase the highest bid amount, module balance will not be enough to refund
				{
					existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
					s.Require().NotNil(existingSO)

					existingSO.HighestBid.Price.Amount = s.coin(moduleOriginalBalance + 1).Amount
					err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
					s.Require().NoError(err)
				}

				// disable trading to toggle refund logic
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Misc.EnableTradingAlias = false
					return p
				})
			},
			wantErr:         true,
			wantErrContains: "insufficient",
			postRunFunc: func(s *KeeperTestSuite) {
				s.Equal(originalBalanceCreator1, s.balance(creator_1_asOwner))
				s.Equal(originalBalanceCreator2, s.balance(creator_2_asBuyer))
				s.Equal(moduleOriginalBalance, s.balance(dymNsModuleAccAddr.String()))

				s.requireAlias(rollApp_1_byOwner_asSrc.alias).
					mustHaveActiveSO().
					LinkedToRollApp(rollApp_1_byOwner_asSrc.rollAppId)
			},
		},
		{
			name:            "fail - non-participant cannot complete",
			participant:     testAddr(0).bech32(),
			wantErr:         true,
			wantErrContains: gerrc.ErrPermissionDenied.Error(),
		},
		{
			name:        "fail - can not complete when no bid placed",
			participant: creator_1_asOwner,
			preRunFunc: func(s *KeeperTestSuite) {
				existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
				s.Require().NotNil(existingSO)

				existingSO.HighestBid = nil
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "no bid placed on the Sell-Order",
		},
		{
			name:        "fail - can not complete when SO not yet expired",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
				s.Require().NotNil(existingSO)

				existingSO.ExpireAt = s.now.Add(time.Second).Unix()
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *existingSO)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "Sell-Order not yet completed",
		},
		{
			name:        "fail - can not complete when alias no longer linking to any RollApp",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.RemoveAliasFromRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, expiredSO.AssetId)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "alias is not in-use",
		},
		{
			name:        "fail - can not complete when the RollApp which alias was linked to, does not exists (deleted)",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				s.rollAppKeeper.RemoveRollapp(s.ctx, rollApp_1_byOwner_asSrc.rollAppId)
			},
			wantErr:         true,
			wantErrContains: "not found",
		},
		{
			name:        "fail - can not complete when the RollApp which should be linked to, does not exists (deleted)",
			participant: creator_2_asBuyer,
			preRunFunc: func(s *KeeperTestSuite) {
				existingSO := s.dymNsKeeper.GetSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
				s.Require().NotNil(existingSO)

				s.rollAppKeeper.RemoveRollapp(s.ctx, existingSO.HighestBid.Params[0])
			},
			wantErr:         true,
			wantErrContains: "destination Roll-App does not exists",
		},
		{
			name:        "fail - can not complete when SO does not exists",
			participant: creator_1_asOwner,
			preRunFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.DeleteSellOrder(s.ctx, expiredSO.AssetId, expiredSO.AssetType)
			},
			wantErr:         true,
			wantErrContains: gerrc.ErrNotFound.Error(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			// test

			resp, errCompleteSO := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
				AssetId:     expiredSO.AssetId,
				AssetType:   dymnstypes.TypeAlias,
				Participant: tt.participant,
			})

			defer func() {
				if tt.postRunFunc != nil {
					tt.postRunFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(errCompleteSO, tt.wantErrContains)
				s.Nil(resp)

				s.Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCompleteSellOrder,
					"should not consume params gas on failed operation",
				)

				return
			}

			s.Require().NoError(errCompleteSO)
			s.NotNil(resp)
			s.GreaterOrEqual(
				s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasCompleteSellOrder,
				"should consume params gas",
			)
		})
	}

	s.Run("independently charge gas", func() {
		s.RefreshContext()

		s.ctx.GasMeter().ConsumeGas(100_000_000, "simulate previous run")

		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CompleteSellOrder(s.ctx, &dymnstypes.MsgCompleteSellOrder{
			AssetId:     rollApp_1_byOwner_asSrc.alias,
			AssetType:   dymnstypes.TypeAlias,
			Participant: creator_2_asBuyer,
		})
		s.Require().NoError(err)

		s.Require().GreaterOrEqual(
			s.ctx.GasMeter().GasConsumed(), 100_000_000+dymnstypes.OpGasCompleteSellOrder,
			"gas consumption should be stacked",
		)
	})
}
