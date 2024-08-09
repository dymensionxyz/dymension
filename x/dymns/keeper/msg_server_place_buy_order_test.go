package keeper_test

import (
	"testing"
	"time"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_PlaceBuyOrder_DymName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	anotherBuyerA := testAddr(3).bech32()

	setupTest := func() (dymnskeeper.Keeper, dymnstypes.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Add(9 * 365 * 24 * time.Hour).Unix(),
	}

	tests := []struct {
		name                        string
		existingDymName             *dymnstypes.DymName
		existingOffer               *dymnstypes.BuyOrder
		dymName                     string
		buyer                       string
		offer                       sdk.Coin
		existingBuyOrderId          string
		originalModuleBalance       int64
		originalBuyerBalance        int64
		originalAnotherBuyerBalance int64
		preRunSetupFunc             func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                     bool
		wantErrContains             string
		wantBuyOrderId              string
		wantLaterOffer              *dymnstypes.BuyOrder
		wantLaterModuleBalance      int64
		wantLaterBuyerBalance       int64
		wantMinConsumeGas           sdk.Gas
		afterTestFunc               func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                  "pass - can place offer",
			existingDymName:       dymName,
			existingOffer:         nil,
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			wantErr:               false,
			wantBuyOrderId:        "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:            "pass - can extends offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "102",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  0,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "fail - can NOT extends offer of type mis-match",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{"rollapp_1-1"},
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "202",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			wantErr:               true,
			wantErrContains:       "asset type mismatch with existing offer",
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{"rollapp_1-1"},
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  1,
			wantMinConsumeGas:      1,
		},
		{
			name:            "pass - can extends offer with counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "102",
			originalModuleBalance: 1,
			originalBuyerBalance:  2,
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 1),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 1,
			wantLaterBuyerBalance:  2 - 1,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - can extends offer with offer equals to counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 3),
			existingBuyOrderId:    "102",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 3),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 3,
			wantLaterBuyerBalance:  5 - 3,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - can extends offer with offer greater than counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 4),
			existingBuyOrderId:    "102",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 4),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 4,
			wantLaterBuyerBalance:  5 - 4,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - extends an existing offer only take the extra amount instead of all",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2),
			existingBuyOrderId:    "101",
			originalModuleBalance: 5,
			originalBuyerBalance:  3,
			wantErr:               false,
			wantBuyOrderId:        "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 5 + 2,
			wantLaterBuyerBalance:  3 - 2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "fail - can NOT place offer if trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         nil,
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingName = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Dym-Name is disabled",
			wantBuyOrderId:         "",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 5,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - reject offer for non-existing Dym-Name",
			existingDymName:        nil,
			dymName:                "non-exists",
			buyer:                  buyerA,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "Dym-Name: non-exists: not found",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name: "fail - reject offer for expired Dym-Name",
			existingDymName: &dymnstypes.DymName{
				Name:       "expired",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() - 1,
			},
			dymName:                "expired",
			buyer:                  buyerA,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "Dym-Name: expired: not found",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:                   "fail - can not offer own Dym-Name",
			existingDymName:        dymName,
			dymName:                dymName.Name,
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
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name: "fail - reject Dym-Name that almost expired",
			existingDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      dymName.Owner,
				Controller: dymName.Owner,
				ExpireAt:   now.Add(1 * time.Second).Unix(),
			},
			dymName:                dymName.Name,
			buyer:                  buyerA,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "duration before Dym-Name expiry, prohibited to trade",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:            "fail - offer denom must match params",
			existingDymName: dymName,
			dymName:         dymName.Name,
			buyer:           buyerA,
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
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:                   "fail - offer price can not lower than min defined in params",
			existingDymName:        dymName,
			dymName:                dymName.Name,
			buyer:                  buyerA,
			offer:                  dymnsutils.TestCoin(minOfferPrice - 1),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "offer price must be greater than or equal to",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:            "pass - if NOT continue offer, create another and charges full offer price",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "",
			originalModuleBalance: minOfferPrice,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:        false,
			wantBuyOrderId: "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: minOfferPrice + (minOfferPrice + 1),
			wantLaterBuyerBalance:  (minOfferPrice + 2) - (minOfferPrice + 1),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:            "fail - continue a non-existing offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "102",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Buy-Order ID: 102: not found",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but not yours",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  anotherBuyerA, // not the buyer
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA, // not the existing offer's buyer
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "101",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "not the owner of the offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      anotherBuyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but the Dym-Name mismatch",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                "another-name",
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "101",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Dym-Name mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    "another-name",
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but mis-match offer denom",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:        "101",
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				Buyer:     buyerA,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "101",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer denomination mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:        "101",
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				Buyer:     buyerA,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but new offer less than previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1), // less
			existingBuyOrderId:    "101",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2), // keep
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but new offer equals to previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2), // same
			existingBuyOrderId:    "101",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "102"))
			},
		},
		{
			name:                        "pass - reverse record added after successful offer",
			existingDymName:             dymName,
			existingOffer:               nil,
			dymName:                     dymName.Name,
			buyer:                       buyerA,
			offer:                       dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:          "",
			originalModuleBalance:       5,
			originalBuyerBalance:        minOfferPrice + 2,
			originalAnotherBuyerBalance: minOfferPrice,
			wantErr:                     false,
			wantBuyOrderId:              "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"101"}, orderIds.OrderIds)

				offers, err := dk.GetBuyOrdersOfDymName(ctx, dymName.Name)
				require.NoError(t, err)
				require.Equal(t, "101", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"101"}, orderIds.OrderIds)

				offers, err = dk.GetBuyOrdersByBuyer(ctx, buyerA)
				require.NoError(t, err)
				require.Equal(t, "101", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     anotherBuyerA,
					Offer:     dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "102", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"102"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"101"}, orderIds.OrderIds)

				key = dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"101", "102"}, orderIds.OrderIds)
			},
		},
		{
			name:            "pass - reverse record added after successful offer extends",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			dymName:                     dymName.Name,
			buyer:                       buyerA,
			offer:                       dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:          "102",
			originalModuleBalance:       0,
			originalBuyerBalance:        1,
			originalAnotherBuyerBalance: minOfferPrice,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 2)

				err := dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyerA, "102")
				require.NoError(t, err)

				err = dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName.Name, dymnstypes.TypeName, "102")
				require.NoError(t, err)
			},
			wantErr:        false,
			wantBuyOrderId: "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  0,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"102"}, orderIds.OrderIds)

				offers, err := dk.GetBuyOrdersOfDymName(ctx, dymName.Name)
				require.NoError(t, err)
				require.Equal(t, "102", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"102"}, orderIds.OrderIds)

				offers, err = dk.GetBuyOrdersByBuyer(ctx, buyerA)
				require.NoError(t, err)
				require.Equal(t, "102", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     anotherBuyerA,
					Offer:     dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "103", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"103"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"102"}, orderIds.OrderIds)

				key = dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"102", "103"}, orderIds.OrderIds)
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
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(anotherBuyerA),
					dymnsutils.TestCoins(tt.originalAnotherBuyerBalance),
				)
				require.NoError(t, err)
			}

			if tt.existingDymName != nil {
				err := dk.SetDymName(ctx, *tt.existingDymName)
				require.NoError(t, err)
			}

			if tt.existingOffer != nil {
				err := dk.SetBuyOrder(ctx, *tt.existingOffer)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
				AssetId:         tt.dymName,
				AssetType:       dymnstypes.TypeName,
				Buyer:           tt.buyer,
				ContinueOrderId: tt.existingBuyOrderId,
				Offer:           tt.offer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOrder(ctx, tt.wantLaterOffer.Id)
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
			require.Equal(t, tt.wantBuyOrderId, resp.OrderId)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func Test_msgServer_PlaceBuyOrder_Alias(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	creator_3_asBuyer := testAddr(3).bech32()

	type rollapp struct {
		rollAppID string
		creator   string
		alias     string
	}

	rollApp_1_by1_asSrc := rollapp{
		rollAppID: "rollapp_1-1",
		creator:   creator_1_asOwner,
		alias:     "one",
	}
	rollApp_2_by2_asDest := rollapp{
		rollAppID: "rollapp_2-1",
		creator:   creator_2_asBuyer,
		alias:     "two",
	}
	rollApp_3_by3_asDest_noAlias := rollapp{
		rollAppID: "rollapp_3-1",
		creator:   creator_3_asBuyer,
		alias:     "",
	}
	rollApp_4_by1_asDest_noAlias := rollapp{
		rollAppID: "rollapp_4-2",
		creator:   creator_1_asOwner,
		alias:     "",
	}

	setupTest := func() (sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper, dymnstypes.BankKeeper) {
		dk, bk, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return ctx, dk, rk, bk
	}

	tests := []struct {
		name                        string
		existingRollApps            []rollapp
		existingOffer               *dymnstypes.BuyOrder
		alias                       string
		buyer                       string
		dstRollAppId                string // destination RollApp ID
		offer                       sdk.Coin
		existingBuyOrderId          string
		originalModuleBalance       int64
		originalBuyerBalance        int64
		originalAnotherBuyerBalance int64
		preRunSetupFunc             func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                     bool
		wantErrContains             string
		wantBuyOrderId              string
		wantLaterOffer              *dymnstypes.BuyOrder
		wantLaterModuleBalance      int64
		wantLaterBuyerBalance       int64
		wantMinConsumeGas           sdk.Gas
		afterTestFunc               func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                  "pass - can place offer",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:                  "fail - can not place offer of alias which presents in params",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{rollApp_1_by1_asSrc.alias},
					},
				}
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                true,
			wantErrContains:        "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 5,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - can extends offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "202",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  0,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "fail - can NOT extend offer of alias which presents in params",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "202",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{rollApp_1_by1_asSrc.alias},
					},
				}
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantBuyOrderId:  "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  1,
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - can NOT extends offer of type mis-match",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeName,
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "102",
			originalModuleBalance: 0,
			originalBuyerBalance:  1,
			wantErr:               true,
			wantErrContains:       "asset type mismatch with existing offer",
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeName,
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  1,
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - can extends offer with counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "202",
			originalModuleBalance: 1,
			originalBuyerBalance:  2,
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 1),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 1,
			wantLaterBuyerBalance:  2 - 1,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - can extends offer with offer equals to counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 3),
			existingBuyOrderId:    "202",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 3),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 3,
			wantLaterBuyerBalance:  5 - 3,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - can extends offer with offer greater than counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 4),
			existingBuyOrderId:    "202",
			originalModuleBalance: 1,
			originalBuyerBalance:  5,
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 4),
				CounterpartyOfferPrice: dymnsutils.TestCoinP(minOfferPrice + 3),
			},
			wantLaterModuleBalance: 1 + 4,
			wantLaterBuyerBalance:  5 - 4,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - extends an existing offer only take the extra amount instead of all",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2),
			existingBuyOrderId:    "201",
			originalModuleBalance: 5,
			originalBuyerBalance:  3,
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 5 + 2,
			wantLaterBuyerBalance:  3 - 2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "fail - can NOT place offer if trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingAlias = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Alias is disabled",
			wantBuyOrderId:         "",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 5,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - reject offer for non-existing Alias",
			existingRollApps:       []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:                  "void",
			buyer:                  rollApp_2_by2_asDest.creator,
			dstRollAppId:           rollApp_2_by2_asDest.rollAppID,
			offer:                  dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "alias is not in-used: void: not found",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:                  "pass - can place offer buy own alias, different RollApp",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_4_by1_asDest_noAlias},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_1_by1_asSrc.creator,
			dstRollAppId:          rollApp_4_by1_asDest_noAlias.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:    "",
			originalModuleBalance: 5,
			originalBuyerBalance:  minOfferPrice + 2,
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_4_by1_asDest_noAlias.rollAppID},
				Buyer:      rollApp_1_by1_asSrc.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:             "fail - offer denom must match params",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:            rollApp_1_by1_asSrc.alias,
			buyer:            rollApp_2_by2_asDest.creator,
			dstRollAppId:     rollApp_2_by2_asDest.rollAppID,
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
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:                   "fail - offer price can not lower than min defined in params",
			existingRollApps:       []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:                  rollApp_1_by1_asSrc.alias,
			buyer:                  rollApp_2_by2_asDest.creator,
			dstRollAppId:           rollApp_2_by2_asDest.rollAppID,
			offer:                  dymnsutils.TestCoin(minOfferPrice - 1),
			originalModuleBalance:  1,
			originalBuyerBalance:   minOfferPrice,
			wantErr:                true,
			wantErrContains:        "offer price must be greater than or equal to",
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
			},
		},
		{
			name:             "pass - if NOT continue offer, create another and charges full offer price",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "",
			originalModuleBalance: minOfferPrice,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:        false,
			wantBuyOrderId: "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: minOfferPrice + (minOfferPrice + 1),
			wantLaterBuyerBalance:  (minOfferPrice + 2) - (minOfferPrice + 1),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:             "fail - continue a non-existing offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "202",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Buy-Order ID: 202: not found",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but not yours",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  creator_3_asBuyer, // not the buyer
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator, // not the existing offer's buyer
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "201",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "not the owner of the offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      creator_3_asBuyer,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but the Alias mismatch",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                "another",
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "201",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "alias mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    "another",
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but mis-match offer denom",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:        "201",
				AssetId:   rollApp_1_by1_asSrc.alias,
				AssetType: dymnstypes.TypeAlias,
				Params:    []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:     rollApp_2_by2_asDest.creator,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:    "201",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer denomination mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:        "201",
				AssetId:   rollApp_1_by1_asSrc.alias,
				AssetType: dymnstypes.TypeAlias,
				Params:    []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:     rollApp_2_by2_asDest.creator,
				OfferPrice: sdk.Coin{
					Denom:  "u" + denom,
					Amount: sdk.NewInt(minOfferPrice),
				},
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but new offer less than previous",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 1), // less
			existingBuyOrderId:    "201",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2), // keep
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but new offer equals to previous",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice + 2),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice + 2), // same
			existingBuyOrderId:    "201",
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Len(t, dk.GetAllBuyOrders(ctx), 1)
				require.Nil(t, dk.GetBuyOrder(ctx, "202"))
			},
		},
		{
			name:                  "fail - destination RollApp is not found",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 creator_2_asBuyer,
			dstRollAppId:          "nah_0-0",
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
			},
			wantErr:                true,
			wantErrContains:        "destination Roll-App does not exists",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
				require.Nil(t, dk.GetBuyOrder(ctx, "201"))
			},
		},
		{
			name:                  "fail - destination RollApp is not found",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_3_by3_asDest_noAlias},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 creator_2_asBuyer,
			dstRollAppId:          rollApp_3_by3_asDest_noAlias.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
			},
			wantErr:                true,
			wantErrContains:        "not the owner of the RollApp",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
				require.Nil(t, dk.GetBuyOrder(ctx, "201"))
			},
		},
		{
			name:                  "fail - destination RollApp is the same as source",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_1_by1_asSrc.creator,
			dstRollAppId:          rollApp_1_by1_asSrc.rollAppID,
			offer:                 dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalBuyerBalance:  minOfferPrice + 2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
			},
			wantErr:                true,
			wantErrContains:        "destination Roll-App ID is the same as the source",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  minOfferPrice + 2,
			wantMinConsumeGas:      1,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Empty(t, dk.GetAllBuyOrders(ctx))
				require.Nil(t, dk.GetBuyOrder(ctx, "201"))
			},
		},
		{
			name:                        "pass - reverse record added after successful offer",
			existingRollApps:            []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest, rollApp_3_by3_asDest_noAlias},
			existingOffer:               nil,
			alias:                       rollApp_1_by1_asSrc.alias,
			buyer:                       rollApp_2_by2_asDest.creator,
			dstRollAppId:                rollApp_2_by2_asDest.rollAppID,
			offer:                       dymnsutils.TestCoin(minOfferPrice),
			existingBuyOrderId:          "",
			originalModuleBalance:       5,
			originalBuyerBalance:        minOfferPrice + 2,
			originalAnotherBuyerBalance: minOfferPrice,
			wantErr:                     false,
			wantBuyOrderId:              "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice),
			},
			wantLaterModuleBalance: 5 + minOfferPrice,
			wantLaterBuyerBalance:  (minOfferPrice + 2) - minOfferPrice,
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"201"}, orderIds.OrderIds)

				offers, err := dk.GetBuyOrdersOfAlias(ctx, rollApp_1_by1_asSrc.alias)
				require.NoError(t, err)
				require.Equal(t, "201", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"201"}, orderIds.OrderIds)

				offers, err = dk.GetBuyOrdersByBuyer(ctx, rollApp_2_by2_asDest.creator)
				require.NoError(t, err)
				require.Equal(t, "201", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   rollApp_1_by1_asSrc.alias,
					AssetType: dymnstypes.TypeAlias,
					Params:    []string{rollApp_3_by3_asDest_noAlias.rollAppID},
					Buyer:     rollApp_3_by3_asDest_noAlias.creator,
					Offer:     dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "202", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_3_by3_asDest_noAlias.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"202"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"201"}, orderIds.OrderIds)

				key = dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"201", "202"}, orderIds.OrderIds)
			},
		},
		{
			name:             "pass - reverse record added after successful offer extends",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest, rollApp_3_by3_asDest_noAlias},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             dymnsutils.TestCoin(minOfferPrice),
				CounterpartyOfferPrice: nil,
			},
			alias:                       rollApp_1_by1_asSrc.alias,
			buyer:                       rollApp_2_by2_asDest.creator,
			dstRollAppId:                rollApp_2_by2_asDest.rollAppID,
			offer:                       dymnsutils.TestCoin(minOfferPrice + 1),
			existingBuyOrderId:          "202",
			originalModuleBalance:       0,
			originalBuyerBalance:        1,
			originalAnotherBuyerBalance: minOfferPrice,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dk.SetCountBuyOrders(ctx, 2)

				err := dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, rollApp_2_by2_asDest.creator, "202")
				require.NoError(t, err)

				err = dk.AddReverseMappingAssetIdToBuyOrder(ctx, rollApp_1_by1_asSrc.alias, dymnstypes.TypeAlias, "202")
				require.NoError(t, err)
			},
			wantErr:        false,
			wantBuyOrderId: "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: dymnsutils.TestCoin(minOfferPrice + 1),
			},
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  0,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"202"}, orderIds.OrderIds)

				offers, err := dk.GetBuyOrdersOfAlias(ctx, rollApp_1_by1_asSrc.alias)
				require.NoError(t, err)
				require.Equal(t, "202", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"202"}, orderIds.OrderIds)

				offers, err = dk.GetBuyOrdersByBuyer(ctx, rollApp_2_by2_asDest.creator)
				require.NoError(t, err)
				require.Equal(t, "202", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   rollApp_1_by1_asSrc.alias,
					AssetType: dymnstypes.TypeAlias,
					Params:    []string{rollApp_3_by3_asDest_noAlias.rollAppID},
					Buyer:     rollApp_3_by3_asDest_noAlias.creator,
					Offer:     dymnsutils.TestCoin(minOfferPrice),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, "203", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_3_by3_asDest_noAlias.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"203"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"202"}, orderIds.OrderIds)

				key = dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds = dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key)
				require.Equal(t, []string{"202", "203"}, orderIds.OrderIds)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, dk, rk, bk := setupTest()

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
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(creator_3_asBuyer),
					dymnsutils.TestCoins(tt.originalAnotherBuyerBalance),
				)
				require.NoError(t, err)
			}

			for _, existingRollApp := range tt.existingRollApps {
				rk.SetRollapp(ctx, rollapptypes.Rollapp{
					RollappId: existingRollApp.rollAppID,
					Owner:     existingRollApp.creator,
				})
				if existingRollApp.alias != "" {
					err := dk.SetAliasForRollAppId(ctx, existingRollApp.rollAppID, existingRollApp.alias)
					require.NoError(t, err)
				}
			}

			if tt.existingOffer != nil {
				err := dk.SetBuyOrder(ctx, *tt.existingOffer)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).PlaceBuyOrder(ctx, &dymnstypes.MsgPlaceBuyOrder{
				AssetId:         tt.alias,
				AssetType:       dymnstypes.TypeAlias,
				Buyer:           tt.buyer,
				Params:          []string{tt.dstRollAppId},
				ContinueOrderId: tt.existingBuyOrderId,
				Offer:           tt.offer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOrder(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.buyer), denom)
				require.Equal(t, tt.wantLaterBuyerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				for _, existingRollApp := range tt.existingRollApps {
					rollApp, found := rk.GetRollapp(ctx, existingRollApp.rollAppID)
					require.True(t, found)
					require.Equal(t, existingRollApp.creator, rollApp.Owner)
					if existingRollApp.alias != "" {
						requireAssignedAliasPairs(existingRollApp.rollAppID, existingRollApp.alias, t, ctx, dk)
					} else {
						requireRollAppHasNoAlias(existingRollApp.rollAppID, t, ctx, dk)
					}
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
			require.Equal(t, tt.wantBuyOrderId, resp.OrderId)
		})
	}
}
