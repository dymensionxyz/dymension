package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_RegisterName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const firstYearPrice1L = 6
	const firstYearPrice2L = 5
	const firstYearPrice3L = 4
	const firstYearPrice4L = 3
	const firstYearPrice5PlusL = 2
	const extendsPrice = 1
	const gracePeriod = 30

	buyerA := testAddr(1).bech32()
	previousOwnerA := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()

	const preservedDymName = "preserved"

	preservedAddr1a := testAddr(4).bech32()
	preservedAddr2a := testAddr(5).bech32()

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.Price_1Letter = sdk.NewInt(firstYearPrice1L)
		moduleParams.Price.Price_2Letters = sdk.NewInt(firstYearPrice2L)
		moduleParams.Price.Price_3Letters = sdk.NewInt(firstYearPrice3L)
		moduleParams.Price.Price_4Letters = sdk.NewInt(firstYearPrice4L)
		moduleParams.Price.Price_5PlusLetters = sdk.NewInt(firstYearPrice5PlusL)
		moduleParams.Price.PriceExtends = sdk.NewInt(extendsPrice)
		moduleParams.Price.PriceDenom = denom
		// misc
		moduleParams.Misc.GracePeriodDuration = gracePeriod * 24 * time.Hour
		// preserved
		moduleParams.PreservedRegistration.ExpirationEpoch = now.Add(time.Hour).Unix()
		moduleParams.PreservedRegistration.PreservedDymNames = []dymnstypes.PreservedDymName{
			{
				DymName:            preservedDymName,
				WhitelistedAddress: preservedAddr1a,
			},
			{
				DymName:            preservedDymName,
				WhitelistedAddress: preservedAddr2a,
			},
		}
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).RegisterName(ctx, &dymnstypes.MsgRegisterName{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	const originalModuleBalance int64 = 88

	tests := []struct {
		name                    string
		buyer                   string
		originalBalance         int64
		duration                int64
		confirmPayment          sdk.Coin
		contact                 string
		customDymName           string
		existingDymName         *dymnstypes.DymName
		setupHistoricalData     bool
		preRunSetup             func(dymnskeeper.Keeper, sdk.Context)
		wantLaterDymName        *dymnstypes.DymName
		wantErr                 bool
		wantErrContains         string
		wantLaterBalance        int64
		wantPruneHistoricalData bool
	}{
		{
			name:            "can register, new Dym-Name",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "contact@example.com",
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Contact:    "contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "not allow to takeover a non-expired Dym-Name",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Add(time.Hour).Unix(),
				Contact:    "existing@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Add(time.Hour).Unix(),
				Contact:    "existing@example.com",
			},
			wantErr:          true,
			wantErrContains:  sdkerrors.ErrUnauthorized.Error(),
			wantLaterBalance: 1,
		},
		{
			name:            "not allow to takeover an expired Dym-Name which in grace period",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			wantErr:          true,
			wantErrContains:  "can be taken over after",
			wantLaterBalance: 1,
		},
		{
			name:             "not enough balance to pay for the Dym-Name",
			buyer:            buyerA,
			originalBalance:  1,
			duration:         2,
			confirmPayment:   dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			wantErr:          true,
			wantErrContains:  "insufficient funds",
			wantLaterBalance: 1,
		},
		{
			name:             "mis-match confirm payment",
			buyer:            buyerA,
			originalBalance:  firstYearPrice5PlusL + extendsPrice + 3,
			duration:         2,
			confirmPayment:   dymnsutils.TestCoin(1),
			wantErr:          true,
			wantErrContains:  dymnstypes.ErrUnAcknowledgedPayment.Error(),
			wantLaterBalance: firstYearPrice5PlusL + extendsPrice + 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 5+ letters, multiple years",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice*2 + 3,
			duration:        3,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice*2),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*3,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 5+ letters, 1 year",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 4 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "kids",
			originalBalance: firstYearPrice4L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice4L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 4 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "kids",
			originalBalance: firstYearPrice4L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice4L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 3 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "abc",
			originalBalance: firstYearPrice3L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice3L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 3 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "abc",
			originalBalance: firstYearPrice3L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice3L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 2 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "ab",
			originalBalance: firstYearPrice2L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice2L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 2 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "ab",
			originalBalance: firstYearPrice2L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice2L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 1 letter, multiple years",
			buyer:           buyerA,
			customDymName:   "a",
			originalBalance: firstYearPrice1L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice1L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "deduct balance for new Dym-Name, 1 letter, 1 year",
			buyer:           buyerA,
			customDymName:   "a",
			originalBalance: firstYearPrice1L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice1L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "can extend owned Dym-Name, not expired",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "when extend owned non-expired Dym-Name, keep config and historical data",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: false,
		},
		{
			name:            "when extend owned non-expired Dym-Name, keep config and historical data, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "new-contact@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: false,
		},
		{
			name:            "can renew owned Dym-Name, expired",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "can renew owned Dym-Name, expired, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "when renew previously-owned expired Dym-Name, reset config",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   5,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: true,
		},
		{
			name:            "when renew previously-owned expired Dym-Name, reset config, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   5,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: true,
		},
		{
			name:            "can take over an expired Dym-Name after grace period has passed",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "take over an expired when ownership changed, reset config",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "old-contact@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "take over an expired when ownership changed, reset config, update contact if provided",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: buyerA,
				}},
				Contact: "old-contact@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "not enough balance to take over an expired Dym-Name after grace period has passed",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   3,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   3,
			},
			wantErr:          true,
			wantErrContains:  "insufficient funds",
			wantLaterBalance: 1,
		},
		{
			name:            "address in the preserved Dym-Name list, can still buy other Dym-Names",
			buyer:           preservedAddr1a,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr1a,
				Controller: preservedAddr1a,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "only whitelisted address can buy preserved Dym-Name, addr 1",
			buyer:           preservedAddr1a,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr1a,
				Controller: preservedAddr1a,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "only whitelisted address can buy preserved Dym-Name, addr 2",
			buyer:           preservedAddr2a,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr2a,
				Controller: preservedAddr2a,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
		{
			name:             "only whitelisted address can buy preserved Dym-Name, reject others",
			buyer:            buyerA,
			customDymName:    preservedDymName,
			originalBalance:  firstYearPrice5PlusL + 3,
			duration:         1,
			confirmPayment:   dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantErr:          true,
			wantErrContains:  "only able to be registered by specific addresses",
			wantLaterBalance: firstYearPrice5PlusL + 3,
		},
		{
			name:            "after preserved expiration date, anyone can buy preserved Dym-Name",
			buyer:           buyerA,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			preRunSetup: func(dk dymnskeeper.Keeper, ctx sdk.Context) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.PreservedRegistration.ExpirationEpoch = now.Add(-time.Hour).Unix()
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, bk, ctx := setupTest()

			err := bk.MintCoins(ctx, dymnstypes.ModuleName, dymnsutils.TestCoins(originalModuleBalance))
			require.NoError(t, err)

			if tt.originalBalance > 0 {
				coin := dymnsutils.TestCoins(tt.originalBalance)
				err := bk.MintCoins(ctx, dymnstypes.ModuleName, coin)
				require.NoError(t, err)
				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(tt.buyer),
					coin,
				)
				require.NoError(t, err)
			}

			useRecordName := "my-name"
			if tt.customDymName != "" {
				useRecordName = tt.customDymName
			}

			if tt.existingDymName != nil {
				tt.existingDymName.Name = useRecordName
				err := dk.SetDymName(ctx, *tt.existingDymName)
				require.NoError(t, err)

				if tt.setupHistoricalData {
					so1 := dymnstypes.SellOrder{
						Name:     useRecordName,
						ExpireAt: now.Unix() - 1,
						MinPrice: dymnsutils.TestCoin(1),
					}
					err := dk.SetSellOrder(ctx, so1)
					require.NoError(t, err)

					err = dk.MoveSellOrderToHistorical(ctx, useRecordName)
					require.NoError(t, err)

					so2 := dymnstypes.SellOrder{
						Name:      useRecordName,
						ExpireAt:  tt.existingDymName.ExpireAt - 1,
						MinPrice:  dymnsutils.TestCoin(1),
						SellPrice: dymnsutils.TestCoinP(2),
						HighestBid: &dymnstypes.SellOrderBid{
							Bidder: anotherA,
							Price:  dymnsutils.TestCoin(2),
						},
					}
					err = dk.SetSellOrder(ctx, so2)
					require.NoError(t, err)

					require.Len(t, dk.GetHistoricalSellOrders(ctx, useRecordName), 1)
				}
			} else {
				require.False(t, tt.setupHistoricalData, "bad setup testcase")
			}
			if tt.wantLaterDymName != nil {
				tt.wantLaterDymName.Name = useRecordName
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(dk, ctx)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).RegisterName(ctx, &dymnstypes.MsgRegisterName{
				Name:           useRecordName,
				Duration:       tt.duration,
				Owner:          tt.buyer,
				ConfirmPayment: tt.confirmPayment,
				Contact:        tt.contact,
			})
			laterDymName := dk.GetDymName(ctx, useRecordName)

			defer func() {
				laterBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.buyer), denom).Amount.Int64()
				require.Equal(t, tt.wantLaterBalance, laterBalance)
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				require.Nil(t, resp)

				defer func() {
					laterModuleBalance := bk.GetBalance(
						ctx,
						dymNsModuleAccAddr, denom,
					).Amount.Int64()
					require.Equal(t, originalModuleBalance, laterModuleBalance, "module account balance should not be changed")
				}()

				if tt.existingDymName != nil {
					require.Equal(t, *tt.existingDymName, *laterDymName, "should not change existing record")
					require.NotNil(t, tt.wantLaterDymName, "bad setup testcase")
					require.Equal(t, *tt.wantLaterDymName, *laterDymName)
				} else {
					require.Nil(t, laterDymName)
					require.Nil(t, tt.wantLaterDymName, "bad setup testcase")
				}

				if tt.setupHistoricalData {
					require.NotNil(t, dk.GetSellOrder(ctx, useRecordName), "sell order must be kept")
					require.Len(t, dk.GetHistoricalSellOrders(ctx, useRecordName), 1, "historical data must be kept")
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			defer func() {
				laterModuleBalance := bk.GetBalance(
					ctx,
					dymNsModuleAccAddr, denom,
				).Amount.Int64()
				require.Equal(t, originalModuleBalance, laterModuleBalance, "token should be burned")
			}()

			require.NotNil(t, laterDymName)
			require.NotNil(t, tt.wantLaterDymName, "bad setup testcase")
			require.Equal(t, *tt.wantLaterDymName, *laterDymName)

			if tt.setupHistoricalData {
				if tt.wantPruneHistoricalData {
					require.Nil(t, dk.GetSellOrder(ctx, useRecordName), "sell order must be pruned")
					require.Empty(t, dk.GetHistoricalSellOrders(ctx, useRecordName), "historical data must be pruned")

					if tt.existingDymName.Owner != laterDymName.Owner {
						ownedByPreviousOwner, err := dk.GetDymNamesOwnedBy(ctx, tt.existingDymName.Owner)
						require.NoError(t, err)
						require.Empty(t, ownedByPreviousOwner, "reverse mapping should be removed")

						mappedDymNamesByPreviousOwner, err := dk.GetDymNamesContainsConfiguredAddress(ctx, tt.existingDymName.Owner)
						require.NoError(t, err)
						require.Empty(t, mappedDymNamesByPreviousOwner, "reverse mapping should be removed")

						mappedDymNamesByPreviousOwner, err = dk.GetDymNamesContainsHexAddress(ctx,
							sdk.MustAccAddressFromBech32(tt.existingDymName.Owner).Bytes(),
						)
						require.NoError(t, err)
						require.Empty(t, mappedDymNamesByPreviousOwner, "reverse mapping should be removed")
					}
				} else {
					require.NotNil(t, dk.GetSellOrder(ctx, useRecordName), "sell order must be kept")
					require.Len(t, dk.GetHistoricalSellOrders(ctx, useRecordName), 1, "historical data must be kept")
				}
			} else {
				require.False(t, tt.wantPruneHistoricalData, "bad setup testcase")
			}

			ownedByBuyer, err := dk.GetDymNamesOwnedBy(ctx, tt.buyer)
			require.NoError(t, err)
			require.Len(t, ownedByBuyer, 1, "reverse mapping should be set")
			require.Equal(t, useRecordName, ownedByBuyer[0].Name)

			mappedDymNamesByBuyer, err := dk.GetDymNamesContainsConfiguredAddress(ctx, tt.buyer)
			require.NoError(t, err)
			require.Len(t, mappedDymNamesByBuyer, 1, "reverse mapping should be set")
			require.Equal(t, useRecordName, mappedDymNamesByBuyer[0].Name)

			mappedDymNamesByBuyer, err = dk.GetDymNamesContainsHexAddress(ctx,
				sdk.MustAccAddressFromBech32(tt.buyer).Bytes(),
			)
			require.NoError(t, err)
			require.Len(t, mappedDymNamesByBuyer, 1, "reverse mapping should be set")
			require.Equal(t, useRecordName, mappedDymNamesByBuyer[0].Name)
		})
	}
}

