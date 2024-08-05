package keeper_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestKeeper_CompleteDymNameSellOrder(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, bk, ctx
	}

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	const contactEmail = "contact@example.com"

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
		Contact:    contactEmail,
	}

	t.Run("Dym-Name not found", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorContains(t, dk.CompleteDymNameSellOrder(ctx, "non-exists"), "Dym-Name: non-exists: not found")
	})

	t.Run("SO not found", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		requireErrorContains(t,
			dk.CompleteDymNameSellOrder(ctx, dymName.Name),
			fmt.Sprintf("Sell-Order: %s: not found", dymName.Name),
		)
	})

	t.Run("SO not yet completed, no bidder", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			GoodsId:  dymName.Name,
			Type:     dymnstypes.MarketOrderType_MOT_DYM_NAME,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteDymNameSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	t.Run("SO has bidder but not yet completed", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			GoodsId:   dymName.Name,
			Type:      dymnstypes.MarketOrderType_MOT_DYM_NAME,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyerA,
				Price:  dymnsutils.TestCoin(200), // lower than sell price
			},
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteDymNameSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	t.Run("SO expired without bidder", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			GoodsId:   dymName.Name,
			Type:      dymnstypes.MarketOrderType_MOT_DYM_NAME,
			ExpireAt:  now.Unix() - 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteDymNameSellOrder(ctx, dymName.Name), "no bid placed")
	})

	t.Run("SO without sell price, with bid, finished by expiry", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			GoodsId:  dymName.Name,
			Type:     dymnstypes.MarketOrderType_MOT_DYM_NAME,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyerA,
				Price:  dymnsutils.TestCoin(200),
			},
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteDymNameSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
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
		t.Run(tt.name, func(t *testing.T) {
			// setup execution context
			dk, bk, ctx := setupTest()

			err := bk.MintCoins(ctx,
				dymnstypes.ModuleName,
				dymnsutils.TestCoins(ownerOriginalBalance+buyerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(ownerA),
				dymnsutils.TestCoins(ownerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(buyerA),
				dymnsutils.TestCoins(buyerOriginalBalance),
			)
			require.NoError(t, err)

			dymName.Configs = []dymnstypes.DymNameConfig{
				{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				},
			}
			setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

			so := dymnstypes.SellOrder{
				GoodsId:  dymName.Name,
				Type:     dymnstypes.MarketOrderType_MOT_DYM_NAME,
				MinPrice: dymnsutils.TestCoin(100),
			}

			if tt.expiredSO {
				so.ExpireAt = now.Unix() - 1
			} else {
				so.ExpireAt = now.Unix() + 1
			}

			require.GreaterOrEqual(t, tt.sellPrice, int64(0), "bad setup")
			so.SellPrice = dymnsutils.TestCoinP(tt.sellPrice)

			require.GreaterOrEqual(t, tt.bid, int64(0), "bad setup")
			if tt.bid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: buyerA,
					Price:  dymnsutils.TestCoin(tt.bid),
				}

				// mint coin to module account because we charged buyer before update SO
				err = bk.MintCoins(ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
				require.NoError(t, err)
			}
			err = dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			// test

			errCompleteSellOrder := dk.CompleteDymNameSellOrder(ctx, dymName.Name)
			laterDymName := dk.GetDymName(ctx, dymName.Name)
			require.NotNil(t, laterDymName)
			laterSo := dk.GetSellOrder(ctx, dymName.Name, dymnstypes.MarketOrderType_MOT_DYM_NAME)
			historicalSo := dk.GetHistoricalSellOrders(ctx, dymName.Name, dymnstypes.MarketOrderType_MOT_DYM_NAME)
			laterOwnerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(ownerA), params.BaseDenom)
			laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(buyerA), params.BaseDenom)
			laterDymNamesOwnedByOwner, err := dk.GetDymNamesOwnedBy(ctx, ownerA)
			require.NoError(t, err)
			laterDymNamesOwnedByBuyer, err := dk.GetDymNamesOwnedBy(ctx, buyerA)
			require.NoError(t, err)
			laterConfiguredAddressOwnerDymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, ownerA)
			require.NoError(t, err)
			laterConfiguredAddressBuyerDymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, buyerA)
			require.NoError(t, err)
			laterFallbackAddressOwnerDymNames, err := dk.GetDymNamesContainsFallbackAddress(ctx, sdk.MustAccAddressFromBech32(ownerA).Bytes())
			require.NoError(t, err)
			laterFallbackAddressBuyerDymNames, err := dk.GetDymNamesContainsFallbackAddress(ctx, sdk.MustAccAddressFromBech32(buyerA).Bytes())
			require.NoError(t, err)

			require.Equal(t, dymName.Name, laterDymName.Name, "name should not be changed")
			require.Equal(t, dymName.ExpireAt, laterDymName.ExpireAt, "expiry should not be changed")

			if tt.wantErr {
				require.Error(t, errCompleteSellOrder, "action should be failed")
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Contains(t, errCompleteSellOrder.Error(), tt.wantErrContains)

				require.NotNil(t, laterSo, "SO should not be deleted")
				require.Empty(t, historicalSo, "SO should not be moved to historical")

				require.Equal(t, ownerA, laterDymName.Owner, "ownership should not be changed")
				require.Equal(t, ownerA, laterDymName.Controller, "controller should not be changed")
				require.NotEmpty(t, laterDymName.Configs, "configs should be kept")
				require.Equal(t, dymName.Configs, laterDymName.Configs, "configs not be changed")
				require.Equal(t, contactEmail, dymName.Contact, "contact should not be changed")
				require.Len(t, laterDymNamesOwnedByOwner, 1, "reverse record should be kept")
				require.Empty(t, laterDymNamesOwnedByBuyer, "reverse record should not be added")
				require.Len(t, laterConfiguredAddressOwnerDymNames, 1, "reverse record should be kept")
				require.Empty(t, laterConfiguredAddressBuyerDymNames, "reverse record should not be added")
				require.Len(t, laterFallbackAddressOwnerDymNames, 1, "reverse record should be kept")
				require.Empty(t, laterFallbackAddressBuyerDymNames, "reverse record should not be added")

				require.Equal(t, ownerOriginalBalance, laterOwnerBalance.Amount.Int64(), "owner balance should not be changed")
				require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
				require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
				return
			}

			require.NoError(t, errCompleteSellOrder, "action should be successful")

			require.Nil(t, laterSo, "SO should be deleted")
			require.Len(t, historicalSo, 1, "SO should be moved to historical")

			require.Equal(t, buyerA, laterDymName.Owner, "ownership should be changed")
			require.Equal(t, buyerA, laterDymName.Controller, "controller should be changed")
			require.Empty(t, laterDymName.Configs, "configs should be cleared")
			require.Empty(t, laterDymName.Contact, "contact should be cleared")
			require.Empty(t, laterDymNamesOwnedByOwner, "reverse record should be removed")
			require.Len(t, laterDymNamesOwnedByBuyer, 1, "reverse record should be added")
			require.Empty(t, laterConfiguredAddressOwnerDymNames, "reverse record should be removed")
			require.Len(t, laterConfiguredAddressBuyerDymNames, 1, "reverse record should be added")
			require.Empty(t, laterFallbackAddressOwnerDymNames, "reverse record should be removed")
			require.Len(t, laterFallbackAddressBuyerDymNames, 1, "reverse record should be added")

			require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
			require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
		})
	}
}

