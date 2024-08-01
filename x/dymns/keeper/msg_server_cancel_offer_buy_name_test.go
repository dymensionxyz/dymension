package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_CancelOfferBuyName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	buyerA := testAddr(1).bech32()
	anotherBuyerA := testAddr(2).bech32()
	ownerA := testAddr(3).bech32()

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).CancelOfferBuyName(ctx, &dymnstypes.MsgCancelOfferBuyName{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}

	offer := &dymnstypes.OfferToBuy{
		Id:         "1",
		Name:       dymName.Name,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	offerByAnother := &dymnstypes.OfferToBuy{
		Id:         "999",
		Name:       dymName.Name,
		Buyer:      anotherBuyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.OfferToBuy
		offerId                string
		buyer                  string
		originalModuleBalance  int64
		originalBuyerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.OfferToBuy
		wantLaterModuleBalance int64
		wantLaterBuyerBalance  int64
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                   "pass - can cancel offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			offerId:                offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   0,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseOffer,
		},
		{
			name:                   "pass - cancel offer will refund the buyer",
			existingDymName:        dymName,
			existingOffer:          offer,
			offerId:                offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  1 + offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   2,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2 + offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseOffer,
		},
		{
			name:                  "pass - cancel offer will remove the offer record",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.NotNil(t, dk.GetOfferToBuy(ctx, offer.Id))
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Nil(t, dk.GetOfferToBuy(ctx, offer.Id))
			},
		},
		{
			name:                  "pass - cancel offer will remove reverse mapping records",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				offerIds, err := dk.GetOfferToBuyByBuyer(ctx, offer.Buyer)
				require.NoError(t, err)
				require.Len(t, offerIds, 1)

				offerIds, err = dk.GetOffersToBuyOfDymName(ctx, offer.Name)
				require.NoError(t, err)
				require.Len(t, offerIds, 1)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				offerIds, err := dk.GetOfferToBuyByBuyer(ctx, offer.Buyer)
				require.NoError(t, err)
				require.Empty(t, offerIds)

				offerIds, err = dk.GetOffersToBuyOfDymName(ctx, offer.Name)
				require.NoError(t, err)
				require.Empty(t, offerIds)
			},
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          nil,
			offerId:                "2142142",
			buyer:                  buyerA,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        dymnstypes.ErrOfferToBuyNotFound.Error(),
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			offerId:                "2142142",
			buyer:                  offer.Buyer,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        dymnstypes.ErrOfferToBuyNotFound.Error(),
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel offer with different buyer",
			existingDymName:        dymName,
			existingOffer:          offerByAnother,
			offerId:                "999",
			buyer:                  buyerA,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "not the owner of the offer",
			wantLaterOffer:         offerByAnother,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not cancel if module account does not have enough balance to refund",
			existingDymName:        dymName,
			existingOffer:          offer,
			offerId:                offer.Id,
			buyer:                  buyerA,
			originalModuleBalance:  0,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "insufficient funds",
			wantLaterOffer:         offer,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, bk, ctx := setupTest()

			if tt.originalModuleBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalModuleBalance),
				)
				require.NoError(t, err)
			}

			if tt.originalBuyerBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalBuyerBalance),
				)
				require.NoError(t, err)

				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(tt.buyer),
					dymnsutils.TestCoins(tt.originalBuyerBalance),
				)
				require.NoError(t, err)
			}

			if tt.existingDymName != nil {
				err := dk.SetDymName(ctx, *tt.existingDymName)
				require.NoError(t, err)
			}

			if tt.existingOffer != nil {
				err := dk.SetOfferToBuy(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingDymNameToOfferToBuy(ctx, tt.existingOffer.Name, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).CancelOfferBuyName(ctx, &dymnstypes.MsgCancelOfferBuyName{
				OfferId: tt.offerId,
				Buyer:   tt.buyer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetOfferToBuy(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetOfferToBuy(ctx, tt.offerId)
					require.Nil(t, laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.buyer), denom)
				require.Equal(t, tt.wantLaterBuyerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				if tt.existingDymName != nil {
					originalDymName := *tt.existingDymName
					laterDymName := dk.GetDymName(ctx, originalDymName.Name)
					require.NotNil(t, laterDymName)
					require.Equal(t, originalDymName, *laterDymName, "Dym-Name record should not be changed")
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(ctx, dk)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
