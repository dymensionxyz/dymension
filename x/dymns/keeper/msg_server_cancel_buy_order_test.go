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

func Test_msgServer_CancelBuyOrder_DymName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	buyerA := testAddr(1).bech32()
	anotherBuyerA := testAddr(2).bech32()
	ownerA := testAddr(3).bech32()

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
			_, err := dymnskeeper.NewMsgServerImpl(dk).CancelBuyOrder(ctx, &dymnstypes.MsgCancelBuyOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}

	offer := &dymnstypes.BuyOrder{
		Id:         "101",
		AssetId:    dymName.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	offerByAnother := &dymnstypes.BuyOrder{
		Id:         "10999",
		AssetId:    dymName.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      anotherBuyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		buyer                  string
		originalModuleBalance  int64
		originalBuyerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterModuleBalance int64
		wantLaterBuyerBalance  int64
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                   "pass - can cancel offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   0,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "pass - cancel offer will refund the buyer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			buyer:                  offer.Buyer,
			originalModuleBalance:  1 + offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   2,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2 + offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                  "pass - cancel offer will remove the offer record",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.NotNil(t, dk.GetBuyOrder(ctx, offer.Id))
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Nil(t, dk.GetBuyOrder(ctx, offer.Id))
			},
		},
		{
			name:                  "pass - cancel offer will remove reverse mapping records",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				buyOrders, err := dk.GetBuyOrdersByBuyer(ctx, offer.Buyer)
				require.NoError(t, err)
				require.Len(t, buyOrders, 1)

				buyOrders, err = dk.GetBuyOrdersOfDymName(ctx, offer.AssetId)
				require.NoError(t, err)
				require.Len(t, buyOrders, 1)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				buyOrders, err := dk.GetBuyOrdersByBuyer(ctx, offer.Buyer)
				require.NoError(t, err)
				require.Empty(t, buyOrders)

				buyOrders, err = dk.GetBuyOrdersOfDymName(ctx, offer.AssetId)
				require.NoError(t, err)
				require.Empty(t, buyOrders)
			},
		},
		{
			name:                  "pass - can cancel offer when trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			buyer:                 offer.Buyer,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingName = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          nil,
			buyOrderId:             "102142142",
			buyer:                  buyerA,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 102142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             "102142142",
			buyer:                  offer.Buyer,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 102142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel offer with different buyer",
			existingDymName:        dymName,
			existingOffer:          offerByAnother,
			buyOrderId:             "10999",
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
			buyOrderId:             offer.Id,
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
				err := dk.SetBuyOrder(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingAssetIdToBuyOrder(ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).CancelBuyOrder(ctx, &dymnstypes.MsgCancelBuyOrder{
				OrderId: tt.buyOrderId,
				Buyer:   tt.buyer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOrder(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetBuyOrder(ctx, tt.buyOrderId)
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

//goland:noinspection GoSnakeCaseUsage
func Test_msgServer_CancelBuyOrder_Alias(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	creator_3_asAnotherBuyer := testAddr(3).bech32()

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

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		ctx, dk, _, _ := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).CancelBuyOrder(ctx, &dymnstypes.MsgCancelBuyOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	type rollapp struct {
		rollAppId string
		creator   string
		aliases   []string
	}

	rollApp_One_By1 := rollapp{
		rollAppId: "rollapp_1-1",
		creator:   creator_1_asOwner,
		aliases:   []string{"one"},
	}
	rollApp_Two_By2 := rollapp{
		rollAppId: "rollapp_2-2",
		creator:   creator_2_asBuyer,
		aliases:   []string{"two"},
	}
	rollApp_Three_By3 := rollapp{
		rollAppId: "rollapp_3-3",
		creator:   creator_3_asAnotherBuyer,
		aliases:   []string{},
	}

	offerAliasOfRollAppOne := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1),
		AssetId:    rollApp_One_By1.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Two_By2.rollAppId},
		Buyer:      rollApp_Two_By2.creator,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	offerAliasOfRollAppOneByAnother := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    rollApp_One_By1.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Three_By3.rollAppId},
		Buyer:      rollApp_Three_By3.creator,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingRollApps       []rollapp
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		buyer                  string
		originalModuleBalance  int64
		originalBuyerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterModuleBalance int64
		wantLaterBuyerBalance  int64
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                   "pass - can cancel offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   0,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "pass - cancel offer will refund the buyer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  1 + offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:   2,
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2 + offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                  "pass - cancel offer will remove the offer record",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.NotNil(t, dk.GetBuyOrder(ctx, offerAliasOfRollAppOne.Id))
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				require.Nil(t, dk.GetBuyOrder(ctx, offerAliasOfRollAppOne.Id))
			},
		},
		{
			name:                  "pass - cancel offer will remove reverse mapping records",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				orderIds, err := dk.GetBuyOrdersByBuyer(ctx, offerAliasOfRollAppOne.Buyer)
				require.NoError(t, err)
				require.Len(t, orderIds, 1)

				orderIds, err = dk.GetBuyOrdersOfAlias(ctx, offerAliasOfRollAppOne.AssetId)
				require.NoError(t, err)
				require.Len(t, orderIds, 1)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				orderIds, err := dk.GetBuyOrdersByBuyer(ctx, offerAliasOfRollAppOne.Buyer)
				require.NoError(t, err)
				require.Empty(t, orderIds)

				orderIds, err = dk.GetBuyOrdersOfAlias(ctx, offerAliasOfRollAppOne.AssetId)
				require.NoError(t, err)
				require.Empty(t, orderIds)
			},
		},
		{
			name:                  "pass - cancel offer will NOT remove reverse mapping records of other offers",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.SetBuyOrder(ctx, *offerAliasOfRollAppOneByAnother)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(
					ctx,
					offerAliasOfRollAppOneByAnother.Buyer,
					offerAliasOfRollAppOneByAnother.Id,
				)
				require.NoError(t, err)

				err = dk.AddReverseMappingAssetIdToBuyOrder(
					ctx,
					offerAliasOfRollAppOneByAnother.AssetId, offerAliasOfRollAppOneByAnother.AssetType,
					offerAliasOfRollAppOneByAnother.Id,
				)
				require.NoError(t, err)

				orderIds, err := dk.GetBuyOrdersByBuyer(ctx, offerAliasOfRollAppOne.Buyer)
				require.NoError(t, err)
				require.Len(t, orderIds, 1)

				orderIds, err = dk.GetBuyOrdersOfAlias(ctx, offerAliasOfRollAppOne.AssetId)
				require.NoError(t, err)
				require.Len(t, orderIds, 2)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				orderIds, err := dk.GetBuyOrdersByBuyer(ctx, offerAliasOfRollAppOne.Buyer)
				require.NoError(t, err)
				require.Empty(t, orderIds)

				orderIds, err = dk.GetBuyOrdersOfAlias(ctx, offerAliasOfRollAppOne.AssetId)
				require.NoError(t, err)
				require.Len(t, orderIds, 1)
				require.Equal(t, offerAliasOfRollAppOneByAnother.Id, orderIds[0].Id)
			},
		},
		{
			name:                  "pass - can cancel offer when trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			buyer:                 offerAliasOfRollAppOne.Buyer,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalBuyerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingAlias = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasCloseBuyOrder,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          nil,
			buyOrderId:             "202142142",
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 202142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel non-existing offer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             "202142142",
			buyer:                  offerAliasOfRollAppOne.Buyer,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order ID: 202142142: not found",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - cannot cancel offer with different buyer",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOneByAnother,
			buyOrderId:             offerAliasOfRollAppOneByAnother.Id,
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  1,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "not the owner of the offer",
			wantLaterOffer:         offerAliasOfRollAppOneByAnother,
			wantLaterModuleBalance: 1,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not cancel if module account does not have enough balance to refund",
			existingRollApps:       []rollapp{rollApp_One_By1, rollApp_Two_By2},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			buyer:                  creator_2_asBuyer,
			originalModuleBalance:  0,
			originalBuyerBalance:   2,
			wantErr:                true,
			wantErrContains:        "insufficient funds",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterModuleBalance: 0,
			wantLaterBuyerBalance:  2,
			wantMinConsumeGas:      1,
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

			for _, rollApp := range tt.existingRollApps {
				rk.SetRollapp(ctx, rollapptypes.Rollapp{
					RollappId: rollApp.rollAppId,
					Owner:     rollApp.creator,
				})
				for _, alias := range rollApp.aliases {
					err := dk.SetAliasForRollAppId(ctx, rollApp.rollAppId, alias)
					require.NoError(t, err)
				}
			}

			if tt.existingOffer != nil {
				err := dk.SetBuyOrder(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingAssetIdToBuyOrder(ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).CancelBuyOrder(ctx, &dymnstypes.MsgCancelBuyOrder{
				OrderId: tt.buyOrderId,
				Buyer:   tt.buyer,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOrder(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetBuyOrder(ctx, tt.buyOrderId)
					require.Nil(t, laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.buyer), denom)
				require.Equal(t, tt.wantLaterBuyerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				for _, rollApp := range tt.existingRollApps {
					if len(rollApp.aliases) == 0 {
						requireRollAppHasNoAlias(rollApp.rollAppId, t, ctx, dk)
					} else {
						for _, alias := range rollApp.aliases {
							requireAliasLinkedToRollApp(alias, rollApp.rollAppId, t, ctx, dk)
						}
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
		})
	}
}