func TestKeeper_GetMinExpiryOfAllHistoricalDymNameSellOrders(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	dk.SetMinExpiryHistoricalSellOrder(ctx, "one", dymnstypes.MarketOrderType_MOT_DYM_NAME, 1)
	dk.SetMinExpiryHistoricalSellOrder(ctx, "two", dymnstypes.MarketOrderType_MOT_DYM_NAME, 22)
	dk.SetMinExpiryHistoricalSellOrder(ctx, "three", dymnstypes.MarketOrderType_MOT_DYM_NAME, 333)

	records := dk.GetMinExpiryOfAllHistoricalDymNameSellOrders(ctx)
	require.Len(t, records, 3)
	require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
		{
			DymName:   "one",
			MinExpiry: 1,
		},
		{
			DymName:   "three",
			MinExpiry: 333,
		},
		{
			DymName:   "two",
			MinExpiry: 22,
		},
	}, records)

	dk.SetMinExpiryHistoricalSellOrder(ctx, "three", dymnstypes.MarketOrderType_MOT_DYM_NAME, 0)
	records = dk.GetMinExpiryOfAllHistoricalDymNameSellOrders(ctx)
	require.Len(t, records, 2)
	require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
		{
			DymName:   "one",
			MinExpiry: 1,
		},
		{
			DymName:   "two",
			MinExpiry: 22,
		},
	}, records)

	t.Run("result must be sorted by Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "a", dymnstypes.MarketOrderType_MOT_DYM_NAME, 1)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "c", dymnstypes.MarketOrderType_MOT_DYM_NAME, 2)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "b", dymnstypes.MarketOrderType_MOT_DYM_NAME, 3)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "d", dymnstypes.MarketOrderType_MOT_DYM_NAME, 4)

		records := dk.GetMinExpiryOfAllHistoricalDymNameSellOrders(ctx)
		require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
			{
				DymName:   "a",
				MinExpiry: 1,
			},
			{
				DymName:   "b",
				MinExpiry: 3,
			},
			{
				DymName:   "c",
				MinExpiry: 2,
			},
			{
				DymName:   "d",
				MinExpiry: 4,
			},
		}, records)
	})
}
