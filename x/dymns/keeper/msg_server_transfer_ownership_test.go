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

func Test_msgServer_TransferDymNameOwnership(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).TransferDymNameOwnership(ctx, &dymnstypes.MsgTransferDymNameOwnership{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	ownerA := testAddr(1).bech32()
	newOwnerA := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()
	bidderA := testAddr(4).bech32()

	const recordName = "my-name"

	tests := []struct {
		name            string
		dymName         *dymnstypes.DymName
		sellOrder       *dymnstypes.SellOrder
		customNewOwner  string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fail - Dym-Name does not exists",
			dymName:         nil,
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", recordName),
		},
		{
			name: "fail - reject if not owned",
			dymName: &dymnstypes.DymName{
				Owner:      anotherA,
				Controller: anotherA,
				ExpireAt:   now.Unix() + 1,
			},
			wantErr:         true,
			wantErrContains: "not the owner of the Dym-Name",
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() - 1,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
		},
		{
			name: "fail - reject if new owner is the same as current owner",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
			customNewOwner:  ownerA,
			wantErr:         true,
			wantErrContains: "new owner must be different from the current owner",
		},
		{
			name: "fail - reject if Sell Order exists, expired SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				Type:     dymnstypes.NameOrder,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				Type:     dymnstypes.NameOrder,
				ExpireAt: now.Unix() + 1,
				MinPrice: dymnsutils.TestCoin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				Type:     dymnstypes.NameOrder,
				ExpireAt: now.Unix() + 1,
				MinPrice: dymnsutils.TestCoin(100),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  dymnsutils.TestCoin(200),
				},
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, completed SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				Type:      dymnstypes.NameOrder,
				ExpireAt:  now.Unix() + 1,
				MinPrice:  dymnsutils.TestCoin(100),
				SellPrice: dymnsutils.TestCoinP(200),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  dymnsutils.TestCoin(200),
				},
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "pass - can transfer ownership",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
			},
		},
		{
			name: "pass - can transfer ownership",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "a",
					Value:   anotherA,
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			if tt.dymName != nil {
				if tt.dymName.Name == "" {
					tt.dymName.Name = recordName
				}
				setDymNameWithFunctionsAfter(ctx, *tt.dymName, t, dk)
			}

			if tt.dymName != nil {
				// setup historical SO

				so := &dymnstypes.SellOrder{
					GoodsId:  recordName,
					Type:     dymnstypes.NameOrder,
					MinPrice: dymnsutils.TestCoin(100),
					ExpireAt: 1,
				}
				require.NoError(t, dk.SetSellOrder(ctx, *so))

				err := dk.MoveSellOrderToHistorical(ctx, recordName, so.Type)
				require.NoError(t, err)

				require.NotEmpty(t, dk.GetHistoricalSellOrders(ctx, recordName, so.Type))
			}

			if tt.sellOrder != nil {
				require.NotNil(t, tt.dymName, "bad test setup")
				tt.sellOrder.GoodsId = recordName
				require.NoError(t, dk.SetSellOrder(ctx, *tt.sellOrder))
			}

			useNewOwner := newOwnerA
			if tt.customNewOwner != "" {
				useNewOwner = tt.customNewOwner
			}

			msg := &dymnstypes.MsgTransferDymNameOwnership{
				Name:     recordName,
				Owner:    ownerA,
				NewOwner: useNewOwner,
			}
			resp, err := dymnskeeper.NewMsgServerImpl(dk).TransferDymNameOwnership(ctx, msg)
			laterDymName := dk.GetDymName(ctx, recordName)

			if tt.dymName != nil {
				require.NotNil(t, laterDymName)
			} else {
				require.Nil(t, laterDymName)
			}

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				require.Nil(t, resp)

				if tt.dymName != nil {
					require.Equal(t, *tt.dymName, *laterDymName, "Dym-Name should not be changed")

					if tt.dymName.ExpireAt > now.Unix() {
						list, err := dk.GetDymNamesOwnedBy(ctx, tt.dymName.Owner)
						// GetDymNamesOwnedBy does not return expired Dym-Names
						require.NoError(t, err)
						require.Len(t, list, 1, "reverse mapping should be kept")

						names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, tt.dymName.Owner)
						require.NoError(t, err)
						require.Len(t, names, 1, "reverse mapping should be kept")

						names, err = dk.GetDymNamesContainsFallbackAddress(ctx,
							sdk.MustAccAddressFromBech32(tt.dymName.Owner).Bytes(),
						)
						require.NoError(t, err)
						require.Len(t, names, 1, "reverse mapping should be kept")
					}

					require.NotEmpty(t, dk.GetHistoricalSellOrders(ctx, recordName, dymnstypes.NameOrder), "historical SO should be kept")
				}
				return
			}

			require.NotNil(t, tt.dymName, "bad test setup")

			require.NoError(t, err)

			require.NotNil(t, resp)

			previousOwner := ownerA

			require.NotNil(t, laterDymName)

			require.Equal(t,
				tt.dymName.ExpireAt, laterDymName.ExpireAt,
				"expiration date should not be changed",
			)

			wantLaterDymName := dymnstypes.DymName{
				Name:       recordName,
				Owner:      useNewOwner,
				Controller: useNewOwner,
				ExpireAt:   tt.dymName.ExpireAt,
				Configs:    nil,
			}
			require.Equal(t, wantLaterDymName, *laterDymName)

			list, err := dk.GetDymNamesOwnedBy(ctx, previousOwner)
			require.NoError(t, err)
			require.Empty(t, list, "reverse mapping of previous owner should be removed")

			names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, previousOwner)
			require.NoError(t, err)
			require.Empty(t, names, "reverse mapping of previous owner should be removed")

			names, err = dk.GetDymNamesContainsFallbackAddress(ctx,
				sdk.MustAccAddressFromBech32(previousOwner).Bytes(),
			)
			require.NoError(t, err)
			require.Empty(t, names, "reverse mapping of previous owner should be removed")

			list, err = dk.GetDymNamesOwnedBy(ctx, useNewOwner)
			require.NoError(t, err)
			require.Len(t, list, 1, "reverse mapping of new owner should be added")

			names, err = dk.GetDymNamesContainsConfiguredAddress(ctx, useNewOwner)
			require.NoError(t, err)
			require.Len(t, names, 1, "reverse mapping of new owner should be added")

			names, err = dk.GetDymNamesContainsFallbackAddress(ctx,
				sdk.MustAccAddressFromBech32(useNewOwner).Bytes(),
			)
			require.NoError(t, err)
			require.Len(t, names, 1, "reverse mapping of new owner should be added")

			require.Empty(t, dk.GetHistoricalSellOrders(ctx, recordName, dymnstypes.NameOrder), "historical SO should be removed")
		})
	}
}
