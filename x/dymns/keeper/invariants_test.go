package keeper_test

import (
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestInvariants() {
	tests := []struct {
		name            string
		nameASoe        []dymnstypes.ActiveSellOrdersExpirationRecord
		aliasASoe       []dymnstypes.ActiveSellOrdersExpirationRecord
		sellOrders      []dymnstypes.SellOrder
		wantBroken      bool
		wantMsgContains string
	}{
		{
			name:       "pass - all empty",
			wantBroken: false,
		},
		{
			name: "pass - all correct",
			nameASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "name1",
					ExpireAt: 1,
				},
				{
					AssetId:  "name2",
					ExpireAt: 2,
				},
			},
			aliasASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "alias1",
					ExpireAt: 3,
				},
				{
					AssetId:  "alias2",
					ExpireAt: 4,
				},
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newDymNameSellOrder("name1").WithExpiry(1).WithMinPrice(1).Build(),
				s.newDymNameSellOrder("name2").WithExpiry(2).WithMinPrice(2).Build(),
				s.newAliasSellOrder("alias1").WithExpiry(3).WithMinPrice(3).Build(),
				s.newAliasSellOrder("alias2").WithExpiry(4).WithMinPrice(4).Build(),
			},
			wantBroken: false,
		},
		{
			name: "fail - missing a Dym-Name Sell-Order",
			nameASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "name1",
					ExpireAt: 1,
				},
				{
					AssetId:  "name2",
					ExpireAt: 2,
				},
			},
			aliasASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "alias1",
					ExpireAt: 3,
				},
				{
					AssetId:  "alias2",
					ExpireAt: 4,
				},
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newDymNameSellOrder("name1").WithExpiry(1).WithMinPrice(1).Build(),
				s.newAliasSellOrder("alias1").WithExpiry(3).WithMinPrice(3).Build(),
				s.newAliasSellOrder("alias2").WithExpiry(4).WithMinPrice(4).Build(),
			},
			wantBroken:      true,
			wantMsgContains: "sell order not found",
		},
		{
			name: "fail - missing a Alias sell order",
			nameASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "name1",
					ExpireAt: 1,
				},
				{
					AssetId:  "name2",
					ExpireAt: 2,
				},
			},
			aliasASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "alias1",
					ExpireAt: 3,
				},
				{
					AssetId:  "alias2",
					ExpireAt: 4,
				},
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newDymNameSellOrder("name1").WithExpiry(1).WithMinPrice(1).Build(),
				s.newDymNameSellOrder("name2").WithExpiry(2).WithMinPrice(2).Build(),
				s.newAliasSellOrder("alias1").WithExpiry(3).WithMinPrice(3).Build(),
			},
			wantBroken:      true,
			wantMsgContains: "sell order not found",
		},
		{
			name: "fail - mis-match expiry of a Dym-Name sell order",
			nameASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "name1",
					ExpireAt: 1,
				},
				{
					AssetId:  "name2",
					ExpireAt: 999,
				},
			},
			aliasASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "alias1",
					ExpireAt: 3,
				},
				{
					AssetId:  "alias2",
					ExpireAt: 4,
				},
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newDymNameSellOrder("name1").WithExpiry(1).WithMinPrice(1).Build(),
				s.newDymNameSellOrder("name2").WithExpiry(2).WithMinPrice(2).Build(),
				s.newAliasSellOrder("alias1").WithExpiry(3).WithMinPrice(3).Build(),
				s.newAliasSellOrder("alias2").WithExpiry(4).WithMinPrice(4).Build(),
			},
			wantBroken:      true,
			wantMsgContains: "sell order expiration mismatch",
		},
		{
			name: "fail - mis-match expiry of a Alias sell order",
			nameASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "name1",
					ExpireAt: 1,
				},
				{
					AssetId:  "name2",
					ExpireAt: 2,
				},
			},
			aliasASoe: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "alias1",
					ExpireAt: 3,
				},
				{
					AssetId:  "alias2",
					ExpireAt: 99,
				},
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newDymNameSellOrder("name1").WithExpiry(1).WithMinPrice(1).Build(),
				s.newDymNameSellOrder("name2").WithExpiry(2).WithMinPrice(2).Build(),
				s.newAliasSellOrder("alias1").WithExpiry(3).WithMinPrice(3).Build(),
				s.newAliasSellOrder("alias2").WithExpiry(4).WithMinPrice(4).Build(),
			},
			wantBroken:      true,
			wantMsgContains: "sell order expiration mismatch",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.nameASoe,
			}, dymnstypes.TypeName)
			s.Require().NoError(err)
			err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.aliasASoe,
			}, dymnstypes.TypeAlias)
			s.Require().NoError(err)

			for _, so := range tt.sellOrders {
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			}

			msg, broken := dymnskeeper.SellOrderExpirationInvariant(s.dymNsKeeper)(s.ctx)
			if tt.wantBroken {
				s.Require().True(broken)
				s.Require().NotEmpty(tt.wantMsgContains, "bad setup")
				s.Require().Contains(msg, tt.wantMsgContains)
			} else {
				s.Require().False(broken)
			}
		})
	}
}
