package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_PlaceSellOrder(t *testing.T) {
	now := time.Now().UTC()

	const daysProhibitSell = 30
	const daysSellOrderDuration = 7
	denom := dymnsutils.TestCoin(0).Denom

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		moduleParams.Price.PriceDenom = denom
		moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
		moduleParams.Misc.SellOrderDuration = daysSellOrderDuration * 24 * time.Hour
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PlaceSellOrder(ctx, &dymnstypes.MsgPlaceSellOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	const name = "my-name"

	ownerA := testAddr(1).bech32()
	notOwnerA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	coin100 := dymnsutils.TestCoin(100)
	coin200 := dymnsutils.TestCoin(200)
	coin300 := dymnsutils.TestCoin(300)

	tests := []struct {
		name                    string
		withoutDymName          bool
		existingSo              *dymnstypes.SellOrder
		dymNameExpiryOffsetDays int64
		customOwner             string
		customDymNameOwner      string
		minPrice                sdk.Coin
		sellPrice               *sdk.Coin
		wantErr                 bool
		wantErrContains         string
	}{
		{
			name:            "fail - Dym-Name does not exists",
			withoutDymName:  true,
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", name),
		},
		{
			name:               "fail - wrong owner",
			customOwner:        ownerA,
			customDymNameOwner: notOwnerA,
			minPrice:           coin100,
			wantErr:            true,
			wantErrContains:    "not the owner of the Dym-Name",
		},
		{
			name:                    "fail - expired Dym-Name",
			withoutDymName:          false,
			existingSo:              nil,
			dymNameExpiryOffsetDays: -1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "Dym-Name is already expired",
		},
		{
			name: "fail - existing active SO, not finished",
			existingSo: &dymnstypes.SellOrder{
				Type:      dymnstypes.NameOrder,
				ExpireAt:  now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active Sell-Order already exists",
		},
		{
			name: "fail - existing active SO, expired",
			existingSo: &dymnstypes.SellOrder{
				Type:      dymnstypes.NameOrder,
				ExpireAt:  now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, not expired, completed",
			existingSo: &dymnstypes.SellOrder{
				Type:      dymnstypes.NameOrder,
				ExpireAt:  now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, expired, completed",
			existingSo: &dymnstypes.SellOrder{
				Type:      dymnstypes.NameOrder,
				ExpireAt:  now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists",
		},
		{
			name:            "fail - not allowed denom",
			minPrice:        sdk.NewInt64Coin("u"+denom, 100),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("the only denom allowed as price: %s", denom),
		},
		{
			name:                    "fail - can not sell Dym-Name that almost expired",
			dymNameExpiryOffsetDays: daysProhibitSell - 1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "duration before Dym-Name expiry, prohibited to sell",
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               nil,
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               dymnsutils.TestCoinP(0),
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, with sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin300,
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, with sell price equals to min-price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, ctx := setupTest()

			useDymNameOwner := ownerA
			if tt.customDymNameOwner != "" {
				useDymNameOwner = tt.customDymNameOwner
			}
			useDymNameExpiry := ctx.BlockTime().Add(
				time.Hour * 24 * time.Duration(tt.dymNameExpiryOffsetDays),
			).Unix()

			if !tt.withoutDymName {
				dymName := dymnstypes.DymName{
					Name:       name,
					Owner:      useDymNameOwner,
					Controller: useDymNameOwner,
					ExpireAt:   useDymNameExpiry,
				}
				err := dk.SetDymName(ctx, dymName)
				require.NoError(t, err)
			}

			if tt.existingSo != nil {
				tt.existingSo.GoodsId = name
				err := dk.SetSellOrder(ctx, *tt.existingSo)
				require.NoError(t, err)
			}

			useOwner := ownerA
			if tt.customOwner != "" {
				useOwner = tt.customOwner
			}
			msg := &dymnstypes.MsgPlaceSellOrder{
				GoodsId:   name,
				OrderType: dymnstypes.NameOrder,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
				Owner:     useOwner,
			}
			resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceSellOrder(ctx, msg)
			moduleParams := dk.GetParams(ctx)

			defer func() {
				laterDymName := dk.GetDymName(ctx, name)
				if tt.withoutDymName {
					require.Nil(t, laterDymName)
					return
				}

				require.NotNil(t, laterDymName)
				require.Equal(t, dymnstypes.DymName{
					Name:       name,
					Owner:      useDymNameOwner,
					Controller: useDymNameOwner,
					ExpireAt:   useDymNameExpiry,
				}, *laterDymName, "Dym-Name record should not be changed in any case")
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				require.Nil(t, resp)

				so := dk.GetSellOrder(ctx, name, dymnstypes.NameOrder)
				if tt.existingSo != nil {
					require.NotNil(t, so)
					require.Equal(t, *tt.existingSo, *so)
				} else {
					require.Nil(t, so)
				}

				require.Less(t,
					ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
					"should not consume params gas on failed operation",
				)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			so := dk.GetSellOrder(ctx, name, dymnstypes.NameOrder)
			require.NotNil(t, so)

			expectedSo := dymnstypes.SellOrder{
				GoodsId:    name,
				Type:       dymnstypes.NameOrder,
				ExpireAt:   ctx.BlockTime().Add(moduleParams.Misc.SellOrderDuration).Unix(),
				MinPrice:   msg.MinPrice,
				SellPrice:  msg.SellPrice,
				HighestBid: nil,
			}
			if !expectedSo.HasSetSellPrice() {
				expectedSo.SellPrice = nil
			}

			require.Nil(t, so.HighestBid, "highest bid should not be set")

			require.Equal(t, expectedSo, *so)

			require.GreaterOrEqual(t,
				ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
				"should consume params gas",
			)

			aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.NameOrder)

			var found bool
			for _, record := range aSoe.Records {
				if record.GoodsId == name {
					found = true
					require.Equal(t, expectedSo.ExpireAt, record.ExpireAt)
					break
				}
			}

			require.True(t, found)
		})
	}
}
