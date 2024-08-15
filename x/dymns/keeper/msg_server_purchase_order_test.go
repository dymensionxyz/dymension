package keeper_test

import (
	"fmt"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_PurchaseOrder_DymName() {
	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
			AssetType: dymnstypes.TypeName,
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	s.Run("reject if message asset type is Unknown", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.AssetType_AT_UNKNOWN,
			Buyer:     testAddr(0).bech32(),
			Offer:     s.coin(1),
		})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	previousBidderA := testAddr(3).bech32()

	originalDymNameExpiry := s.now.Unix() + 100
	dymName := dymnstypes.DymName{
		Name:       "my-name",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
	}

	const ownerOriginalBalance int64 = 1000
	const buyerOriginalBalance int64 = 500
	const previousBidderOriginalBalance int64 = 400
	const minPrice int64 = 100
	tests := []struct {
		name                           string
		withoutDymName                 bool
		withoutSellOrder               bool
		expiredSellOrder               bool
		sellPrice                      int64
		previousBid                    int64
		skipPreMintModuleAccount       bool
		overrideBuyerOriginalBalance   int64
		customBuyer                    string
		newBid                         int64
		customBidDenom                 string
		wantOwnershipChanged           bool
		wantErr                        bool
		wantErrContains                string
		wantOwnerBalanceLater          int64
		wantBuyerBalanceLater          int64
		wantPreviousBidderBalanceLater int64
	}{
		{
			name:                           "fail - Dym-Name does not exists, SO does not exists",
			withoutDymName:                 true,
			withoutSellOrder:               true,
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                fmt.Sprintf("Dym-Name: %s: not found", dymName.Name),
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - Dym-Name does not exists, SO exists",
			withoutDymName:                 true,
			withoutSellOrder:               false,
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                fmt.Sprintf("Dym-Name: %s: not found", dymName.Name),
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - Dym-Name exists, SO does not exists",
			withoutDymName:                 false,
			withoutSellOrder:               true,
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                fmt.Sprintf("Sell-Order: %s: not found", dymName.Name),
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - self-purchase is not allowed",
			customBuyer:                    ownerA,
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                "cannot purchase your own dym name",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - invalid buyer address",
			customBuyer:                    "invalidAddress",
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                "buyer is not a valid bech32 account address",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase an expired order, no bid",
			expiredSellOrder:               true,
			newBid:                         100,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, expired, with bid, without sell price",
			expiredSellOrder:               true,
			sellPrice:                      0,
			previousBid:                    200,
			newBid:                         300,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, expired, with sell price, with bid under sell price",
			expiredSellOrder:               true,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         300,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, expired, with sell price, with bid = sell price",
			expiredSellOrder:               true,
			sellPrice:                      300,
			previousBid:                    300,
			newBid:                         300,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, not expired, fail because previous bid matches sell price",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    300,
			newBid:                         300,
			wantErr:                        true,
			wantErrContains:                "cannot purchase a completed order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase order, not expired, fail because lower than previous bid",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200 - 1,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase order, not expired, fail because equals to previous bid",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, expired, bid equals to previous bid",
			expiredSellOrder:               true,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - purchase a completed order, expired, bid lower than previous bid",
			expiredSellOrder:               true,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200 - 1,
			wantErr:                        true,
			wantErrContains:                "cannot purchase an expired order",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - mis-match denom",
			expiredSellOrder:               false,
			newBid:                         200,
			customBidDenom:                 "u" + params.BaseDenom,
			wantErr:                        true,
			wantErrContains:                "offer denom does not match the order denom",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer lower than min-price",
			expiredSellOrder:               false,
			newBid:                         minPrice - 1,
			wantErr:                        true,
			wantErrContains:                "offer is lower than minimum price",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - zero bid amount",
			expiredSellOrder:               false,
			newBid:                         0,
			wantErr:                        true,
			wantErrContains:                "offer must be positive",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer higher than sell-price",
			expiredSellOrder:               false,
			sellPrice:                      300,
			newBid:                         300 + 1,
			wantErr:                        true,
			wantErrContains:                "offer is higher than sell price",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer equals to previous bid, no sell price",
			expiredSellOrder:               false,
			previousBid:                    200,
			newBid:                         200,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer lower than previous bid, no sell price",
			expiredSellOrder:               false,
			previousBid:                    200,
			newBid:                         200 - 1,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer equals to previous bid, has sell price",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - offer lower than previous bid, has sell price",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    200,
			newBid:                         200 - 1,
			wantErr:                        true,
			wantErrContains:                "new offer must be higher than current highest bid",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "pass - place bid, = min price, no previous bid, no sell price",
			expiredSellOrder:               false,
			newBid:                         minPrice,
			wantOwnershipChanged:           false,
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance - minPrice,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "pass - place bid, greater than previous bid, no sell price",
			expiredSellOrder:               false,
			previousBid:                    minPrice,
			newBid:                         minPrice + 1,
			wantOwnershipChanged:           false,
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance - (minPrice + 1),
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance + minPrice, // refund
		},
		{
			name:                           "fail - failed to refund previous bid",
			expiredSellOrder:               false,
			previousBid:                    minPrice,
			skipPreMintModuleAccount:       true,
			newBid:                         minPrice + 1,
			wantOwnershipChanged:           false,
			wantErr:                        true,
			wantErrContains:                "insufficient funds",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "fail - insufficient buyer funds",
			expiredSellOrder:               false,
			overrideBuyerOriginalBalance:   1,
			newBid:                         minPrice + 1,
			wantOwnershipChanged:           false,
			wantErr:                        true,
			wantErrContains:                "insufficient funds",
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          1,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance,
		},
		{
			name:                           "pass - place bid, greater than previous bid, under sell price",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    minPrice,
			newBid:                         300 - 1,
			wantOwnershipChanged:           false,
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance - (300 - 1),
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance + minPrice, // refund
		},
		{
			name:                           "pass - place bid, greater than previous bid, equals sell price, transfer ownership",
			expiredSellOrder:               false,
			sellPrice:                      300,
			previousBid:                    minPrice,
			newBid:                         300,
			wantOwnershipChanged:           true,
			wantOwnerBalanceLater:          ownerOriginalBalance + 300,
			wantBuyerBalanceLater:          buyerOriginalBalance - 300,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance + minPrice, // refund
		},
		{
			name:                           "pass - refund previous bidder",
			expiredSellOrder:               false,
			previousBid:                    minPrice,
			newBid:                         200,
			wantOwnershipChanged:           false,
			wantOwnerBalanceLater:          ownerOriginalBalance,
			wantBuyerBalanceLater:          buyerOriginalBalance - 200,
			wantPreviousBidderBalanceLater: previousBidderOriginalBalance + minPrice, // refund
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			useOwnerOriginalBalance := ownerOriginalBalance
			useBuyerOriginalBalance := buyerOriginalBalance
			if tt.overrideBuyerOriginalBalance > 0 {
				useBuyerOriginalBalance = tt.overrideBuyerOriginalBalance
			}
			usePreviousBidderOriginalBalance := previousBidderOriginalBalance

			s.mintToAccount(ownerA, useOwnerOriginalBalance)
			s.mintToAccount(buyerA, useBuyerOriginalBalance)
			s.mintToAccount(previousBidderA, usePreviousBidderOriginalBalance)

			dymName.Configs = []dymnstypes.DymNameConfig{
				{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				},
			}

			if !tt.withoutDymName {
				s.setDymNameWithFunctionsAfter(dymName)
			}

			so := dymnstypes.SellOrder{
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				MinPrice:  s.coin(minPrice),
			}

			if tt.expiredSellOrder {
				so.ExpireAt = s.now.Unix() - 1
			} else {
				so.ExpireAt = s.now.Unix() + 100
			}

			s.Require().GreaterOrEqual(tt.sellPrice, int64(0), "bad setup")
			if tt.sellPrice > 0 {
				so.SellPrice = uptr.To(s.coin(tt.sellPrice))
			}

			s.Require().GreaterOrEqual(tt.previousBid, int64(0), "bad setup")
			if tt.previousBid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: previousBidderA,
					Price:  s.coin(tt.previousBid),
				}

				// mint coin to module account because we charged bidder before update SO
				if !tt.skipPreMintModuleAccount {
					err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
					s.Require().NoError(err)
				}
			}

			if !tt.withoutSellOrder {
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			}

			// test

			s.Require().GreaterOrEqual(tt.newBid, int64(0), "mis-configured test case")
			useBuyer := buyerA
			if tt.customBuyer != "" {
				useBuyer = tt.customBuyer
			}
			useDenom := params.BaseDenom
			if tt.customBidDenom != "" {
				useDenom = tt.customBidDenom
			}
			resp, errPurchaseName := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				Offer:     sdk.NewInt64Coin(useDenom, tt.newBid),
				Buyer:     useBuyer,
			})
			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName.Name)
			if !tt.withoutDymName {
				s.Require().NotNil(laterDymName)
				s.Require().Equal(dymName.Name, laterDymName.Name, "name should not be changed")
				s.Require().Equal(originalDymNameExpiry, laterDymName.ExpireAt, "expiry should not be changed")
			}

			laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, dymName.Name, dymnstypes.TypeName)
			laterOwnerBalance := s.balance(ownerA)
			laterBuyerBalance := s.balance(buyerA)
			laterPreviousBidderBalance := s.balance(previousBidderA)
			laterDymNamesOwnedByOwner, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
			s.Require().NoError(err)
			laterDymNamesOwnedByBuyer, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, buyerA)
			s.Require().NoError(err)
			laterDymNamesOwnedByPreviousBidder, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, previousBidderA)
			s.Require().NoError(err)

			s.Require().Equal(tt.wantOwnerBalanceLater, laterOwnerBalance, "owner balance mis-match")
			s.Require().Equal(tt.wantBuyerBalanceLater, laterBuyerBalance, "buyer balance mis-match")
			s.Require().Equal(tt.wantPreviousBidderBalanceLater, laterPreviousBidderBalance, "previous bidder balance mis-match")

			s.Require().Empty(laterDymNamesOwnedByPreviousBidder, "no reverse record should be made for previous bidder")

			if tt.wantErr {
				s.Require().Error(errPurchaseName, "action should be failed")
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Contains(errPurchaseName.Error(), tt.wantErrContains)
				s.Require().Nil(resp)

				s.Require().False(tt.wantOwnershipChanged, "mis-configured test case")

				s.Require().Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should not consume params gas on failed operation",
				)
			} else {
				s.Require().NoError(errPurchaseName, "action should be successful")
				s.Require().NotNil(resp)

				s.Require().GreaterOrEqual(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should consume params gas",
				)
			}

			if tt.wantOwnershipChanged {
				s.Require().False(tt.withoutDymName, "mis-configured test case")
				s.Require().False(tt.withoutSellOrder, "mis-configured test case")

				s.Require().Nil(laterSo, "SO should be deleted")

				s.Require().Equal(buyerA, laterDymName.Owner, "ownership should be changed")
				s.Require().Equal(buyerA, laterDymName.Controller, "controller should be changed")
				s.Require().Empty(laterDymName.Configs, "configs should be cleared")
				s.Require().Empty(laterDymNamesOwnedByOwner, "reverse record should be removed")
				s.Require().Len(laterDymNamesOwnedByBuyer, 1, "reverse record should be added")
			} else {
				if tt.withoutDymName {
					s.Require().Nil(laterDymName)
					s.Require().Empty(laterDymNamesOwnedByOwner)
					s.Require().Empty(laterDymNamesOwnedByBuyer)
				} else {
					s.Require().Equal(ownerA, laterDymName.Owner, "ownership should not be changed")
					s.Require().Equal(ownerA, laterDymName.Controller, "controller should not be changed")
					s.Require().NotEmpty(laterDymName.Configs, "configs should be kept")
					s.Require().Equal(dymName.Configs, laterDymName.Configs, "configs not be changed")
					s.Require().Len(laterDymNamesOwnedByOwner, 1, "reverse record should be kept")
					s.Require().Empty(laterDymNamesOwnedByBuyer, "reverse record should not be added")
				}

				if tt.withoutSellOrder {
					s.Require().Nil(laterSo)
				} else {
					s.Require().NotNil(laterSo, "SO should not be deleted")
				}
			}
		})
	}

	s.Run("reject purchase order when trading is disabled", func() {
		s.RefreshContext()

		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Misc.EnableTradingName = false
			return moduleParams
		})

		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
			AssetId:   "my-name",
			AssetType: dymnstypes.TypeName,
			Offer:     s.coin(100),
			Buyer:     buyerA,
		})
		s.Require().ErrorContains(err, "unmet precondition")
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_PurchaseOrder_Alias() {
	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
			AssetType: dymnstypes.TypeAlias,
		})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	creator_3_asAnotherBuyer := testAddr(3).bech32()

	rollApp_1_byOwner_asSrc := *newRollApp("rollapp_1-1").WithAlias("alias").WithOwner(creator_1_asOwner)
	rollApp_2_byBuyer_asDst := *newRollApp("rollapp_2-2").WithOwner(creator_2_asBuyer)
	rollApp_3_byAnotherBuyer_asDst := *newRollApp("rollapp_3-2").WithOwner(creator_3_asAnotherBuyer)
	rollApp_4_byOwner_asDst := *newRollApp("rollapp_4-2").WithOwner(creator_1_asOwner)

	const originalBalanceCreator1 int64 = 1000
	const originalBalanceCreator2 int64 = 500
	const originalBalanceCreator3 int64 = 400
	const minPrice int64 = 100

	msg := func(buyer string, offer int64, assetId, dstRollAppId string) dymnstypes.MsgPurchaseOrder {
		return dymnstypes.MsgPurchaseOrder{
			AssetId:   assetId,
			AssetType: dymnstypes.TypeAlias,
			Params:    []string{dstRollAppId},
			Offer:     s.coin(offer),
			Buyer:     buyer,
		}
	}

	tests := []struct {
		name                            string
		rollApps                        []rollapp
		sellOrder                       *dymnstypes.SellOrder
		sourceRollAppId                 string
		skipPreMintModuleAccount        bool
		overrideOriginalBalanceCreator2 int64
		msg                             dymnstypes.MsgPurchaseOrder
		preRunFunc                      func(s *KeeperTestSuite)
		wantCompleted                   bool
		wantErr                         bool
		wantErrContains                 string
		wantLaterBalanceCreator1        int64
		wantLaterBalanceCreator2        int64
		wantLaterBalanceCreator3        int64
	}{
		{
			name:      "fail - source Alias/RollApp does not exists, SO does not exists",
			rollApps:  []rollapp{rollApp_2_byBuyer_asDst},
			sellOrder: nil,
			msg: msg(
				creator_2_asBuyer, 100,
				"void", rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantCompleted:   false,
			wantErr:         true,
			wantErrContains: "alias not owned by any RollApp",
		},
		{
			name:      "fail - destination RollApp does not exists, SO does not exists",
			rollApps:  nil,
			sellOrder: nil,
			msg: msg(
				creator_2_asBuyer, 100,
				"void", rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantCompleted:   false,
			wantErr:         true,
			wantErrContains: "destination Roll-App does not exists",
		},
		{
			name:     "fail - source Alias/RollApp does not exists, SO exists",
			rollApps: []rollapp{rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder("void").
				WithMinPrice(minPrice).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, minPrice,
				"void", rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "alias not owned by any RollApp",
		},
		{
			name:     "fail - destination Alias/RollApp does not exists, SO exists",
			rollApps: nil,
			sellOrder: s.newAliasSellOrder("void").
				WithMinPrice(minPrice).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, minPrice,
				"void", rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "destination Roll-App does not exists",
		},
		{
			name:      "fail - Alias/RollApp exists, SO does not exists",
			rollApps:  []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: nil,
			msg: msg(
				creator_2_asBuyer, 100,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Sell-Order: %s: not found", rollApp_1_byOwner_asSrc.alias),
		},
		{
			name:     "pass - self-purchase is allowed",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_4_byOwner_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				BuildP(),
			msg: msg(
				creator_1_asOwner, minPrice,
				rollApp_1_byOwner_asSrc.alias, rollApp_4_byOwner_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            false,
			wantLaterBalanceCreator1: originalBalanceCreator1 - minPrice,
			wantLaterBalanceCreator2: originalBalanceCreator2,
			wantLaterBalanceCreator3: originalBalanceCreator3,
		},
		{
			name:     "pass - self-purchase is allowed, complete order",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_4_byOwner_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(200).
				BuildP(),
			sourceRollAppId: rollApp_1_byOwner_asSrc.rollAppId,
			msg: msg(
				creator_1_asOwner, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_4_byOwner_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            true,
			wantLaterBalanceCreator1: originalBalanceCreator1, // unchanged
			wantLaterBalanceCreator2: originalBalanceCreator2,
			wantLaterBalanceCreator3: originalBalanceCreator3,
		},
		{
			name:      "fail - invalid buyer address",
			rollApps:  []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).WithMinPrice(minPrice).BuildP(),
			msg: msg(
				"invalidAddress", minPrice,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:     "fail - buyer is not the owner of the destination RollApp",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).WithMinPrice(minPrice).
				Expired().
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_3_byAnotherBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "not the owner of the RollApp",
		},
		{
			name:     "fail - destination RollApp is the same as source",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).WithMinPrice(minPrice).
				Expired().
				BuildP(),
			msg: msg(
				creator_1_asOwner, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_1_byOwner_asSrc.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "destination Roll-App ID is the same as the source",
		},
		{
			name:     "fail - purchase an expired order, no bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).WithMinPrice(minPrice).
				Expired().
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase a completed order, expired, with bid, without sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 150, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase a completed order, expired, with sell price, with bid under sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 150, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase a completed order, expired, with sell price, with bid = sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 300 /*equals sell price*/, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 300, /*equals sell price*/
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase a completed order, not expired, fail because previous bid matches sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, 300 /*equals sell price*/, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 300, /*equals sell price*/
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase a completed order",
		},
		{
			name:     "fail - purchase order, not expired, fail because lower than previous bid, with sell-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, 250, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "new offer must be higher than current highest bid",
		},
		{
			name:     "fail - purchase order, not expired, fail because lower than previous bid, without sell-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				// without sell-price
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 150,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "new offer must be higher than current highest bid",
		},
		{
			name:     "fail - purchase order, expired, fail because lower than previous bid, with sell-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 300, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase order, expired, fail because lower than previous bid, without sell-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				// without sell-price
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 150,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - purchase order, not expired, fail because equals to previous bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "new offer must be higher than current highest bid",
		},
		{
			name:     "fail - purchase a completed order, expired, bid equals to previous bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byAnotherBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				Expired().
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "cannot purchase an expired order",
		},
		{
			name:     "fail - mis-match denom",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).WithSellPrice(300).
				BuildP(),
			msg: func() dymnstypes.MsgPurchaseOrder {
				msg := msg(
					creator_2_asBuyer, 200,
					rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
				)
				msg.Offer = sdk.Coin{
					Denom:  "u" + params.BaseDenom,
					Amount: msg.Offer.Amount,
				}
				return msg
			}(),
			wantErr:         true,
			wantErrContains: "offer denom does not match the order denom",
		},
		{
			name:     "fail - offer lower than min-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 1, // very low
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "offer is lower than minimum price",
		},
		{
			name:     "fail - zero bid amount",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 0,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:     "fail - offer higher than sell-price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 300+1,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:         true,
			wantErrContains: "offer is higher than sell price",
		},
		{
			name:     "pass - place bid, = min price, no previous bid, no sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				BuildP(),
			sourceRollAppId: rollApp_1_byOwner_asSrc.rollAppId,
			msg: msg(
				creator_2_asBuyer, minPrice,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            false,
			wantLaterBalanceCreator1: originalBalanceCreator1,
			wantLaterBalanceCreator2: originalBalanceCreator2 - minPrice,
			wantLaterBalanceCreator3: originalBalanceCreator3,
		},
		{
			name:     "fail - can not purchase if alias is presents in params",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				BuildP(),
			sourceRollAppId: rollApp_1_byOwner_asSrc.rollAppId,
			msg: msg(
				creator_2_asBuyer, minPrice,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "some-chain",
							Aliases: []string{rollApp_1_byOwner_asSrc.alias},
						},
					}
					return p
				})
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantCompleted:   false,
		},
		{
			name:     "pass - place bid, greater than previous bid, no sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			sourceRollAppId: rollApp_1_byOwner_asSrc.rollAppId,
			msg: msg(
				creator_2_asBuyer, 250,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            false,
			wantLaterBalanceCreator1: originalBalanceCreator1,
			wantLaterBalanceCreator2: originalBalanceCreator2 - 250,
			wantLaterBalanceCreator3: originalBalanceCreator3 + 200, // refund
		},
		{
			name:     "fail - failed to refund previous bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 250,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			skipPreMintModuleAccount: true,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:     "fail - insufficient buyer funds",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithAliasBid(creator_3_asAnotherBuyer, 200, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 250,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			overrideOriginalBalanceCreator2: 1,
			wantErr:                         true,
			wantErrContains:                 "insufficient funds",
		},
		{
			name:     "pass - place bid, greater than previous bid, under sell price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, minPrice, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 250,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            false,
			wantLaterBalanceCreator1: originalBalanceCreator1,
			wantLaterBalanceCreator2: originalBalanceCreator2 - 250,      // charge bid
			wantLaterBalanceCreator3: originalBalanceCreator3 + minPrice, // refund
		},
		{
			name:     "pass - place bid, greater than previous bid, equals sell price, transfer ownership",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, minPrice, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			sourceRollAppId: rollApp_1_byOwner_asSrc.rollAppId,
			msg: msg(
				creator_2_asBuyer, 300,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            true,
			wantLaterBalanceCreator1: originalBalanceCreator1 + 300,      // transfer sale
			wantLaterBalanceCreator2: originalBalanceCreator2 - 300,      // charge bid
			wantLaterBalanceCreator3: originalBalanceCreator3 + minPrice, // refund
		},
		{
			name:     "pass - if any bid before, later bid higher, refund previous bidder",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrder: s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
				WithMinPrice(minPrice).
				WithSellPrice(300).
				WithAliasBid(creator_3_asAnotherBuyer, minPrice, rollApp_3_byAnotherBuyer_asDst.rollAppId).
				BuildP(),
			msg: msg(
				creator_2_asBuyer, 200,
				rollApp_1_byOwner_asSrc.alias, rollApp_2_byBuyer_asDst.rollAppId,
			),
			wantErr:                  false,
			wantCompleted:            false,
			wantLaterBalanceCreator1: originalBalanceCreator1,
			wantLaterBalanceCreator2: originalBalanceCreator2 - 200,      // charge bid
			wantLaterBalanceCreator3: originalBalanceCreator3 + minPrice, // refund
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			useOriginalBalanceCreator1 := originalBalanceCreator1
			useOriginalBalanceCreator2 := originalBalanceCreator2
			if tt.overrideOriginalBalanceCreator2 > 0 {
				useOriginalBalanceCreator2 = tt.overrideOriginalBalanceCreator2
			}
			useOriginalBalanceCreator3 := originalBalanceCreator3

			s.mintToAccount(creator_1_asOwner, useOriginalBalanceCreator1)
			s.mintToAccount(creator_2_asBuyer, useOriginalBalanceCreator2)
			s.mintToAccount(creator_3_asAnotherBuyer, useOriginalBalanceCreator3)

			for _, rollApp := range tt.rollApps {
				s.persistRollApp(rollApp)
			}

			if tt.sellOrder != nil {
				s.Require().Equal(tt.sellOrder.AssetId, tt.msg.AssetId, "bad setup")

				err := s.dymNsKeeper.SetSellOrder(s.ctx, *tt.sellOrder)
				s.Require().NoError(err)

				if tt.sellOrder.HighestBid != nil {
					if !tt.skipPreMintModuleAccount {
						s.mintToModuleAccount(tt.sellOrder.HighestBid.Price.Amount.Int64())
					}
				}
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			// test

			resp, errPurchaseName := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &tt.msg)

			for _, ra := range tt.rollApps {
				s.True(s.dymNsKeeper.IsRollAppId(s.ctx, ra.rollAppId))
			}

			laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, tt.msg.AssetId, dymnstypes.TypeAlias)

			if tt.wantErr {
				s.Require().ErrorContains(errPurchaseName, tt.wantErrContains)
				s.Nil(resp)
				s.Require().False(tt.wantCompleted, "mis-configured test case")
				s.Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should not consume params gas on failed operation",
				)

				s.Zero(tt.wantLaterBalanceCreator1, "bad setup, won't check balance on error")
				s.Zero(tt.wantLaterBalanceCreator2, "bad setup, won't check balance on error")
				s.Zero(tt.wantLaterBalanceCreator3, "bad setup, won't check balance on error")
			} else {
				s.NotNil(resp)
				s.GreaterOrEqual(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should consume params gas",
				)

				s.Equal(tt.wantLaterBalanceCreator1, s.balance(creator_1_asOwner), "owner balance mis-match")
				s.Equal(tt.wantLaterBalanceCreator2, s.balance(creator_2_asBuyer), "buyer balance mis-match")
				s.Equal(tt.wantLaterBalanceCreator3, s.balance(creator_3_asAnotherBuyer), "previous bidder balance mis-match")
			}

			destinationRollAppId := tt.msg.Params[0]
			if tt.wantCompleted {
				s.Require().NotEmpty(tt.sourceRollAppId, "mis-configured test case")

				s.Require().NotEmpty(tt.rollApps, "mis-configured test case")
				s.Require().NotNil(tt.sellOrder, "mis-configured test case")

				s.Nil(laterSo, "SO should be deleted")

				s.requireRollApp(tt.sourceRollAppId).HasNoAlias()
				s.requireRollApp(destinationRollAppId).HasAlias(tt.msg.AssetId)
			} else {
				if len(tt.rollApps) > 0 {
					for _, ra := range tt.rollApps {
						if ra.alias != "" {
							s.requireRollApp(ra.rollAppId).HasAlias(ra.alias)
						} else {
							s.requireRollApp(ra.rollAppId).HasNoAlias()
						}
					}
				}

				if tt.sellOrder != nil {
					s.NotNil(laterSo, "SO should not be deleted")
				} else {
					s.Nil(laterSo, "SO should not exists")
				}
			}
		})
	}

	s.Run("reject purchase order when trading is disabled", func() {
		s.RefreshContext()

		moduleParams := s.dymNsKeeper.GetParams(s.ctx)
		moduleParams.Misc.EnableTradingAlias = false
		err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
		s.Require().NoError(err)

		_, err = dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
			AssetId:   "alias",
			AssetType: dymnstypes.TypeAlias,
			Params:    []string{rollApp_2_byBuyer_asDst.rollAppId},
			Offer:     s.coin(100),
			Buyer:     creator_2_asBuyer,
		})
		s.Require().ErrorContains(err, "trading of Alias is disabled")
	})
}
