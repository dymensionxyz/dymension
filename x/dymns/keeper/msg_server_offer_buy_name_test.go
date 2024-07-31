package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_msgServer_OfferBuyName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	const buyer = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const anotherBuyer = "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"
	const owner = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	dymNsModuleAccAddr := authtypes.NewModuleAddress(dymnstypes.ModuleName)
	const name = "bonded-pool"

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockHeader(tmproto.Header{
			Time: now,
		})

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
			_, err := dymnskeeper.NewMsgServerImpl(dk).OfferBuyName(ctx, &dymnstypes.MsgOfferBuyName{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       name,
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
	}

	tests := []struct {
		name                        string
		existingDymName             *dymnstypes.DymName
		existingOffer               *dymnstypes.OfferToBuy
		dymName                     string
		buyer                       string
		offer                       sdk.Coin
		existingOfferId             string
		originalModuleBalance       int64
		originalBuyerBalance        int64
		originalAnotherBuyerBalance int64
		preRunSetupFunc             func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                     bool
		wantErrContains             string
		wantOfferId                 string
		wantLaterOffer              *dymnstypes.OfferToBuy
		wantLaterModuleBalance      int64
		wantLaterBuyerBalance       int64
		wantMinConsumeGas           sdk.Gas
		afterTestFunc               func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                  "pass - can place offer",
			existingDymName:       dymName,
			existingOffer:         nil,
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingOfferId:       "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			wantErr:               false,
			wantOfferId:           "1",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutOffer,
		},
		{
			name:            "pass - can extends offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "2",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			wantErr:               false,
			wantOfferId:           "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "2",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 0 + 1,
			wantLaterBuyerBalance:  1 - 1,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
		},
		{
			name:            "pass - can extends offer with counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "2",
			originalModuleBalance: 1,
			originalBuyerBalance:  2,
			wantErr:               false,
			wantOfferId:           "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 1),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 1,
			wantLaterBuyerBalance:  2 - 1,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
		},
		{
			name:            "pass - can extends offer with offer equals to counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 3),
			existingOfferId:       "2",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantOfferId:           "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 3),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 3,
			wantLaterBuyerBalance:  5 - 3,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
		},
		{
			name:            "pass - can extends offer with offer greater than counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 4),
			existingOfferId:       "2",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantOfferId:           "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 4),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 4,
			wantLaterBuyerBalance:  5 - 4,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
		},
		{
			name:            "pass - extends an existing offer only take the extra amount instead of all",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2),
			existingOfferId:       "1",
			originalModuleBalance: 5,
			originalBuyerBalance:  3,
			wantErr:               false,
			wantOfferId:           "1",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 5 + 2,
			wantLaterBuyerBalance:  3 - 2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
		},
		{
			name:                   "fail - reject offer for non-existing Dym-Name",
			existingDymName:        nil,
			dymName:                name,
			buyer:                  buyer,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        dymnstypes.ErrDymNameNotFound.Error(),
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllOffersToBuy(ctx))
			},
		},
		{
			name: "fail - reject offer for expired Dym-Name",
			existingDymName: &dymnstypes.DymName{
				Name:       name,
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() - 1,
			},
			dymName:                name,
			buyer:                  buyer,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        dymnstypes.ErrDymNameNotFound.Error(),
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllOffersToBuy(ctx))
			},
		},
		{
			name:                   "fail - can not offer own Dym-Name",
			existingDymName:        dymName,
			dymName:                name,
			buyer:                  dymName.Owner,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "cannot buy own Dym-Name",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllOffersToBuy(ctx))
			},
		},
		{
			name:            "fail - offer denom must match params",
			existingDymName: dymName,
			dymName:         name,
			buyer:           buyer,
			offer: sdk.Coin{
				Denom:  "u" + denom,
				Amount: sdk.NewInt(minOfferPrice),
			},
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "invalid offer denomination, only accept",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllOffersToBuy(ctx))
			},
		},
		{
			name:                   "fail - offer price can not lower than min defined in params",
			existingDymName:        dymName,
			dymName:                name,
			buyer:                  buyer,
			offer:                  dymnsutils.TestCoin(minOfferPrice - 1),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "offer price must be greater than or equal to",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllOffersToBuy(ctx))
			},
		},
		{
			name:            "pass - if not continue offer, create another and charges full offer price",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "",
			originalModuleBalance: minOfferPrice,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:     false,
			wantOfferId: "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "2",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: minOfferPrice + (minOfferPrice + 1),
			wantLaterBuyerBalance:  (minOfferPrice + 2) - (minOfferPrice + 1),
			wantMinConsumeGas:      dymnstypes.OpGasPutOffer,
		},
		{
			name:            "reject - continue a non-existing offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "2",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: dymnstypes.ErrOfferToBuyNotFound.Error(),
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:            "reject - continue an existing offer but not yours",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  anotherBuyer, // not the buyer
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer, // not the existing offer's buyer
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "1",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: sdkerrors.ErrUnauthorized.Error(),
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      anotherBuyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:            "reject - continue an existing offer but the Dym-Name mismatch",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   "another-name",
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "1",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Dym-Name mismatch with existing offer",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       "another-name",
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:            "reject - continue an existing offer but mis-match offer denom",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:    "1",
				Name:  name,
				Buyer: buyer,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "1",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer denomination mismatch with existing offer",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:    "1",
				Name:  name,
				Buyer: buyer,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:            "reject - continue an existing offer but new offer less than previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:       "1",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:            "reject - continue an existing offer but new offer equals to previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "1",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			dymName:               name,
			buyer:                 buyer,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2),
			existingOfferId:       "1",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllOffersToBuy(ctx), 1)
				require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
			},
		},
		{
			name:                        "pass - reverse record added after successful offer",
			existingDymName:             dymName,
			existingOffer:               nil,
			dymName:                     name,
			buyer:                       buyer,
			offer:                       dymnsutils.TestCoin(minOfferPrice),
			existingOfferId:             "",
			originalModuleBalance:       5,
			originalBuyerBalance:        minOfferPrice + 2,
			originalAnotherBuyerBalance: minOfferPrice,
			wantErr:                     false,
			wantOfferId:                 "1",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(name)
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"1"}, offerIds.OfferIds)

				offers, err := dk.GetOffersToBuyOfDymName(ctx, name)
				require.NoError(t, err)
				require.Equal(t, "1", offers[0].Id)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"1"}, offerIds.OfferIds)

				offers, err = dk.GetOfferToBuyByBuyer(ctx, buyer)
				require.NoError(t, err)
				require.Equal(t, "1", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).OfferBuyName(ctx, &dymnstypes.MsgOfferBuyName{
					Name:  name,
					Buyer: anotherBuyer,
					Offer: dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "2", resp.OfferId)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"2"}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"1"}, offerIds.OfferIds)

				key = dymnstypes.DymNameToOfferIdsRvlKey(name)
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"1", "2"}, offerIds.OfferIds)
			},
		},
		{
			name:            "pass - reverse record added after successful offer extends",
			existingDymName: dymName,
			existingOffer: &dymnstypes.OfferToBuy{
				Id:                     "2",
				Name:                   name,
				Buyer:                  buyer,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:                     name,
			buyer:                       buyer,
			offer:                       dymnsutils.TestCoin(minOfferPrice + 1),
			existingOfferId:             "2",
			originalModuleBalance:       0,
			originalBuyerBalance:        1,
			originalAnotherBuyerBalance: minOfferPrice,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountOfferToBuy(ctx, 2)

				err := dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer, "2")
				require.NoError(t, err)

				err = dk.AddReverseMappingDymNameToOfferToBuy(ctx, name, "2")
				require.NoError(t, err)
			},
			wantErr:     false,
			wantOfferId: "2",
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         "2",
				Name:       name,
				Buyer:      buyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 0 + 1,
			wantLaterBuyerBalance:  1 - 1,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(name)
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"2"}, offerIds.OfferIds)

				offers, err := dk.GetOffersToBuyOfDymName(ctx, name)
				require.NoError(t, err)
				require.Equal(t, "2", offers[0].Id)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"2"}, offerIds.OfferIds)

				offers, err = dk.GetOfferToBuyByBuyer(ctx, buyer)
				require.NoError(t, err)
				require.Equal(t, "2", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).OfferBuyName(ctx, &dymnstypes.MsgOfferBuyName{
					Name:  name,
					Buyer: anotherBuyer,
					Offer: dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "3", resp.OfferId)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"3"}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"2"}, offerIds.OfferIds)

				key = dymnstypes.DymNameToOfferIdsRvlKey(name)
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{"2", "3"}, offerIds.OfferIds)
			},
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

			if tt.originalAnotherBuyerBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalAnotherBuyerBalance),
				)
				require.NoError(t, err)

				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(anotherBuyer),
					dymnsutils.TestCoins(tt.originalAnotherBuyerBalance),
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
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).OfferBuyName(ctx, &dymnstypes.MsgOfferBuyName{
				Name:            tt.dymName,
				Buyer:           tt.buyer,
				ContinueOfferId: tt.existingOfferId,
				Offer:           tt.offer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetOfferToBuy(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.buyer), denom)
				require.Equal(t, tt.wantLaterBuyerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				if tt.existingDymName != nil {
					originalDymName := *tt.existingDymName
					laterDymName := dk.GetDymName(ctx, tt.dymName)
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
			require.Equal(t, tt.wantOfferId, resp.OfferId)
		})
	}
}
