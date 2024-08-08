package keeper_test

import (
	"fmt"
	"testing"
	"time"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection GoSnakeCaseUsage
func TestKeeper_CompleteAliasSellOrder(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper, dymnskeeper.BankKeeper) {
		dk, bk, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return ctx, dk, rk, bk
	}

	type rollapp struct {
		rollAppId string
		creator   string
		alias     string
	}

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()

	rollApp_1_asSrc := rollapp{
		rollAppId: "rollapp_1-1",
		creator:   creator_1_asOwner,
		alias:     "alias",
	}
	rollApp_2_asDst := rollapp{
		rollAppId: "rollapp_2-2",
		creator:   creator_2_asBuyer,
		alias:     "",
	}
	rollApp_3_asDst_byOwner := rollapp{
		rollAppId: "rollapp_3-1",
		creator:   creator_1_asOwner,
		alias:     "exists",
	}

	registerRollApp := func(ra rollapp, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
		rk.SetRollapp(ctx, rollapptypes.Rollapp{
			RollappId: ra.rollAppId,
			Owner:     ra.creator,
		})
		if ra.alias != "" {
			err := dk.SetAliasForRollAppId(ctx, ra.rollAppId, ra.alias)
			require.NoError(t, err)
		}
	}

	t.Run("alias not found", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollapp{
			rollAppId: rollApp_1_asSrc.rollAppId,
			creator:   rollApp_1_asSrc.creator,
			alias:     "", // no alias
		}, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:   "void",
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(200),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_2_asDst.creator,
				Price:  dymnsutils.TestCoin(200),
				Params: []string{rollApp_2_asDst.rollAppId},
			},
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "alias not owned by any RollApp: void: not found")
	})

	t.Run("destination Roll-App not found", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)

		const sellPrice = 200

		so := dymnstypes.SellOrder{
			GoodsId:   rollApp_1_asSrc.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(sellPrice),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_2_asDst.creator,
				Price:  dymnsutils.TestCoin(sellPrice),
				Params: []string{"nah_0-0"},
			},
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "destination Roll-App does not exists")
	})

	t.Run("SO not found", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		requireErrorContains(t,
			dk.CompleteAliasSellOrder(ctx, rollApp_1_asSrc.alias),
			fmt.Sprintf("Sell-Order: %s: not found", rollApp_1_asSrc.alias),
		)
	})

	t.Run("SO not yet completed, no bidder", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:  rollApp_1_asSrc.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "Sell-Order has not finished yet")
	})

	t.Run("SO has bidder but not yet completed", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:   rollApp_1_asSrc.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_2_asDst.creator,
				Price:  dymnsutils.TestCoin(200), // lower than sell price
				Params: []string{rollApp_2_asDst.rollAppId},
			},
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "Sell-Order has not finished yet")
	})

	t.Run("SO expired without bidder", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:   rollApp_1_asSrc.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() - 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "no bid placed")
	})

	t.Run("SO without sell price, with bid, finished by expiry", func(t *testing.T) {
		ctx, dk, rk, _ := setupTest()

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_2_asDst, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:  rollApp_1_asSrc.alias,
			Type:     dymnstypes.AliasOrder,
			ExpireAt: now.Unix() + 1,
			MinPrice: dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_2_asDst.creator,
				Price:  dymnsutils.TestCoin(200),
				Params: []string{rollApp_2_asDst.rollAppId},
			},
		}
		err := dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteAliasSellOrder(ctx, so.GoodsId), "Sell-Order has not finished yet")
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
			ctx, dk, rk, bk := setupTest()

			err := bk.MintCoins(ctx,
				dymnstypes.ModuleName,
				dymnsutils.TestCoins(ownerOriginalBalance+buyerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(creator_1_asOwner),
				dymnsutils.TestCoins(ownerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(creator_2_asBuyer),
				dymnsutils.TestCoins(buyerOriginalBalance),
			)
			require.NoError(t, err)

			registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
			registerRollApp(rollApp_2_asDst, ctx, dk, rk)

			so := dymnstypes.SellOrder{
				GoodsId:  rollApp_1_asSrc.alias,
				Type:     dymnstypes.AliasOrder,
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
					Bidder: rollApp_2_asDst.creator,
					Price:  dymnsutils.TestCoin(tt.bid),
					Params: []string{rollApp_2_asDst.rollAppId},
				}

				// mint coin to module account because we charged buyer before update SO
				err = bk.MintCoins(ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
				require.NoError(t, err)
			}
			err = dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			rollApp1, found := rk.GetRollapp(ctx, rollApp_1_asSrc.rollAppId)
			require.True(t, found)
			rollApp2, found := rk.GetRollapp(ctx, rollApp_2_asDst.rollAppId)
			require.True(t, found)

			// test

			errCompleteSellOrder := dk.CompleteAliasSellOrder(ctx, so.GoodsId)

			laterSo := dk.GetSellOrder(ctx, so.GoodsId, dymnstypes.AliasOrder)

			historicalSo := dk.GetHistoricalSellOrders(ctx, so.GoodsId, dymnstypes.AliasOrder)
			require.Empty(t, historicalSo, "historical should be empty as not supported for order type Alias")

			laterOwnerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(rollApp_1_asSrc.creator), params.BaseDenom)
			laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(rollApp_2_asDst.creator), params.BaseDenom)

			laterAliasOfRollApp1, _ := dk.GetAliasByRollAppId(ctx, rollApp_1_asSrc.rollAppId)
			laterAliasOfRollApp2, _ := dk.GetAliasByRollAppId(ctx, rollApp_2_asDst.rollAppId)

			laterAliasLinkedToRollAppId, found := dk.GetRollAppIdByAlias(ctx, rollApp_1_asSrc.alias)
			require.True(t, found)

			laterRollApp1, found := rk.GetRollapp(ctx, rollApp_1_asSrc.rollAppId)
			require.True(t, found, "rollapp should be kept")
			require.Equal(t, rollApp1, laterRollApp1, "rollapp should not be changed")
			laterRollApp2, found := rk.GetRollapp(ctx, rollApp_2_asDst.rollAppId)
			require.True(t, found, "rollapp should be kept")
			require.Equal(t, rollApp2, laterRollApp2, "rollapp should not be changed")

			if tt.wantErr {
				require.Error(t, errCompleteSellOrder, "action should be failed")
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Contains(t, errCompleteSellOrder.Error(), tt.wantErrContains)

				require.NotNil(t, laterSo, "SO should not be deleted")

				require.Equal(t, rollApp_1_asSrc.alias, laterAliasOfRollApp1, "alias should be kept")
				require.Equal(t, rollApp_1_asSrc.rollAppId, laterAliasLinkedToRollAppId, "alias should be kept")
				require.Empty(t, laterAliasOfRollApp2, "should not be linked to RollApp 2")

				require.Equal(t, ownerOriginalBalance, laterOwnerBalance.Amount.Int64(), "owner balance should not be changed")
				require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
				require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
				return
			}

			require.NoError(t, errCompleteSellOrder, "action should be successful")

			require.Nil(t, laterSo, "SO should be deleted")
			require.Empty(t, historicalSo, "historical should be empty as not supported for order type Alias")

			require.Empty(t, laterAliasOfRollApp1, "should not be linked to RollApp 1 anymore")
			require.Equal(t, rollApp_2_asDst.rollAppId, laterAliasLinkedToRollAppId, "alias should be linked to RollApp 2")

			require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
			require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
		})
	}

	t.Run("if buyer is owner, can still process", func(t *testing.T) {
		ctx, dk, rk, bk := setupTest()

		const ownerOriginalBalance = 100
		const moduleAccountOriginalBalance = 1000
		const offerValue = 300

		err := bk.MintCoins(ctx,
			dymnstypes.ModuleName,
			dymnsutils.TestCoins(ownerOriginalBalance+moduleAccountOriginalBalance),
		)
		require.NoError(t, err)
		err = bk.SendCoinsFromModuleToAccount(ctx,
			dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(creator_1_asOwner),
			dymnsutils.TestCoins(ownerOriginalBalance),
		)
		require.NoError(t, err)

		registerRollApp(rollApp_1_asSrc, ctx, dk, rk)
		registerRollApp(rollApp_3_asDst_byOwner, ctx, dk, rk)

		so := dymnstypes.SellOrder{
			GoodsId:   rollApp_1_asSrc.alias,
			Type:      dymnstypes.AliasOrder,
			ExpireAt:  now.Unix() + 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(offerValue),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: rollApp_3_asDst_byOwner.creator,
				Price:  dymnsutils.TestCoin(offerValue),
				Params: []string{rollApp_3_asDst_byOwner.rollAppId},
			},
		}

		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		err = dk.CompleteAliasSellOrder(ctx, so.GoodsId)
		require.NoError(t, err)

		// Alias should be transferred as normal
		laterAliasOfRollApp1, _ := dk.GetAliasByRollAppId(ctx, rollApp_1_asSrc.rollAppId)
		laterAliasOfRollApp3, _ := dk.GetAliasByRollAppId(ctx, rollApp_3_asDst_byOwner.rollAppId)
		laterAliasLinkedToRollAppId, _ := dk.GetRollAppIdByAlias(ctx, so.GoodsId)

		require.Empty(t, laterAliasOfRollApp1, "should not be linked to RollApp 1 anymore")
		require.Equal(t, rollApp_3_asDst_byOwner.alias, laterAliasOfRollApp3, "alias should be linked to RollApp 3")
		require.Equal(t, rollApp_3_asDst_byOwner.rollAppId, laterAliasLinkedToRollAppId, "alias should be linked to RollApp 3")

		// ensure all existing alias are linked to the correct RollApp
		for _, alias := range []string{rollApp_3_asDst_byOwner.alias, so.GoodsId} {
			requireAliasLinkedToRollApp(alias, rollApp_3_asDst_byOwner.rollAppId, t, ctx, dk)
		}

		// owner receives the offer amount because owner also the buyer
		laterOwnerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(rollApp_1_asSrc.creator), params.BaseDenom)
		require.Equal(t, int64(offerValue+ownerOriginalBalance), laterOwnerBalance.Amount.Int64())

		// SO records should be processed as normal
		laterSo := dk.GetSellOrder(ctx, so.GoodsId, dymnstypes.AliasOrder)
		require.Nil(t, laterSo, "SO should be deleted")

		historicalSo := dk.GetHistoricalSellOrders(ctx, so.GoodsId, dymnstypes.AliasOrder)
		require.Empty(t, historicalSo, "historical should be empty as not supported for order type Alias")
	})
}
