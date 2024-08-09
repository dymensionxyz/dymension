package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_PurchaseOrder_DymName(t *testing.T) {
	now := time.Now().UTC()
	futureEpoch := now.Unix() + 1

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		// force enable trading
		moduleParams := dk.GetParams(ctx)
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		require.NoError(t, dk.SetParams(ctx, moduleParams))

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PurchaseOrder(ctx, &dymnstypes.MsgPurchaseOrder{
				OrderType: dymnstypes.NameOrder,
			})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	t.Run("reject if message order type is Unknown", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PurchaseOrder(ctx, &dymnstypes.MsgPurchaseOrder{
				GoodsId:   "goods",
				OrderType: dymnstypes.OrderType_OT_UNKNOWN,
				Buyer:     testAddr(0).bech32(),
				Offer:     dymnsutils.TestCoin(1),
			})
			return err
		}, "invalid order type")
	})

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	previousBidderA := testAddr(3).bech32()

	originalDymNameExpiry := futureEpoch
	dymName := dymnstypes.DymName{
		Name:       "my-name",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   originalDymNameExpiry,
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
		t.Run(tt.name, func(t *testing.T) {
			// setup execution context
			dk, bk, ctx := setupTest()

			useOwnerOriginalBalance := ownerOriginalBalance
			useBuyerOriginalBalance := buyerOriginalBalance
			if tt.overrideBuyerOriginalBalance > 0 {
				useBuyerOriginalBalance = tt.overrideBuyerOriginalBalance
			}
			usePreviousBidderOriginalBalance := previousBidderOriginalBalance

			err := bk.MintCoins(ctx,
				dymnstypes.ModuleName,
				dymnsutils.TestCoins(
					useOwnerOriginalBalance+useBuyerOriginalBalance+usePreviousBidderOriginalBalance,
				),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(ownerA),
				dymnsutils.TestCoins(useOwnerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(buyerA),
				dymnsutils.TestCoins(useBuyerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(previousBidderA),
				dymnsutils.TestCoins(usePreviousBidderOriginalBalance),
			)
			require.NoError(t, err)

			dymName.Configs = []dymnstypes.DymNameConfig{
				{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				},
			}

			if !tt.withoutDymName {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			so := dymnstypes.SellOrder{
				GoodsId:  dymName.Name,
				Type:     dymnstypes.NameOrder,
				MinPrice: dymnsutils.TestCoin(minPrice),
			}

			if tt.expiredSellOrder {
				so.ExpireAt = now.Unix() - 1
			} else {
				so.ExpireAt = futureEpoch
			}

			require.GreaterOrEqual(t, tt.sellPrice, int64(0), "bad setup")
			if tt.sellPrice > 0 {
				so.SellPrice = dymnsutils.TestCoinP(tt.sellPrice)
			}

			require.GreaterOrEqual(t, tt.previousBid, int64(0), "bad setup")
			if tt.previousBid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: previousBidderA,
					Price:  dymnsutils.TestCoin(tt.previousBid),
				}

				// mint coin to module account because we charged bidder before update SO
				if !tt.skipPreMintModuleAccount {
					err = bk.MintCoins(ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
					require.NoError(t, err)
				}
			}

			if !tt.withoutSellOrder {
				err = dk.SetSellOrder(ctx, so)
				require.NoError(t, err)
			}

			// test

			require.GreaterOrEqual(t, tt.newBid, int64(0), "mis-configured test case")
			useBuyer := buyerA
			if tt.customBuyer != "" {
				useBuyer = tt.customBuyer
			}
			useDenom := params.BaseDenom
			if tt.customBidDenom != "" {
				useDenom = tt.customBidDenom
			}
			resp, errPurchaseName := dymnskeeper.NewMsgServerImpl(dk).PurchaseOrder(ctx, &dymnstypes.MsgPurchaseOrder{
				GoodsId:   dymName.Name,
				OrderType: dymnstypes.NameOrder,
				Offer:     sdk.NewInt64Coin(useDenom, tt.newBid),
				Buyer:     useBuyer,
			})
			laterDymName := dk.GetDymName(ctx, dymName.Name)
			if !tt.withoutDymName {
				require.NotNil(t, laterDymName)
				require.Equal(t, dymName.Name, laterDymName.Name, "name should not be changed")
				require.Equal(t, originalDymNameExpiry, laterDymName.ExpireAt, "expiry should not be changed")
			}

			laterSo := dk.GetSellOrder(ctx, dymName.Name, dymnstypes.NameOrder)
			historicalSo := dk.GetHistoricalSellOrders(ctx, dymName.Name, dymnstypes.NameOrder)
			laterOwnerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(ownerA), params.BaseDenom)
			laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(buyerA), params.BaseDenom)
			laterPreviousBidderBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(previousBidderA), params.BaseDenom)
			laterDymNamesOwnedByOwner, err := dk.GetDymNamesOwnedBy(ctx, ownerA)
			require.NoError(t, err)
			laterDymNamesOwnedByBuyer, err := dk.GetDymNamesOwnedBy(ctx, buyerA)
			require.NoError(t, err)
			laterDymNamesOwnedByPreviousBidder, err := dk.GetDymNamesOwnedBy(ctx, previousBidderA)
			require.NoError(t, err)

			require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
			require.Equal(t, tt.wantBuyerBalanceLater, laterBuyerBalance.Amount.Int64(), "buyer balance mis-match")
			require.Equal(t, tt.wantPreviousBidderBalanceLater, laterPreviousBidderBalance.Amount.Int64(), "previous bidder balance mis-match")

			require.Empty(t, laterDymNamesOwnedByPreviousBidder, "no reverse record should be made for previous bidder")

			if tt.wantErr {
				require.Error(t, errPurchaseName, "action should be failed")
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Contains(t, errPurchaseName.Error(), tt.wantErrContains)
				require.Nil(t, resp)

				require.False(t, tt.wantOwnershipChanged, "mis-configured test case")

				require.Less(t,
					ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should not consume params gas on failed operation",
				)
			} else {
				require.NoError(t, errPurchaseName, "action should be successful")
				require.NotNil(t, resp)

				require.GreaterOrEqual(t,
					ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceBidOnSellOrder,
					"should consume params gas",
				)
			}

			if tt.wantOwnershipChanged {
				if tt.withoutDymName {
					t.Errorf("mis-configured test case")
					return
				}
				if tt.withoutSellOrder {
					t.Errorf("mis-configured test case")
					return
				}

				require.Nil(t, laterSo, "SO should be deleted")
				require.Len(t, historicalSo, 1, "SO should be moved to historical")

				require.Equal(t, buyerA, laterDymName.Owner, "ownership should be changed")
				require.Equal(t, buyerA, laterDymName.Controller, "controller should be changed")
				require.Empty(t, laterDymName.Configs, "configs should be cleared")
				require.Empty(t, laterDymNamesOwnedByOwner, "reverse record should be removed")
				require.Len(t, laterDymNamesOwnedByBuyer, 1, "reverse record should be added")
			} else {
				if tt.withoutDymName {
					require.Nil(t, laterDymName)
					require.Empty(t, laterDymNamesOwnedByOwner)
					require.Empty(t, laterDymNamesOwnedByBuyer)
				} else {
					require.Equal(t, ownerA, laterDymName.Owner, "ownership should not be changed")
					require.Equal(t, ownerA, laterDymName.Controller, "controller should not be changed")
					require.NotEmpty(t, laterDymName.Configs, "configs should be kept")
					require.Equal(t, dymName.Configs, laterDymName.Configs, "configs not be changed")
					require.Len(t, laterDymNamesOwnedByOwner, 1, "reverse record should be kept")
					require.Empty(t, laterDymNamesOwnedByBuyer, "reverse record should not be added")
				}

				if tt.withoutSellOrder {
					require.Nil(t, laterSo)
					require.Empty(t, historicalSo)
				} else {
					require.NotNil(t, laterSo, "SO should not be deleted")
					require.Empty(t, historicalSo, "SO should not be moved to historical")
				}
			}
		})
	}

	t.Run("reject purchase order when trading is disabled", func(t *testing.T) {
		// setup execution context
		dk, _, ctx := setupTest()

		moduleParams := dk.GetParams(ctx)
		moduleParams.Misc.EnableTradingName = false
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PurchaseOrder(ctx, &dymnstypes.MsgPurchaseOrder{
				GoodsId:   "my-name",
				OrderType: dymnstypes.NameOrder,
				Offer:     dymnsutils.TestCoin(100),
				Buyer:     buyerA,
			})
			return err
		}, "trading of Dym-Name is disabled")
	})
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_PurchaseOrder_Alias() {
	s.Run("reject if message not pass validate basic", func() {
		s.SetupTest()

		s.requireErrorFContains(func() error {
			_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
				OrderType: dymnstypes.AliasOrder,
			})
			return err
		}, gerrc.ErrInvalidArgument.Error())
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

	msg := func(buyer string, offer int64, goodsId, dstRollAppId string) dymnstypes.MsgPurchaseOrder {
		return dymnstypes.MsgPurchaseOrder{
			GoodsId:   goodsId,
			OrderType: dymnstypes.AliasOrder,
			Params:    []string{dstRollAppId},
			Offer:     dymnsutils.TestCoin(offer),
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
			// setup execution context
			s.SetupTest()

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
				s.Require().Equal(tt.sellOrder.GoodsId, tt.msg.GoodsId, "bad setup")

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

			historicalSo := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, tt.msg.GoodsId, dymnstypes.AliasOrder)
			s.Empty(historicalSo, "no historical SO should be made for alias order regardless state of tx")

			laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, tt.msg.GoodsId, dymnstypes.AliasOrder)

			if tt.wantErr {
				s.requireErrorContains(errPurchaseName, tt.wantErrContains)
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
				s.requireRollApp(destinationRollAppId).HasAlias(tt.msg.GoodsId)
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
		s.SetupTest()

		moduleParams := s.dymNsKeeper.GetParams(s.ctx)
		moduleParams.Misc.EnableTradingAlias = false
		err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
		s.Require().NoError(err)

		s.requireErrorFContains(func() error {
			_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PurchaseOrder(s.ctx, &dymnstypes.MsgPurchaseOrder{
				GoodsId:   "alias",
				OrderType: dymnstypes.AliasOrder,
				Params:    []string{rollApp_2_byBuyer_asDst.rollAppId},
				Offer:     dymnsutils.TestCoin(100),
				Buyer:     creator_2_asBuyer,
			})
			return err
		}, "trading of Alias is disabled")
	})
}
