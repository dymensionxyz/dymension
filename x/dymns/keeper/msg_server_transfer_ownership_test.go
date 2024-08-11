package keeper_test

import (
	"fmt"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_TransferDymNameOwnership() {
	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).TransferDymNameOwnership(s.ctx, &dymnstypes.MsgTransferDymNameOwnership{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
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
				ExpireAt:   s.now.Unix() + 100,
			},
			wantErr:         true,
			wantErrContains: "not the owner of the Dym-Name",
		},
		{
			name: "fail - reject if Dym-Name expired",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() - 1,
			},
			wantErr:         true,
			wantErrContains: "Dym-Name is already expired",
		},
		{
			name: "fail - reject if new owner is the same as current owner",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
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
				ExpireAt:   s.now.Unix() + 100,
			},
			sellOrder: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  1,
				MinPrice:  s.coin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			sellOrder: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Unix() + 100,
				MinPrice:  s.coin(100),
			},
			wantErr:         true,
			wantErrContains: "can not transfer ownership while there is an active Sell Order",
		},
		{
			name: "fail - reject if Sell Order exists, not finished SO",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
			},
			sellOrder: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Unix() + 100,
				MinPrice:  s.coin(100),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  s.coin(200),
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
				ExpireAt:   s.now.Unix() + 100,
			},
			sellOrder: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Unix() + 100,
				MinPrice:  s.coin(100),
				SellPrice: uptr.To(s.coin(200)),
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  s.coin(200),
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
				ExpireAt:   s.now.Unix() + 100,
			},
		},
		{
			name: "pass - can transfer ownership",
			dymName: &dymnstypes.DymName{
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() + 100,
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
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.dymName != nil {
				if tt.dymName.Name == "" {
					tt.dymName.Name = recordName
				}
				s.setDymNameWithFunctionsAfter(*tt.dymName)
			}

			if tt.dymName != nil {
				// setup historical SO

				so := &dymnstypes.SellOrder{
					AssetId:   recordName,
					AssetType: dymnstypes.TypeName,
					MinPrice:  s.coin(100),
					ExpireAt:  1,
				}
				s.Require().NoError(s.dymNsKeeper.SetSellOrder(s.ctx, *so))

				err := s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, recordName, so.AssetType)
				s.Require().NoError(err)

				s.Require().NotEmpty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, recordName, so.AssetType))
			}

			if tt.sellOrder != nil {
				s.Require().NotNil(tt.dymName, "bad test setup")
				tt.sellOrder.AssetId = recordName
				s.Require().NoError(s.dymNsKeeper.SetSellOrder(s.ctx, *tt.sellOrder))
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
			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).TransferDymNameOwnership(s.ctx, msg)
			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, recordName)

			if tt.dymName != nil {
				s.Require().NotNil(laterDymName)
			} else {
				s.Require().Nil(laterDymName)
			}

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				s.Require().Nil(resp)

				if tt.dymName != nil {
					s.Require().Equal(*tt.dymName, *laterDymName, "Dym-Name should not be changed")

					if tt.dymName.ExpireAt > s.now.Unix() {
						list, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, tt.dymName.Owner)
						// GetDymNamesOwnedBy does not return expired Dym-Names
						s.Require().NoError(err)
						s.Require().Len(list, 1, "reverse mapping should be kept")

						names, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, tt.dymName.Owner)
						s.Require().NoError(err)
						s.Require().Len(names, 1, "reverse mapping should be kept")

						names, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx,
							sdk.MustAccAddressFromBech32(tt.dymName.Owner).Bytes(),
						)
						s.Require().NoError(err)
						s.Require().Len(names, 1, "reverse mapping should be kept")
					}

					s.Require().NotEmpty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, recordName, dymnstypes.TypeName), "historical SO should be kept")
				}
				return
			}

			s.Require().NotNil(tt.dymName, "bad test setup")

			s.Require().NoError(err)

			s.Require().NotNil(resp)

			previousOwner := ownerA

			s.Require().NotNil(laterDymName)

			s.Require().Equal(
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
			s.Require().Equal(wantLaterDymName, *laterDymName)

			list, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, previousOwner)
			s.Require().NoError(err)
			s.Require().Empty(list, "reverse mapping of previous owner should be removed")

			names, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, previousOwner)
			s.Require().NoError(err)
			s.Require().Empty(names, "reverse mapping of previous owner should be removed")

			names, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx,
				sdk.MustAccAddressFromBech32(previousOwner).Bytes(),
			)
			s.Require().NoError(err)
			s.Require().Empty(names, "reverse mapping of previous owner should be removed")

			list, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, useNewOwner)
			s.Require().NoError(err)
			s.Require().Len(list, 1, "reverse mapping of new owner should be added")

			names, err = s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, useNewOwner)
			s.Require().NoError(err)
			s.Require().Len(names, 1, "reverse mapping of new owner should be added")

			names, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx,
				sdk.MustAccAddressFromBech32(useNewOwner).Bytes(),
			)
			s.Require().NoError(err)
			s.Require().Len(names, 1, "reverse mapping of new owner should be added")

			s.Require().Empty(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, recordName, dymnstypes.TypeName), "historical SO should be removed")
		})
	}
}
