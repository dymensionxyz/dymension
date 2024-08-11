package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_CancelBuyOrder_DymName() {
	const minOfferPrice = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	buyerA := testAddr(1).bech32()
	anotherBuyerA := testAddr(2).bech32()
	ownerA := testAddr(3).bech32()

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		return moduleParams
	})
	s.MakeAnchorContext()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CancelBuyOrder(s.ctx, &dymnstypes.MsgCancelBuyOrder{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 1,
	}

	offer := &dymnstypes.BuyOrder{
		Id:         "101",
		AssetId:    dymName.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: minOfferPriceCoin,
	}

	offerByAnother := &dymnstypes.BuyOrder{
		Id:         "10999",
		AssetId:    dymName.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      anotherBuyerA,
		OfferPrice: minOfferPriceCoin,
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		buyer                  string
		originalModuleBalance  sdkmath.Int
		originalBuyerBalance   sdkmath.Int
		preRunSetupFunc        func(s *KeeperTestSuite)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterModuleBalance sdkmath.Int
		wantLaterBuyerBalance  sdkmath.Int
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(s *KeeperTestSuite)
	}{
		{
			name:                   "pass - can cancel offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  offer.OfferPrice.Amount,
			originalBuyerBalance:   sdk.NewInt(0),
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "pass - cancel offer will refund the buyer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  offer.OfferPrice.Amount.AddRaw(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.AddRaw(2),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                  "pass - cancel offer will remove the offer record",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer.Id))
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer.Id))
			},
		},
		{
			name:                  "pass - cancel offer will remove reverse mapping records",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				buyOrders, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offer.Buyer)
				s.Require().NoError(err)
				s.Require().Len(buyOrders, 1)

				buyOrders, err = s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, offer.AssetId)
				s.Require().NoError(err)
				s.Require().Len(buyOrders, 1)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				buyOrders, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offer.Buyer)
				s.Require().NoError(err)
				s.Require().Empty(buyOrders)

				buyOrders, err = s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, offer.AssetId)
				s.Require().NoError(err)
				s.Require().Empty(buyOrders)
			},
		},
		{
			name:                  "pass - can cancel offer when trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Misc.EnableTradingName = false
					return moduleParams
				})
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          nil,
			buyOrderId:             "102142142",
			buyer:                  buyerA,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 102142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             "102142142",
			buyer:                  offer.Buyer,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 102142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel offer with different buyer",
			existingDymName:        dymName,
			existingOffer:          offerByAnother,
			buyOrderId:             "10999",
			buyer:                  buyerA,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "not the owner of the offer",
			wantLaterOffer:         offerByAnother,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not cancel if module account does not have enough balance to refund",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			buyer:                  buyerA,
			originalModuleBalance:  sdk.NewInt(0),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "insufficient funds",
			wantLaterOffer:         offer,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.UseAnchorContext()

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalBuyerBalance.IsPositive() {
				s.mintToAccount2(tt.buyer, tt.originalBuyerBalance)
			}

			if tt.existingDymName != nil {
				err := s.dymNsKeeper.SetDymName(s.ctx, *tt.existingDymName)
				s.Require().NoError(err)
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CancelBuyOrder(s.ctx, &dymnstypes.MsgCancelBuyOrder{
				OrderId: tt.buyOrderId,
				Buyer:   tt.buyer,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Require().Equal(*tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.buyOrderId)
					s.Require().Nil(laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Require().Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.buyer)
				s.Require().Equal(tt.wantLaterBuyerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				if tt.existingDymName != nil {
					originalDymName := *tt.existingDymName
					laterDymName := s.dymNsKeeper.GetDymName(s.ctx, originalDymName.Name)
					s.Require().NotNil(laterDymName)
					s.Require().Equal(originalDymName, *laterDymName, "Dym-Name record should not be changed")
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Nil(resp)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_CancelBuyOrder_Alias() {
	const minOfferPrice = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	creator_3_asAnotherBuyer := testAddr(3).bech32()

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		return moduleParams
	})
	s.MakeAnchorContext()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CancelBuyOrder(s.ctx, &dymnstypes.MsgCancelBuyOrder{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	type rollapp struct {
		rollAppId string
		creator   string
		aliases   []string
	}

	rollApp_One_By1 := rollapp{
		rollAppId: "rollapp_1-1",
		creator:   creator_1_asOwner,
		aliases:   []string{"one"},
	}
	rollApp_Two_By2 := rollapp{
		rollAppId: "rollapp_2-2",
		creator:   creator_2_asBuyer,
		aliases:   []string{"two"},
	}
	rollApp_Three_By3 := rollapp{
		rollAppId: "rollapp_3-3",
		creator:   creator_3_asAnotherBuyer,
		aliases:   []string{},
	}

	offerAliasOfRollAppOne := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1),
		AssetId:    rollApp_One_By1.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Two_By2.rollAppId},
		Buyer:      rollApp_Two_By2.creator,
		OfferPrice: minOfferPriceCoin,
	}

	offerAliasOfRollAppOneByAnother := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    rollApp_One_By1.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Three_By3.rollAppId},
		Buyer:      rollApp_Three_By3.creator,
		OfferPrice: minOfferPriceCoin,
	}

	tests := []struct {
		name                   string
		existingRollApps       []rollapp
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		buyer                  string
		originalModuleBalance  sdkmath.Int
		originalBuyerBalance   sdkmath.Int
		preRunSetupFunc        func(s *KeeperTestSuite)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterModuleBalance sdkmath.Int
		wantLaterBuyerBalance  sdkmath.Int
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(s *KeeperTestSuite)
	}{
		{
			name:                   "pass - can cancel offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			originalBuyerBalance:   sdk.NewInt(0),
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "pass - cancel offer will refund the buyer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.AddRaw(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.AddRaw(2),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                  "pass - cancel offer will remove the offer record",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offerAliasOfRollAppOne.Id))
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offerAliasOfRollAppOne.Id))
			},
		},
		{
			name:                  "pass - cancel offer will remove reverse mapping records",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				orderIds, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offerAliasOfRollAppOne.Buyer)
				s.Require().NoError(err)
				s.Require().Len(orderIds, 1)

				orderIds, err = s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, offerAliasOfRollAppOne.AssetId)
				s.Require().NoError(err)
				s.Require().Len(orderIds, 1)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				orderIds, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offerAliasOfRollAppOne.Buyer)
				s.Require().NoError(err)
				s.Require().Empty(orderIds)

				orderIds, err = s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, offerAliasOfRollAppOne.AssetId)
				s.Require().NoError(err)
				s.Require().Empty(orderIds)
			},
		},
		{
			name:                  "pass - cancel offer will NOT remove reverse mapping records of other offers",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *offerAliasOfRollAppOneByAnother)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(
					s.ctx,
					offerAliasOfRollAppOneByAnother.Buyer,
					offerAliasOfRollAppOneByAnother.Id,
				)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(
					s.ctx,
					offerAliasOfRollAppOneByAnother.AssetId, offerAliasOfRollAppOneByAnother.AssetType,
					offerAliasOfRollAppOneByAnother.Id,
				)
				s.Require().NoError(err)

				orderIds, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offerAliasOfRollAppOne.Buyer)
				s.Require().NoError(err)
				s.Require().Len(orderIds, 1)

				orderIds, err = s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, offerAliasOfRollAppOne.AssetId)
				s.Require().NoError(err)
				s.Require().Len(orderIds, 2)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				orderIds, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, offerAliasOfRollAppOne.Buyer)
				s.Require().NoError(err)
				s.Require().Empty(orderIds)

				orderIds, err = s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, offerAliasOfRollAppOne.AssetId)
				s.Require().NoError(err)
				s.Require().Len(orderIds, 1)
				s.Require().Equal(offerAliasOfRollAppOneByAnother.Id, orderIds[0].Id)
			},
		},
		{
			name:                  "pass - can cancel offer when trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalBuyerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Misc.EnableTradingAlias = false
					return moduleParams
				})
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          nil,
			buyOrderId:             "202142142",
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 202142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             "202142142",
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 202142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel offer with different buyer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOneByAnother,
			buyOrderId:             offerAliasOfRollAppOneByAnother.Id,
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "not the owner of the offer",
			wantLaterOffer:         offerAliasOfRollAppOneByAnother,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not cancel if module account does not have enough balance to refund",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  sdk.NewInt(0),
			originalBuyerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "insufficient funds",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.UseAnchorContext()

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalBuyerBalance.IsPositive() {
				s.mintToAccount2(tt.buyer, tt.originalBuyerBalance)
			}

			for _, rollApp := range tt.existingRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: rollApp.rollAppId,
					Owner:     rollApp.creator,
				})
				for _, alias := range rollApp.aliases {
					err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp.rollAppId, alias)
					s.Require().NoError(err)
				}
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).CancelBuyOrder(s.ctx, &dymnstypes.MsgCancelBuyOrder{
				OrderId: tt.buyOrderId,
				Buyer:   tt.buyer,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Require().Equal(*tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.buyOrderId)
					s.Require().Nil(laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Require().Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.buyer)
				s.Require().Equal(tt.wantLaterBuyerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				for _, rollApp := range tt.existingRollApps {
					if len(rollApp.aliases) == 0 {
						s.requireRollApp(rollApp.rollAppId).HasNoAlias()
					} else {
						s.requireRollApp(rollApp.rollAppId).HasAlias(rollApp.aliases...)
					}
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Nil(resp)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
		})
	}
}
