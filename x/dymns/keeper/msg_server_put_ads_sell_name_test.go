package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_msgServer_PutAdsSellName(t *testing.T) {
	now := time.Now().UTC()

	const daysProhibitSell = 30
	const daysSellOrderDuration = 7

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
		moduleParams.Misc.SellOrderDuration = daysSellOrderDuration * 24 * time.Hour
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PutAdsSellName(ctx, &dymnstypes.MsgPutAdsSellName{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	const name = "bonded-pool"
	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
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
			name:            "Dym-Name does not exists",
			withoutDymName:  true,
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name:               "wrong owner",
			customOwner:        owner,
			customDymNameOwner: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
			minPrice:           coin100,
			wantErr:            true,
			wantErrContains:    sdkerrors.ErrUnauthorized.Error(),
		},
		{
			name:                    "expired Dym-Name",
			withoutDymName:          false,
			existingSo:              nil,
			dymNameExpiryOffsetDays: -1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "Dym-Name is already expired",
		},
		{
			name: "existing active SO, not finished",
			existingSo: &dymnstypes.SellOrder{
				ExpireAt:  now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active Sell-Order already exists",
		},
		{
			name: "existing active SO, expired",
			existingSo: &dymnstypes.SellOrder{
				ExpireAt:  now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "existing active SO, not expired, completed",
			existingSo: &dymnstypes.SellOrder{
				ExpireAt:  now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "existing active SO, expired, completed",
			existingSo: &dymnstypes.SellOrder{
				ExpireAt:  now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists",
		},
		{
			name:            "not allowed denom",
			minPrice:        sdk.NewInt64Coin("u"+params.BaseDenom, 100),
			wantErr:         true,
			wantErrContains: "only adym is allowed as price",
		},
		{
			name:                    "can not sell Dym-Name that almost expired",
			dymNameExpiryOffsetDays: daysProhibitSell - 1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "before Dym-Name expiry, can not sell",
		},
		{
			name:                    "successfully place ads for selling Dym-Name, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               nil,
		},
		{
			name:                    "successfully place ads for selling Dym-Name, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               dymnsutils.TestCoinP(0),
		},
		{
			name:                    "successfully place ads for selling Dym-Name, with sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin300,
		},
		{
			name:                    "successfully place ads for selling Dym-Name, with sell price equals to min-price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, ctx := setupTest()

			useDymNameOwner := owner
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
				tt.existingSo.Name = name
				err := dk.SetSellOrder(ctx, *tt.existingSo)
				require.NoError(t, err)
			}

			useOwner := owner
			if tt.customOwner != "" {
				useOwner = tt.customOwner
			}
			msg := &dymnstypes.MsgPutAdsSellName{
				Name:      name,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
				Owner:     useOwner,
			}
			resp, err := dymnskeeper.NewMsgServerImpl(dk).PutAdsSellName(ctx, msg)
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

				so := dk.GetSellOrder(ctx, name)
				if tt.existingSo != nil {
					require.NotNil(t, so)
					require.Equal(t, *tt.existingSo, *so)
				} else {
					require.Nil(t, so)
				}

				require.Less(t,
					ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPutAds,
					"should not consume params gas on failed operation",
				)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			so := dk.GetSellOrder(ctx, name)
			require.NotNil(t, so)

			expectedSo := dymnstypes.SellOrder{
				Name:       name,
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
				ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPutAds,
				"should consume params gas",
			)

			apoe := dk.GetActiveSellOrdersExpiration(ctx)

			var found bool
			for _, record := range apoe.Records {
				if record.Name == name {
					found = true
					require.Equal(t, expectedSo.ExpireAt, record.ExpireAt)
					break
				}
			}

			require.True(t, found)
		})
	}
}