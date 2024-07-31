package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_msgServer_TransferOwnership(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockHeader(tmproto.Header{
			Time: now,
		})

		return dk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).TransferOwnership(ctx, &dymnstypes.MsgTransferOwnership{})
			return err
		}, dymnstypes.ErrValidationFailed.Error())
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const newOwner = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	const anotherAcc = "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"
	const recordName = "bonded-pool"

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
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name: "fail - reject if not owned",
			dymName: &dymnstypes.DymName{
				Owner:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
				Controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
				ExpireAt:   now.Unix() + 1,
			},
			wantErr:         true,
			wantErrContains: sdkerrors.ErrUnauthorized.Error(),
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() - 1,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
		},
		{
			name: "fail - reject if new owner is the same as current owner",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
			customNewOwner:  owner,
			wantErr:         true,
			wantErrContains: "new owner is same as current owner",
		},
		{
			name: "fail - reject if Sell Order exists, expired SO",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				ExpireAt: now.Unix() + 1,
				MinPrice: dymnsutils.TestCoin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				ExpireAt: now.Unix() + 1,
				MinPrice: dymnsutils.TestCoin(100),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
					Price:  dymnsutils.TestCoin(200),
				},
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, completed SO",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
			sellOrder: &dymnstypes.SellOrder{
				ExpireAt:  now.Unix() + 1,
				MinPrice:  dymnsutils.TestCoin(100),
				SellPrice: dymnsutils.TestCoinP(200),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
					Price:  dymnsutils.TestCoin(200),
				},
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an Sell Order",
		},
		{
			name: "success - can transfer ownership",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
			},
		},
		{
			name: "success - can transfer ownership",
			dymName: &dymnstypes.DymName{
				Owner:      owner,
				Controller: owner,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "a",
					Value:   anotherAcc,
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
					Name:     recordName,
					MinPrice: dymnsutils.TestCoin(100),
					ExpireAt: 1,
				}
				require.NoError(t, dk.SetSellOrder(ctx, *so))

				err := dk.MoveSellOrderToHistorical(ctx, recordName)
				require.NoError(t, err)

				require.NotEmpty(t, dk.GetHistoricalSellOrders(ctx, recordName))
			}

			if tt.sellOrder != nil {
				require.NotNil(t, tt.dymName, "bad test setup")
				tt.sellOrder.Name = recordName
				require.NoError(t, dk.SetSellOrder(ctx, *tt.sellOrder))
			}

			useNewOwner := newOwner
			if tt.customNewOwner != "" {
				useNewOwner = tt.customNewOwner
			}

			msg := &dymnstypes.MsgTransferOwnership{
				Name:     recordName,
				Owner:    owner,
				NewOwner: useNewOwner,
			}
			resp, err := dymnskeeper.NewMsgServerImpl(dk).TransferOwnership(ctx, msg)
			laterDymName := dk.GetDymName(ctx, recordName)

			if tt.dymName != nil {
				require.NotNil(t, laterDymName)
			} else {
				require.Nil(t, laterDymName)
			}

			if tt.wantErr {
				require.Error(t, err)

				require.Nil(t, resp)

				if tt.dymName != nil {
					require.Equal(t, *tt.dymName, *laterDymName, "Dym-Name should not be changed")

					if tt.dymName.ExpireAt > now.Unix() {
						list, err := dk.GetDymNamesOwnedBy(ctx, tt.dymName.Owner, now.Unix())
						// GetDymNamesOwnedBy does not return expired Dym-Names
						require.NoError(t, err)
						require.Len(t, list, 1, "reverse mapping should be kept")

						names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, tt.dymName.Owner, now.Unix())
						require.NoError(t, err)
						require.Len(t, names, 1, "reverse mapping should be kept")

						names, err = dk.GetDymNamesContainsHexAddress(ctx,
							sdk.MustAccAddressFromBech32(tt.dymName.Owner).Bytes(),
							now.Unix(),
						)
						require.NoError(t, err)
						require.Len(t, names, 1, "reverse mapping should be kept")
					}

					require.NotEmpty(t, dk.GetHistoricalSellOrders(ctx, recordName), "historical SO should be kept")
				}
				return
			}

			require.NotNil(t, tt.dymName, "bad test setup")

			require.NoError(t, err)

			require.NotNil(t, resp)

			previousOwner := owner

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

			list, err := dk.GetDymNamesOwnedBy(ctx, previousOwner, now.Unix())
			require.NoError(t, err)
			require.Empty(t, list, "reverse mapping of previous owner should be removed")

			names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, previousOwner, now.Unix())
			require.NoError(t, err)
			require.Empty(t, names, "reverse mapping of previous owner should be removed")

			names, err = dk.GetDymNamesContainsHexAddress(ctx,
				sdk.MustAccAddressFromBech32(previousOwner).Bytes(),
				now.Unix(),
			)
			require.NoError(t, err)
			require.Empty(t, names, "reverse mapping of previous owner should be removed")

			list, err = dk.GetDymNamesOwnedBy(ctx, useNewOwner, now.Unix())
			require.NoError(t, err)
			require.Len(t, list, 1, "reverse mapping of new owner should be added")

			names, err = dk.GetDymNamesContainsConfiguredAddress(ctx, useNewOwner, now.Unix())
			require.NoError(t, err)
			require.Len(t, names, 1, "reverse mapping of new owner should be added")

			names, err = dk.GetDymNamesContainsHexAddress(ctx,
				sdk.MustAccAddressFromBech32(useNewOwner).Bytes(),
				now.Unix(),
			)
			require.NoError(t, err)
			require.Len(t, names, 1, "reverse mapping of new owner should be added")

			require.Empty(t, dk.GetHistoricalSellOrders(ctx, recordName), "historical SO should be removed")
		})
	}
}