func TestEstimateRegisterName(t *testing.T) {
	now := time.Now()

	const denom = "atom"
	const price1L int64 = 9
	const price2L int64 = 8
	const price3L int64 = 7
	const price4L int64 = 6
	const price5PlusL int64 = 5
	const extendsPrice int64 = 4

	params := dymnstypes.DefaultParams()
	params.Price.PriceDenom = denom
	params.Price.Price_1Letter = sdk.NewInt(price1L)
	params.Price.Price_2Letters = sdk.NewInt(price2L)
	params.Price.Price_3Letters = sdk.NewInt(price3L)
	params.Price.Price_4Letters = sdk.NewInt(price4L)
	params.Price.Price_5PlusLetters = sdk.NewInt(price5PlusL)
	params.Price.PriceExtends = sdk.NewInt(extendsPrice)

	buyerA := testAddr(1).bech32()
	previousOwnerA := testAddr(2).bech32()

	tests := []struct {
		name               string
		dymName            string
		existingDymName    *dymnstypes.DymName
		newOwner           string
		duration           int64
		wantFirstYearPrice int64
		wantExtendPrice    int64
	}{
		{
			name:               "new registration, 1 letter, 1 year",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:               "new registration, 1 letter, 2 years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "new registration, 1 letter, N years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:               "new registration, 6 letters, 1 year",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    0,
		},
		{
			name:               "new registration, 6 letters, 2 years",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "new registration, 5+ letters, N years",
			dymName:            "central",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:    "extends same owner, 1 letter, 1 year",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:    "extends same owner, 1 letter, 2 years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "extends same owner, 1 letter, N years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 99,
		},
		{
			name:    "extends same owner, 6 letters, 1 year",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:    "extends same owner, 6 letters, 2 years",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "extends same owner, 5+ letters, N years",
			dymName: "central",
			existingDymName: &dymnstypes.DymName{
				Name:       "central",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 99,
		},
		{
			name:    "extends expired, same owner, 5+ letters, 2 years",
			dymName: "central",
			existingDymName: &dymnstypes.DymName{
				Name:       "central",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "take-over, 1 letter, 1 year",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:    "take-over, 1 letter, 3 years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "take-over, 6 letters, 1 year",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    0,
		},
		{
			name:    "take-over, 6 letters, 3 years",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 2 letters",
			dymName:            "aa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price2L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 3 letters",
			dymName:            "aaa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price3L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 4 letters",
			dymName:            "geek",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price4L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 5 letters",
			dymName:            "human",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dymnskeeper.EstimateRegisterName(
				params,
				tt.dymName,
				tt.existingDymName,
				tt.newOwner,
				tt.duration,
			)
			require.Equal(t, tt.wantFirstYearPrice, got.FirstYearPrice.Amount.Int64())
			require.Equal(t, tt.wantExtendPrice, got.ExtendPrice.Amount.Int64())
			require.Equal(
				t,
				tt.wantFirstYearPrice+tt.wantExtendPrice,
				got.TotalPrice.Amount.Int64(),
				"total price must be equals to sum of first year and extend price",
			)
			require.Equal(t, denom, got.FirstYearPrice.Denom)
			require.Equal(t, denom, got.ExtendPrice.Denom)
			require.Equal(t, denom, got.TotalPrice.Denom)
		})
	}
}
