package keeper_test

import (
	"reflect"
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func Test_queryServer_Params(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	params := dk.GetParams(ctx)
	params.Misc.ProhibitSellDuration += time.Hour
	err := dk.SetParams(ctx, params)
	require.NoError(t, err)

	queryServer := dymnskeeper.NewQueryServerImpl(dk)

	resp, err := queryServer.Params(sdk.WrapSDKContext(ctx), &dymnstypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, params, resp.Params)
}

func Test_queryServer_DymName(t *testing.T) {
	t.Run("Dym-Name not found", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		queryServer := dymnskeeper.NewQueryServerImpl(dk)
		resp, err := queryServer.DymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryDymNameRequest{
			DymName: "not-exists",
		})
		require.NoError(t, err)
		require.Nil(t, resp.DymName)
	})

	now := time.Now().UTC()

	ownerA := testAddr(1).bech32()

	tests := []struct {
		name       string
		dymName    *dymnstypes.DymName
		queryName  string
		wantResult bool
	}{
		{
			name: "correct record",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: ownerA,
					},
				},
			},
			queryName:  "a",
			wantResult: true,
		},
		{
			name: "NOT expired record only",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() + 99,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: ownerA,
					},
				},
			},
			queryName:  "a",
			wantResult: true,
		},
		{
			name: "return nil for expired record",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   now.Unix() - 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: ownerA,
					},
				},
			},
			queryName:  "a",
			wantResult: false,
		},
		{
			name:       "return nil if not found",
			dymName:    nil,
			queryName:  "non-exists",
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)
			ctx = ctx.WithBlockTime(now)

			if tt.dymName != nil {
				err := dk.SetDymName(ctx, *tt.dymName)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)
			resp, err := queryServer.DymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryDymNameRequest{
				DymName: tt.queryName,
			})
			require.NoError(t, err, "should never returns error")
			require.NotNil(t, resp, "should never returns nil response")

			if !tt.wantResult {
				require.Nil(t, resp.DymName)
				return
			}

			require.NotNil(t, resp.DymName)
			require.Equal(t, *tt.dymName, *resp.DymName)
		})
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		queryServer := dymnskeeper.NewQueryServerImpl(dk)
		resp, err := queryServer.DymName(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func Test_queryServer_ResolveDymNameAddresses(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now).WithChainID(chainId)

	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()
	addr3a := testAddr(3).bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: addr1a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameA))

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: addr2a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameB))

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: addr3a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameC))

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{
			{
				Type:  dymnstypes.DymNameConfigType_DCT_NAME,
				Path:  "sub",
				Value: addr3a,
			},
			{
				Type:    dymnstypes.DymNameConfigType_DCT_NAME,
				ChainId: "blumbus_111-1",
				Path:    "",
				Value:   addr3a,
			},
		},
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameD))

	queryServer := dymnskeeper.NewQueryServerImpl(dk)

	resp, err := queryServer.ResolveDymNameAddresses(sdk.WrapSDKContext(ctx), &dymnstypes.ResolveDymNameAddressesRequest{
		Addresses: []string{
			"a.dymension_1100-1",
			"b.dymension_1100-1",
			"c.dymension_1100-1",
			"a.blumbus_111-1",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.ResolvedAddresses, 4)

	require.Equal(t, addr1a, resp.ResolvedAddresses[0].ResolvedAddress)
	require.Equal(t, addr2a, resp.ResolvedAddresses[1].ResolvedAddress)
	require.Equal(t, addr3a, resp.ResolvedAddresses[2].ResolvedAddress)
	require.Empty(t, resp.ResolvedAddresses[3].ResolvedAddress)
	require.NotEmpty(t, resp.ResolvedAddresses[3].Error)

	t.Run("reject nil request", func(t *testing.T) {
		resp, err := queryServer.ResolveDymNameAddresses(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("reject empty request", func(t *testing.T) {
		resp, err := queryServer.ResolveDymNameAddresses(
			sdk.WrapSDKContext(ctx),
			&dymnstypes.ResolveDymNameAddressesRequest{},
		)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("resolves default to owner if no config of default (without sub-name)", func(t *testing.T) {
		resp, err := queryServer.ResolveDymNameAddresses(
			sdk.WrapSDKContext(ctx),
			&dymnstypes.ResolveDymNameAddressesRequest{
				Addresses: []string{"d.dymension_1100-1", "d.blumbus_111-1"},
			},
		)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.ResolvedAddresses, 2)
		require.Equal(t, addr1a, resp.ResolvedAddresses[0].ResolvedAddress)
		require.Equal(t, addr3a, resp.ResolvedAddresses[1].ResolvedAddress)
	})
}

func Test_queryServer_DymNamesOwnedByAccount(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now).WithChainID(chainId)

	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()
	addr3a := testAddr(3).bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: addr1a,
		}},
	}
	setDymNameWithFunctionsAfter(ctx, dymNameA, t, dk)

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 1,
	}
	setDymNameWithFunctionsAfter(ctx, dymNameB, t, dk)

	dymNameCExpired := dymnstypes.DymName{
		Name:       "c",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() - 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: addr3a,
		}},
	}
	setDymNameWithFunctionsAfter(ctx, dymNameCExpired, t, dk)

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      addr3a,
		Controller: addr3a,
		ExpireAt:   now.Unix() + 1,
	}
	setDymNameWithFunctionsAfter(ctx, dymNameD, t, dk)

	queryServer := dymnskeeper.NewQueryServerImpl(dk)
	resp, err := queryServer.DymNamesOwnedByAccount(sdk.WrapSDKContext(ctx), &dymnstypes.QueryDymNamesOwnedByAccountRequest{
		Owner: addr1a,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.DymNames, 2)
	require.True(t, resp.DymNames[0].Name == dymNameA.Name || resp.DymNames[1].Name == dymNameA.Name)
	require.True(t, resp.DymNames[0].Name == dymNameB.Name || resp.DymNames[1].Name == dymNameB.Name)

	t.Run("reject nil request", func(t *testing.T) {
		resp, err := queryServer.DymNamesOwnedByAccount(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("reject invalid request", func(t *testing.T) {
		resp, err := queryServer.DymNamesOwnedByAccount(sdk.WrapSDKContext(ctx), &dymnstypes.QueryDymNamesOwnedByAccountRequest{
			Owner: "x",
		})
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func (s *KeeperTestSuite) Test_queryServer_SellOrder() {
	now := time.Now().UTC()

	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "goods",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	dymNameB := dymnstypes.DymName{
		Name:       "mood",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}

	rollAppC := newRollApp("central_1-1").WithAlias("goods")
	rollAppD := newRollApp("donut_2-1").WithAlias("donut")

	soDymNameA := s.newDymNameSellOrder(dymNameA.Name).WithMinPrice(100).Build()
	soAliasRollAppC := s.newAliasSellOrder(rollAppC.alias).WithMinPrice(100).Build()

	tests := []struct {
		name            string
		req             *dymnstypes.QuerySellOrderRequest
		preRunFunc      func(s *KeeperTestSuite)
		wantErr         bool
		wantErrContains string
		wantSellOrder   *dymnstypes.SellOrder
	}{
		{
			name: "pass - returns correct order, type Dym Name",
			preRunFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(s.ctx, soDymNameA)
				s.Require().NoError(err)
			},
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   dymNameA.Name,
				OrderType: dymnstypes.NameOrder.FriendlyString(),
			},
			wantErr:       false,
			wantSellOrder: &soDymNameA,
		},
		{
			name: "pass - returns correct order, type Alias",
			preRunFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(s.ctx, soAliasRollAppC)
				s.Require().NoError(err)
			},
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   rollAppC.alias,
				OrderType: dymnstypes.AliasOrder.FriendlyString(),
			},
			wantErr:       false,
			wantSellOrder: &soAliasRollAppC,
		},
		{
			name: "pass - returns correct order of same goods-id with multiple order types",
			preRunFunc: func(s *KeeperTestSuite) {
				s.Require().Equal(soDymNameA.GoodsId, soAliasRollAppC.GoodsId, "Dym-Name and Alias must be the same for this test")

				err := s.dymNsKeeper.SetSellOrder(s.ctx, soDymNameA)
				s.Require().NoError(err)

				err = s.dymNsKeeper.SetSellOrder(s.ctx, soAliasRollAppC)
				s.Require().NoError(err)
			},
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   dymNameA.Name,
				OrderType: dymnstypes.NameOrder.FriendlyString(),
			},
			wantErr:       false,
			wantSellOrder: &soDymNameA,
		},
		{
			name: "pass - returns correct order of same goods-id with multiple order types",
			preRunFunc: func(s *KeeperTestSuite) {
				s.Require().Equal(soDymNameA.GoodsId, soAliasRollAppC.GoodsId, "Dym-Name and Alias must be the same for this test")

				err := s.dymNsKeeper.SetSellOrder(s.ctx, soDymNameA)
				s.Require().NoError(err)

				err = s.dymNsKeeper.SetSellOrder(s.ctx, soAliasRollAppC)
				s.Require().NoError(err)
			},
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   rollAppC.alias,
				OrderType: dymnstypes.AliasOrder.FriendlyString(),
			},
			wantErr:       false,
			wantSellOrder: &soAliasRollAppC,
		},
		{
			name:            "fail - reject nil request",
			req:             nil,
			wantErr:         true,
			wantErrContains: "invalid request",
		},
		{
			name: "fail - reject bad Dym-Name request",
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   "$$$",
				OrderType: dymnstypes.NameOrder.FriendlyString(),
			},
			wantErr:         true,
			wantErrContains: "invalid Dym-Name",
		},
		{
			name: "fail - reject bad Alias request",
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   "$$$",
				OrderType: dymnstypes.AliasOrder.FriendlyString(),
			},
			wantErr:         true,
			wantErrContains: "invalid alias",
		},
		{
			name: "fail - reject unknown order type",
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   "goods",
				OrderType: "pseudo",
			},
			wantErr:         true,
			wantErrContains: "invalid order type",
		},
		{
			name:       "fail - reject if not found, type Dym Name",
			preRunFunc: nil,
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   dymNameB.Name,
				OrderType: dymnstypes.NameOrder.FriendlyString(),
			},
			wantErr:         true,
			wantErrContains: "no active Sell Order for Dym-Name",
		},
		{
			name:       "fail - reject if not found, type Alias",
			preRunFunc: nil,
			req: &dymnstypes.QuerySellOrderRequest{
				GoodsId:   rollAppD.alias,
				OrderType: dymnstypes.AliasOrder.FriendlyString(),
			},
			wantErr:         true,
			wantErrContains: "no active Sell Order for Alias",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			resp, err := dymnskeeper.NewQueryServerImpl(s.dymNsKeeper).SellOrder(sdk.WrapSDKContext(s.ctx), tt.req)
			if tt.wantErr {
				s.requireErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Equal(*tt.wantSellOrder, resp.Result)
		})
	}
}

func Test_queryServer_HistoricalSellOrderOfDymName(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now).WithChainID(chainId)

	addr1a := testAddr(1).bech32()
	addr2a := testAddr(2).bech32()
	addr3a := testAddr(3).bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 100,
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameA))
	for r := int64(1); r <= 5; r++ {
		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
			GoodsId:   dymNameA.Name,
			Type:      dymnstypes.NameOrder,
			ExpireAt:  now.Unix() + r,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(200),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: addr3a,
				Price:  dymnsutils.TestCoin(200),
			},
		})
		require.NoError(t, err)
		err = dk.MoveSellOrderToHistorical(ctx, dymNameA.Name, dymnstypes.NameOrder)
		require.NoError(t, err)
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      addr1a,
		Controller: addr2a,
		ExpireAt:   now.Unix() + 100,
	}
	require.NoError(t, dk.SetDymName(ctx, dymNameB))
	for r := int64(1); r <= 3; r++ {
		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
			GoodsId:   dymNameB.Name,
			Type:      dymnstypes.NameOrder,
			ExpireAt:  now.Unix() + r,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: addr3a,
				Price:  dymnsutils.TestCoin(300),
			},
		})
		require.NoError(t, err)
		err = dk.MoveSellOrderToHistorical(ctx, dymNameB.Name, dymnstypes.NameOrder)
		require.NoError(t, err)
	}

	queryServer := dymnskeeper.NewQueryServerImpl(dk)
	resp, err := queryServer.HistoricalSellOrderOfDymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryHistoricalSellOrderOfDymNameRequest{
		DymName: dymNameA.Name,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Result, 5)

	resp, err = queryServer.HistoricalSellOrderOfDymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryHistoricalSellOrderOfDymNameRequest{
		DymName: dymNameB.Name,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Result, 3)

	t.Run("returns empty for non-exists Dym-Name", func(t *testing.T) {
		resp, err := queryServer.HistoricalSellOrderOfDymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryHistoricalSellOrderOfDymNameRequest{
			DymName: "not-exists",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.Result)
	})

	t.Run("reject nil request", func(t *testing.T) {
		resp, err := queryServer.HistoricalSellOrderOfDymName(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("reject invalid request", func(t *testing.T) {
		resp, err := queryServer.HistoricalSellOrderOfDymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryHistoricalSellOrderOfDymNameRequest{
			DymName: "$$$",
		})
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func Test_queryServer_EstimateRegisterName(t *testing.T) {
	now := time.Now().UTC()

	const denom = "atom"
	const price1L int64 = 9
	const price2L int64 = 8
	const price3L int64 = 7
	const price4L int64 = 6
	const price5PlusL int64 = 5
	const extendsPrice int64 = 4

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		params := dymnstypes.DefaultParams()
		params.Price.PriceDenom = denom
		params.Price.NamePrice_1Letter = sdk.NewInt(price1L)
		params.Price.NamePrice_2Letters = sdk.NewInt(price2L)
		params.Price.NamePrice_3Letters = sdk.NewInt(price3L)
		params.Price.NamePrice_4Letters = sdk.NewInt(price4L)
		params.Price.NamePrice_5PlusLetters = sdk.NewInt(price5PlusL)
		params.Price.PriceExtends = sdk.NewInt(extendsPrice)
		params.Misc.GracePeriodDuration = 1 * 24 * time.Hour
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)

		return dk, ctx
	}

	buyerA := testAddr(1).bech32()
	previousOwnerA := testAddr(2).bech32()

	tests := []struct {
		name               string
		dymName            string
		existingDymName    *dymnstypes.DymName
		newOwner           string
		duration           int64
		wantErr            bool
		wantErrContains    string
		wantFirstYearPrice int64
		wantExtendPrice    int64
	}{
		{
			name:               "pass - new registration, 1 letter, 1 year",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:               "pass - new registration, empty buyer",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           "",
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:               "pass - new registration, 1 letter, 2 years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "pass - new registration, 1 letter, N years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:               "pass - new registration, 6 letters, 1 year",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    0,
		},
		{
			name:               "pass - new registration, 6 letters, 2 years",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "pass - new registration, 5+ letters, N years",
			dymName:            "my-name",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:    "pass - extends same owner, 1 letter, 1 year",
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
			name:    "pass - extends same owner, 1 letter, 2 years",
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
			name:    "pass - extends same owner, 1 letter, N years",
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
			name:    "pass - extends same owner, 6 letters, 1 year",
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
			name:    "pass - extends same owner, 6 letters, 2 years",
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
			name:    "pass - extends same owner, 5+ letters, N years",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
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
			name:    "pass - extends expired, same owner, 5+ letters, 2 years",
			dymName: "my-name",
			existingDymName: &dymnstypes.DymName{
				Name:       "my-name",
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
			name:    "pass - extends expired, empty buyer, treat as take over",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           "",
			duration:           2,
			wantFirstYearPrice: 5,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:    "pass - take-over, 1 letter, 1 year",
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
			name:    "pass - take-over, 1 letter, 3 years",
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
			name:    "pass - take-over, 6 letters, 1 year",
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
			name:    "pass - take-over, 6 letters, 3 years",
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
			name:               "pass - new registration, 2 letters",
			dymName:            "aa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price2L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "pass - new registration, 3 letters",
			dymName:            "aaa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price3L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "pass - new registration, 4 letters",
			dymName:            "less",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price4L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "pass - new registration, 5 letters",
			dymName:            "angel",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:            "fail - reject invalid Dym-Name",
			dymName:         "-a-",
			existingDymName: nil,
			newOwner:        buyerA,
			duration:        2,
			wantErr:         true,
			wantErrContains: "invalid dym name",
		},
		{
			name:            "fail - reject invalid duration",
			dymName:         "a",
			existingDymName: nil,
			newOwner:        buyerA,
			duration:        0,
			wantErr:         true,
			wantErrContains: "duration must be at least 1 year",
		},
		{
			name:    "fail - reject estimation for Dym-Name owned by another and not expired",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:        buyerA,
			duration:        1,
			wantErr:         true,
			wantErrContains: "you are not the owner",
		},
		{
			name:    "fail - reject estimation for Dym-Name owned by another and not expired, empty buyer",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:        "",
			duration:        1,
			wantErr:         true,
			wantErrContains: "you are not the owner",
		},
		{
			name:    "pass - allow estimation for take-over, regardless grace period",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1, // still in grace period
			},
			newOwner:           buyerA,
			duration:           3,
			wantErr:            false,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "pass - allow estimation for take-over, regardless grace period, empty buyer",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1, // still in grace period
			},
			newOwner:           "",
			duration:           3,
			wantErr:            false,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			require.Positive(t, dk.MiscParams(ctx).GracePeriodDuration, "bad setup, must have grace period")

			if tt.existingDymName != nil {
				err := dk.SetDymName(ctx, *tt.existingDymName)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.EstimateRegisterName(sdk.WrapSDKContext(ctx), &dymnstypes.EstimateRegisterNameRequest{
				Name:     tt.dymName,
				Duration: tt.duration,
				Owner:    tt.newOwner,
			})

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.Equal(t, tt.wantFirstYearPrice, resp.FirstYearPrice.Amount.Int64())
			require.Equal(t, tt.wantExtendPrice, resp.ExtendPrice.Amount.Int64())
			require.Equal(
				t,
				tt.wantFirstYearPrice+tt.wantExtendPrice,
				resp.TotalPrice.Amount.Int64(),
				"total price must be equals to sum of first year and extend price",
			)
			require.Equal(t, denom, resp.FirstYearPrice.Denom)
			require.Equal(t, denom, resp.ExtendPrice.Denom)
			require.Equal(t, denom, resp.TotalPrice.Denom)
		})
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)
		resp, err := queryServer.EstimateRegisterName(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})
}

func Test_queryServer_ReverseResolveAddress(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"
	const nimChainId = "nim_1122-1"

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		moduleParams := dk.GetParams(ctx)
		moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: chainId,
				Aliases: []string{"dym"},
			},
			{
				ChainId: nimChainId,
				Aliases: []string{"nim"},
			},
		}
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		// add rollapp to enable hex address reverse mapping for this chain
		rk.SetRollapp(ctx, rollapptypes.Rollapp{
			RollappId: nimChainId,
			Owner:     testAddr(0).bech32(),
		})

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.ReverseResolveAddress(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("reject empty request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.ReverseResolveAddress(sdk.WrapSDKContext(ctx), &dymnstypes.ReverseResolveAddressRequest{
			Addresses: []string{},
		})
		require.Error(t, err)
		require.Nil(t, resp)
	})

	ownerAcc := testAddr(1)
	anotherAcc := testAddr(2)
	icaAcc := testICAddr(3)
	cosmosAcc := testAddr(4)
	//goland:noinspection SpellCheckingInspection
	bitcoinAddr := "12higDjoCCNXSA95xZMWUdPvXNmkAduhWv"

	tests := []struct {
		name               string
		dymNames           []dymnstypes.DymName
		addresses          []string
		workingChainId     string
		wantErr            bool
		wantErrContains    string
		wantResult         map[string]dymnstypes.ReverseResolveAddressResult
		wantWorkingChainId string
	}{
		{
			name: "pass - mixed addresses type",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
				},
			},
			addresses: []string{ownerAcc.bech32(), ownerAcc.hexStr()},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{"a@dym"},
				},
				ownerAcc.hexStr(): {
					Candidates: []string{"a@dym"},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - ignore bad input address",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
				},
			},
			addresses: []string{ownerAcc.bech32(), ownerAcc.hexStr(), "@", string(make([]rune, 1000))},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{"a@dym"},
				},
				ownerAcc.hexStr(): {
					Candidates: []string{"a@dym"},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name:      "pass - working =-chain-id if empty is host-chain",
			dymNames:  nil,
			addresses: []string{ownerAcc.bech32()},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - multiple addresses",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "another.account",
							Value:   anotherAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmosAcc.bech32(),
						},
					},
				},
			},
			addresses: []string{
				ownerAcc.bech32(),
				anotherAcc.bech32(),
				cosmosAcc.bech32(),
			},
			workingChainId: chainId,
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{"a@dym"},
				},
				anotherAcc.bech32(): {
					Candidates: []string{"another.account.a@dym"},
				},
				cosmosAcc.bech32(): {
					Candidates: []string{},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - only find on matching chain",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "another.account",
							Value:   anotherAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmosAcc.bech32(),
						},
					},
				},
			},
			addresses: []string{
				ownerAcc.bech32(),
				anotherAcc.bech32(),
				cosmosAcc.bech32(),
			},
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{},
				},
				anotherAcc.bech32(): {
					Candidates: []string{},
				},
				cosmosAcc.bech32(): {
					Candidates: []string{"a@cosmoshub-4"},
				},
			},
			wantWorkingChainId: "cosmoshub-4",
		},
		{
			name: "pass - multi-level sub-name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "a.b.c.d",
							Value:   ownerAcc.bech32(),
						},
					},
				},
			},
			addresses:      []string{ownerAcc.bech32()},
			workingChainId: chainId,
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{"a@dym", "a.b.c.d.a@dym"},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - each address match multiple result",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "a.b.c.d",
							Value:   ownerAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "another",
							Value:   anotherAcc.bech32(),
						},
					},
				},
				{
					Name:       "b",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "e.f.g.h",
							Value:   ownerAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "another",
							Value:   anotherAcc.bech32(),
						},
					},
				},
				{
					Name:       "c",
					Owner:      anotherAcc.bech32(),
					Controller: anotherAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "d",
							Value:   ownerAcc.bech32(),
						},
					},
				},
			},
			addresses: []string{ownerAcc.bech32(), anotherAcc.hexStr()},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				ownerAcc.bech32(): {
					Candidates: []string{"a@dym", "b@dym", "d.c@dym", "a.b.c.d.a@dym", "e.f.g.h.b@dym"},
				},
				anotherAcc.hexStr(): {
					Candidates: []string{"c@dym", "another.a@dym", "another.b@dym"},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - alias not mapped if no alias",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmosAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: nimChainId,
							Path:    "",
							Value:   ownerAcc.bech32(),
						},
					},
				},
			},
			addresses:      []string{cosmosAcc.bech32(), ownerAcc.bech32()},
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				cosmosAcc.bech32(): {
					Candidates: []string{"a@cosmoshub-4"},
				},
				ownerAcc.bech32(): {
					Candidates: []string{},
				},
			},
			wantWorkingChainId: "cosmoshub-4",
		},
		{
			name: "pass - support ICA address",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "",
							Path:    "ica",
							Value:   icaAcc.bech32(),
						},
					},
				},
				{
					Name:       "ica",
					Owner:      icaAcc.bech32(),
					Controller: icaAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
				},
			},
			addresses: []string{icaAcc.bech32(), icaAcc.hexStr()},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				icaAcc.bech32(): {
					Candidates: []string{"ica@dym", "ica.a@dym"},
				},
				icaAcc.hexStr(): {
					Candidates: []string{"ica@dym", "ica.a@dym"},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - chains neither host-chain nor RollApp should not support reverse-resolve hex address",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "cosmoshub-4",
							Path:    "",
							Value:   cosmosAcc.bech32(),
						},
					},
				},
			},
			addresses:      []string{cosmosAcc.bech32(), cosmosAcc.hexStr()},
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				cosmosAcc.bech32(): {
					Candidates: []string{"a@cosmoshub-4"},
				},
				cosmosAcc.hexStr(): {
					Candidates: []string{},
				},
			},
			wantWorkingChainId: "cosmoshub-4",
		},
		{
			name: "pass - returns empty for non-reverse-resolvable address",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
				},
			},
			addresses: []string{anotherAcc.bech32(), anotherAcc.hexStr()},
			wantErr:   false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				anotherAcc.bech32(): {
					Candidates: []string{},
				},
				anotherAcc.hexStr(): {
					Candidates: []string{},
				},
			},
			wantWorkingChainId: chainId,
		},
		{
			name: "pass - reverse-resolve bitcoin address (neither bech32 nor hex address)",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerAcc.bech32(),
					Controller: ownerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "bitcoin",
							Value:   bitcoinAddr,
						},
					},
				},
			},
			addresses:      []string{bitcoinAddr},
			workingChainId: "bitcoin",
			wantErr:        false,
			wantResult: map[string]dymnstypes.ReverseResolveAddressResult{
				bitcoinAddr: {
					Candidates: []string{"a@bitcoin"},
				},
			},
			wantWorkingChainId: "bitcoin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.ReverseResolveAddress(sdk.WrapSDKContext(ctx), &dymnstypes.ReverseResolveAddressRequest{
				Addresses:      tt.addresses,
				WorkingChainId: tt.workingChainId,
			})

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if !reflect.DeepEqual(tt.wantResult, resp.Result) {
				t.Errorf("got = %v, want %v", resp.Result, tt.wantResult)
			}
			require.Equal(t, tt.wantWorkingChainId, resp.WorkingChainId)
		})
	}
}

func Test_queryServer_TranslateAliasOrChainIdToChainId(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	registeredAlias := map[string]string{
		chainId:      "dym",
		"nim_1122-1": "nim",
	}

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		moduleParams := dk.GetParams(ctx)
		for chainIdHasAlias, alias := range registeredAlias {
			moduleParams.Chains.AliasesOfChainIds = append(moduleParams.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
				ChainId: chainIdHasAlias,
				Aliases: []string{alias},
			})
		}
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.TranslateAliasOrChainIdToChainId(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("reject empty request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.TranslateAliasOrChainIdToChainId(sdk.WrapSDKContext(ctx), &dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest{
			AliasOrChainId: "",
		})
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("resolve alias to chain-id", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		for chainIdHasAlias, alias := range registeredAlias {
			resp, err := queryServer.TranslateAliasOrChainIdToChainId(sdk.WrapSDKContext(ctx), &dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest{
				AliasOrChainId: alias,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, chainIdHasAlias, resp.ChainId)
		}
	})

	t.Run("resolve chain-id to chain-id", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		for chainIdHasAlias := range registeredAlias {
			resp, err := queryServer.TranslateAliasOrChainIdToChainId(sdk.WrapSDKContext(ctx), &dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest{
				AliasOrChainId: chainIdHasAlias,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, chainIdHasAlias, resp.ChainId)
		}
	})

	t.Run("treat unknown-chain-id as chain-id", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		for _, unknownChainId := range []string{
			"aaa", "bbb", "ccc", "ddd", "eee",
		} {
			resp, err := queryServer.TranslateAliasOrChainIdToChainId(sdk.WrapSDKContext(ctx), &dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest{
				AliasOrChainId: unknownChainId,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, unknownChainId, resp.ChainId)
		}
	})
}

func Test_queryServer_BuyOrderById(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.BuyOrderById(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	buyerA := testAddr(1).bech32()

	tests := []struct {
		name            string
		buyOrders       []dymnstypes.BuyOrder
		buyOrderId      string
		wantErr         bool
		wantErrContains string
		wantOffer       dymnstypes.BuyOrder
	}{
		{
			name: "pass - can return",
			buyOrders: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			buyOrderId: "101",
			wantErr:    false,
			wantOffer: dymnstypes.BuyOrder{
				Id:         "101",
				GoodsId:    "a",
				Type:       dymnstypes.NameOrder,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			},
		},
		{
			name: "pass - can return among multiple records",
			buyOrders: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
			buyOrderId: "102",
			wantErr:    false,
			wantOffer: dymnstypes.BuyOrder{
				Id:         "102",
				GoodsId:    "a",
				Type:       dymnstypes.NameOrder,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(2),
			},
		},
		{
			name: "fail - return error if not found",
			buyOrders: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
			},
			buyOrderId:      "103",
			wantErr:         true,
			wantErrContains: "buy order not found",
		},
		{
			name: "fail - reject empty offer-id",
			buyOrders: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			buyOrderId:      "",
			wantErr:         true,
			wantErrContains: "invalid Buy-Order ID",
		},
		{
			name: "fail - reject bad offer-id",
			buyOrders: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			buyOrderId:      "@",
			wantErr:         true,
			wantErrContains: "invalid Buy-Order ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			for _, offer := range tt.buyOrders {
				err := dk.SetBuyOrder(ctx, offer)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.BuyOrderById(sdk.WrapSDKContext(ctx), &dymnstypes.QueryBuyOrderByIdRequest{
				Id: tt.buyOrderId,
			})

			if tt.wantErr {
				requireErrorContains(t, err, tt.wantErrContains)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.Equal(t, tt.wantOffer, resp.BuyOrder)
		})
	}
}

func Test_queryServer_BuyOrdersPlacedByAccount(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.BuyOrdersPlacedByAccount(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	buyerA := testAddr(1).bech32()
	anotherA := testAddr(2).bech32()

	tests := []struct {
		name       string
		dymNames   []dymnstypes.DymName
		offers     []dymnstypes.BuyOrder
		account    string
		wantErr    bool
		wantOffers []dymnstypes.BuyOrder
	}{
		{
			name: "pass - can return",
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			account: buyerA,
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
		},
		{
			name: "pass - returns all records made by account",
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "c",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA, // should exclude this
					OfferPrice: dymnsutils.TestCoin(3),
				},
				{
					Id:         "104",
					GoodsId:    "d",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(4),
				},
			},
			account: buyerA,
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "104",
					GoodsId:    "d",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(4),
				},
			},
		},
		{
			name: "pass - return empty if no match",
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			account:    buyerA,
			wantErr:    false,
			wantOffers: nil,
		},
		{
			name: "fail - reject empty account",
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			account: "",
			wantErr: true,
		},
		{
			name: "fail - reject bad account",
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			account: "0x1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				err := dk.SetDymName(ctx, dymName)
				require.NoError(t, err)
			}

			for _, offer := range tt.offers {
				err := dk.SetBuyOrder(ctx, offer)
				require.NoError(t, err)

				err = dk.AddReverseMappingGoodsIdToBuyOrder(ctx, offer.GoodsId, offer.Type, offer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, offer.Buyer, offer.Id)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.BuyOrdersPlacedByAccount(sdk.WrapSDKContext(ctx), &dymnstypes.QueryBuyOrdersPlacedByAccountRequest{
				Account: tt.account,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			sort.Slice(tt.wantOffers, func(i, j int) bool {
				return tt.wantOffers[i].Id < tt.wantOffers[j].Id
			})
			sort.Slice(resp.BuyOrders, func(i, j int) bool {
				return resp.BuyOrders[i].Id < resp.BuyOrders[j].Id
			})

			require.Equal(t, tt.wantOffers, resp.BuyOrders)
		})
	}
}

func Test_queryServer_BuyOrdersByDymName(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.BuyOrdersByDymName(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	buyerA := testAddr(1).bech32()
	ownerA := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()

	tests := []struct {
		name       string
		dymNames   []dymnstypes.DymName
		offers     []dymnstypes.BuyOrder
		dymName    string
		wantErr    bool
		wantOffers []dymnstypes.BuyOrder
	}{
		{
			name: "pass - can return",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			dymName: "a",
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
		},
		{
			name: "pass - returns all records by corresponding Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
				{
					Name:       "b",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
			dymName: "a",
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
			},
		},
		{
			name: "pass - return empty if no match",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
				{
					Name:       "b",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
			dymName:    "c",
			wantErr:    false,
			wantOffers: nil,
		},
		{
			name: "fail - reject empty Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			dymName: "",
			wantErr: true,
		},
		{
			name: "fail - reject bad Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			dymName: "@",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				err := dk.SetDymName(ctx, dymName)
				require.NoError(t, err)
			}

			for _, offer := range tt.offers {
				err := dk.SetBuyOrder(ctx, offer)
				require.NoError(t, err)

				err = dk.AddReverseMappingGoodsIdToBuyOrder(ctx, offer.GoodsId, offer.Type, offer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, offer.Buyer, offer.Id)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.BuyOrdersByDymName(sdk.WrapSDKContext(ctx), &dymnstypes.QueryBuyOrdersByDymNameRequest{
				Name: tt.dymName,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			sort.Slice(tt.wantOffers, func(i, j int) bool {
				return tt.wantOffers[i].Id < tt.wantOffers[j].Id
			})
			sort.Slice(resp.BuyOrders, func(i, j int) bool {
				return resp.BuyOrders[i].Id < resp.BuyOrders[j].Id
			})

			require.Equal(t, tt.wantOffers, resp.BuyOrders)
		})
	}
}

func Test_queryServer_BuyOrdersOfDymNamesOwnedByAccount(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, ctx
	}

	t.Run("reject nil request", func(t *testing.T) {
		dk, ctx := setupTest()
		queryServer := dymnskeeper.NewQueryServerImpl(dk)

		resp, err := queryServer.BuyOrdersOfDymNamesOwnedByAccount(sdk.WrapSDKContext(ctx), nil)
		require.Error(t, err)
		require.Nil(t, resp)
	})

	buyerA := testAddr(1).bech32()
	ownerA := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()

	tests := []struct {
		name       string
		dymNames   []dymnstypes.DymName
		offers     []dymnstypes.BuyOrder
		owner      string
		wantErr    bool
		wantOffers []dymnstypes.BuyOrder
	}{
		{
			name: "pass - can return",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			owner:   ownerA,
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
		},
		{
			name: "pass - returns all records by corresponding Dym-Name",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
				{
					Name:       "b",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
				{
					Name:       "c",
					Owner:      anotherA,
					Controller: anotherA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
				{
					Id:         "104",
					GoodsId:    "c",
					Type:       dymnstypes.NameOrder,
					Buyer:      ownerA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
			owner:   ownerA,
			wantErr: false,
			wantOffers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
		},
		{
			name: "pass - return empty if no match",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
				{
					Name:       "b",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				{
					Id:         "102",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(2),
				},
				{
					Id:         "103",
					GoodsId:    "b",
					Type:       dymnstypes.NameOrder,
					Buyer:      anotherA,
					OfferPrice: dymnsutils.TestCoin(3),
				},
			},
			owner:      anotherA,
			wantErr:    false,
			wantOffers: []dymnstypes.BuyOrder{},
		},
		{
			name: "fail - reject empty account",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			owner:   "",
			wantErr: true,
		},
		{
			name: "fail - reject bad account",
			dymNames: []dymnstypes.DymName{
				{
					Name:       "a",
					Owner:      ownerA,
					Controller: ownerA,
					ExpireAt:   now.Unix() + 1,
				},
			},
			offers: []dymnstypes.BuyOrder{
				{
					Id:         "101",
					GoodsId:    "a",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
			},
			owner:   "0x1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			for _, offer := range tt.offers {
				err := dk.SetBuyOrder(ctx, offer)
				require.NoError(t, err)

				err = dk.AddReverseMappingGoodsIdToBuyOrder(ctx, offer.GoodsId, offer.Type, offer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, offer.Buyer, offer.Id)
				require.NoError(t, err)
			}

			queryServer := dymnskeeper.NewQueryServerImpl(dk)

			resp, err := queryServer.BuyOrdersOfDymNamesOwnedByAccount(sdk.WrapSDKContext(ctx), &dymnstypes.QueryBuyOrdersOfDymNamesOwnedByAccountRequest{
				Account: tt.owner,
			})

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			sort.Slice(tt.wantOffers, func(i, j int) bool {
				return tt.wantOffers[i].Id < tt.wantOffers[j].Id
			})
			sort.Slice(resp.BuyOrders, func(i, j int) bool {
				return resp.BuyOrders[i].Id < resp.BuyOrders[j].Id
			})

			require.Equal(t, tt.wantOffers, resp.BuyOrders)
		})
	}
}

func (s *KeeperTestSuite) Test_queryServer_Alias() {
	rollApp1 := newRollApp("rollapp_1-1").WithOwner(testAddr(1).bech32()).WithAlias("one")
	rollApp2 := newRollApp("rollapp_2-2").WithAlias("two")

	tests := []struct {
		name               string
		rollApps           []rollapp
		preRunFunc         func(s *KeeperTestSuite)
		req                *dymnstypes.QueryAliasRequest
		wantErr            bool
		wantErrContains    string
		wantChainId        string
		wantFoundSellOrder bool
		wantBuyOrderIds    []string
	}{
		{
			name:     "pass - can return alias of mapping in params",
			rollApps: nil,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return params
				})
			},
			req:                &dymnstypes.QueryAliasRequest{Alias: "dym"},
			wantErr:            false,
			wantChainId:        "dymension_1100-1",
			wantFoundSellOrder: false,
			wantBuyOrderIds:    nil,
		},
		{
			name:     "pass - can return alias of mapping in params, even if there are multiple mappings",
			rollApps: nil,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym", "dymension"},
						},
						{
							ChainId: "blumbus_111-1",
							Aliases: []string{"blumbus"},
						},
					}
					return params
				})
			},
			req:                &dymnstypes.QueryAliasRequest{Alias: "dymension"},
			wantErr:            false,
			wantChainId:        "dymension_1100-1",
			wantFoundSellOrder: false,
			wantBuyOrderIds:    nil,
		},
		{
			name:     "pass - if alias is mapped both in params and RollApp alias, priority params",
			rollApps: nil,
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return params
				})

				s.persistRollApp(
					*newRollApp("dym_1-1").WithAlias("dym"),
				)

				s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, "dym_1-1"))
			},
			req:                &dymnstypes.QueryAliasRequest{Alias: "dym"},
			wantErr:            false,
			wantChainId:        "dymension_1100-1",
			wantFoundSellOrder: false,
			wantBuyOrderIds:    nil,
		},
		{
			name:     "pass - returns Sell/Buy orders info if alias is mapped in RollApp alias",
			rollApps: nil,
			preRunFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*rollApp1)
				s.persistRollApp(*rollApp2)

				aliasSellOrder := s.newAliasSellOrder(rollApp1.alias).WithMinPrice(100).Build()

				err := s.dymNsKeeper.SetSellOrder(s.ctx, aliasSellOrder)
				s.Require().NoError(err)

				aliasBuyOrder1 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 1)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOrder1)

				aliasBuyOrder2 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 2)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOrder2)
			},
			req:                &dymnstypes.QueryAliasRequest{Alias: rollApp1.alias},
			wantErr:            false,
			wantChainId:        rollApp1.rollAppId,
			wantFoundSellOrder: true,
			wantBuyOrderIds: []string{
				dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 1),
				dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 2),
			},
		},
		{
			name:     "pass - if alias is mapped both in params and RollApp alias, priority params, ignore Sell/Buy orders",
			rollApps: nil,
			preRunFunc: func(s *KeeperTestSuite) {
				rollApp := newRollApp("dym_3-3").WithOwner(testAddr(3).bech32()).WithAlias("dym")
				s.persistRollApp(*rollApp)

				aliasSellOrder := s.newAliasSellOrder("dym").WithMinPrice(100).Build()
				aliasBuyOrder := s.newAliasBuyOrder(rollApp1.owner, "dym", rollApp1.rollAppId).Build()

				err := s.dymNsKeeper.SetSellOrder(s.ctx, aliasSellOrder)
				s.Require().NoError(err)
				_, err = s.dymNsKeeper.InsertNewBuyOrder(s.ctx, aliasBuyOrder)
				s.Require().NoError(err)

				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return params
				})
			},
			req:                &dymnstypes.QueryAliasRequest{Alias: "dym"},
			wantErr:            false,
			wantChainId:        "dymension_1100-1",
			wantFoundSellOrder: false,
			wantBuyOrderIds:    nil,
		},
		{
			name:            "fail - reject nil request",
			req:             nil,
			wantErr:         true,
			wantErrContains: "invalid request",
		},
		{
			name:            "fail - returns error if not found",
			req:             &dymnstypes.QueryAliasRequest{Alias: "void"},
			wantErr:         true,
			wantErrContains: "not found",
		},
		{
			name:            "fail - if input was detected as a chain-id returns as not found",
			req:             &dymnstypes.QueryAliasRequest{Alias: s.chainId},
			wantErr:         true,
			wantErrContains: "not found",
		},
		{
			name: "fail - if input was detected as a RollApp ID returns as not found",
			preRunFunc: func(s *KeeperTestSuite) {
				s.persistRollApp(*rollApp1)
			},
			req:             &dymnstypes.QueryAliasRequest{Alias: rollApp1.rollAppId},
			wantErr:         true,
			wantErrContains: "not found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			for _, rollApp := range tt.rollApps {
				s.persistRollApp(rollApp)
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			resp, err := dymnskeeper.NewQueryServerImpl(s.dymNsKeeper).Alias(sdk.WrapSDKContext(s.ctx), tt.req)

			if tt.wantErr {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)

			s.Equal(tt.wantChainId, resp.ChainId)
			s.Equal(tt.wantFoundSellOrder, resp.FoundSellOrder)

			if len(tt.wantBuyOrderIds) == 0 {
				s.Empty(resp.BuyOrderIds)
			} else {
				sort.Strings(tt.wantBuyOrderIds)
				sort.Strings(resp.BuyOrderIds)
				s.Equal(tt.wantBuyOrderIds, resp.BuyOrderIds)
			}
		})
	}
}

func (s *KeeperTestSuite) Test_queryServer_BuyOrdersByAlias() {
	rollApp1 := *newRollApp("rollapp_1-1").WithOwner(testAddr(1).bech32()).WithAlias("one")
	rollApp2 := *newRollApp("rollapp_2-2").WithAlias("two")

	aliasBuyOrder1 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
		WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 1)).
		Build()
	aliasBuyOrder2 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
		WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 2)).
		Build()

	tests := []struct {
		name            string
		rollapp         []rollapp
		buyOrders       []dymnstypes.BuyOrder
		preRunFunc      func(s *KeeperTestSuite)
		req             *dymnstypes.QueryBuyOrdersByAliasRequest
		wantErr         bool
		wantErrContains string
		wantBuyOrderIds []string
	}{
		{
			name:      "pass - can buy buy orders of the alias",
			rollapp:   []rollapp{rollApp1, rollApp2},
			buyOrders: []dymnstypes.BuyOrder{aliasBuyOrder1, aliasBuyOrder2},
			req:       &dymnstypes.QueryBuyOrdersByAliasRequest{Alias: rollApp1.alias},
			wantErr:   false,
			wantBuyOrderIds: []string{
				aliasBuyOrder1.Id, aliasBuyOrder2.Id,
			},
		},
		{
			name:      "pass - returns empty if alias present in params as alias of a chain",
			rollapp:   []rollapp{rollApp1, rollApp2},
			buyOrders: []dymnstypes.BuyOrder{aliasBuyOrder1, aliasBuyOrder2},
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "some-chain",
							Aliases: []string{rollApp1.alias},
						},
					}
					return params
				})
			},
			req:             &dymnstypes.QueryBuyOrdersByAliasRequest{Alias: rollApp1.alias},
			wantErr:         false,
			wantBuyOrderIds: nil,
		},
		{
			name:      "pass - returns empty if alias present in params as a chain-id",
			rollapp:   []rollapp{rollApp1, rollApp2},
			buyOrders: []dymnstypes.BuyOrder{aliasBuyOrder1, aliasBuyOrder2},
			preRunFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: rollApp1.alias,
							Aliases: nil,
						},
					}
					return params
				})
			},
			req:             &dymnstypes.QueryBuyOrdersByAliasRequest{Alias: rollApp1.alias},
			wantErr:         false,
			wantBuyOrderIds: nil,
		},
		{
			name:            "fail - reject nil request",
			req:             nil,
			wantErr:         true,
			wantErrContains: "invalid request",
			wantBuyOrderIds: nil,
		},
		{
			name:            "fail - reject bad alias",
			req:             &dymnstypes.QueryBuyOrdersByAliasRequest{Alias: "@@@"},
			wantErr:         true,
			wantErrContains: "invalid alias",
			wantBuyOrderIds: nil,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			for _, rollapp := range tt.rollapp {
				s.persistRollApp(rollapp)
			}
			for _, offer := range tt.buyOrders {
				s.setBuyOrderWithFunctionsAfter(offer)
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			resp, err := dymnskeeper.NewQueryServerImpl(s.dymNsKeeper).BuyOrdersByAlias(sdk.WrapSDKContext(s.ctx), tt.req)

			if tt.wantErr {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)

			if len(tt.wantBuyOrderIds) == 0 {
				s.Empty(resp.BuyOrders)
			} else {
				var responseBuyOrderIds []string
				for _, offer := range resp.BuyOrders {
					responseBuyOrderIds = append(responseBuyOrderIds, offer.Id)
				}

				sort.Strings(tt.wantBuyOrderIds)
				sort.Strings(responseBuyOrderIds)

				s.Equal(tt.wantBuyOrderIds, responseBuyOrderIds)
			}
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_queryServer_BuyOffersOfAliasesLinkedToRollApp() {
	rollApp1 := *newRollApp("rollapp_1-1").WithAlias("one")
	rollApp2 := *newRollApp("rollapp_2-2").WithAlias("another")

	aliasBuyOffer1_ra1_alias1 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
		WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 1)).
		Build()
	aliasBuyOffer2_ra1_alias1 := s.newAliasBuyOrder(rollApp2.owner, rollApp1.alias, rollApp2.rollAppId).
		WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 2)).
		Build()

	tests := []struct {
		name            string
		rollapp         []rollapp
		buyOffers       []dymnstypes.BuyOrder
		preRunFunc      func(s *KeeperTestSuite)
		req             *dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest
		wantErr         bool
		wantErrContains string
		wantBuyOrderIds []string
	}{
		{
			name:       "pass - can returns if there is Buy Order",
			rollapp:    []rollapp{rollApp1, rollApp2},
			buyOffers:  []dymnstypes.BuyOrder{aliasBuyOffer1_ra1_alias1},
			preRunFunc: nil,
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: rollApp1.rollAppId,
			},
			wantErr:         false,
			wantBuyOrderIds: []string{aliasBuyOffer1_ra1_alias1.Id},
		},
		{
			name:       "pass - can empty if there is No buy order",
			rollapp:    []rollapp{rollApp1, rollApp2},
			buyOffers:  nil,
			preRunFunc: nil,
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: rollApp2.rollAppId,
			},
			wantErr:         false,
			wantBuyOrderIds: []string{},
		},
		{
			name:       "pass - return multiple if there are many buy orders",
			rollapp:    []rollapp{rollApp1, rollApp2},
			buyOffers:  []dymnstypes.BuyOrder{aliasBuyOffer1_ra1_alias1, aliasBuyOffer2_ra1_alias1},
			preRunFunc: nil,
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: rollApp1.rollAppId,
			},
			wantErr: false,
			wantBuyOrderIds: []string{
				aliasBuyOffer1_ra1_alias1.Id, aliasBuyOffer2_ra1_alias1.Id,
			},
		},
		{
			name:      "pass - return multiple if there are buy orders associated with different aliases of the same RollApp",
			rollapp:   []rollapp{rollApp1, rollApp2},
			buyOffers: []dymnstypes.BuyOrder{aliasBuyOffer1_ra1_alias1, aliasBuyOffer2_ra1_alias1},
			preRunFunc: func(s *KeeperTestSuite) {
				const alias2 = "more"
				const alias3 = "alias"

				s.Require().NoError(
					s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, alias2),
				)
				s.Require().NoError(
					s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, alias3),
				)
				s.requireRollApp(rollApp1.rollAppId).HasAlias(
					rollApp1.alias, alias2, alias3,
				)

				aliasBuyOffer3_ra1_alias2 := s.newAliasBuyOrder(rollApp2.owner, alias2, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 3)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOffer3_ra1_alias2)

				aliasBuyOffer4_ra1_alias3 := s.newAliasBuyOrder(rollApp2.owner, alias3, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 4)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOffer4_ra1_alias3)
			},
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: rollApp1.rollAppId,
			},
			wantErr: false,
			wantBuyOrderIds: []string{
				aliasBuyOffer1_ra1_alias1.Id,
				aliasBuyOffer2_ra1_alias1.Id,
				dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 3),
				dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 4),
			},
		},
		{
			name:      "pass - exclude buy orders of aliases which presents in params as chain-alias",
			rollapp:   []rollapp{rollApp1, rollApp2},
			buyOffers: []dymnstypes.BuyOrder{aliasBuyOffer1_ra1_alias1, aliasBuyOffer2_ra1_alias1},
			preRunFunc: func(s *KeeperTestSuite) {
				const alias2 = "more"
				const alias3 = "alias"

				s.Require().NoError(
					s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, alias2),
				)
				s.Require().NoError(
					s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp1.rollAppId, alias3),
				)
				s.requireRollApp(rollApp1.rollAppId).HasAlias(
					rollApp1.alias, alias2, alias3,
				)

				aliasBuyOffer3_ra1_alias2 := s.newAliasBuyOrder(rollApp2.owner, alias2, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 3)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOffer3_ra1_alias2)

				aliasBuyOffer4_ra1_alias3 := s.newAliasBuyOrder(rollApp2.owner, alias3, rollApp2.rollAppId).
					WithID(dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 4)).
					Build()
				s.setBuyOrderWithFunctionsAfter(aliasBuyOffer4_ra1_alias3)

				s.updateModuleParams(func(params dymnstypes.Params) dymnstypes.Params {
					params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "some-chain",
							Aliases: []string{alias2},
						},
					}
					return params
				})
			},
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: rollApp1.rollAppId,
			},
			wantErr: false,
			wantBuyOrderIds: []string{
				aliasBuyOffer1_ra1_alias1.Id,
				aliasBuyOffer2_ra1_alias1.Id,
				dymnstypes.CreateBuyOrderId(dymnstypes.AliasOrder, 4),
			},
		},
		{
			name:            "fail - reject nil request",
			req:             nil,
			wantErr:         true,
			wantErrContains: "invalid request",
			wantBuyOrderIds: nil,
		},
		{
			name: "fail - reject bad RollApp ID",
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: "@@@",
			},
			wantErr:         true,
			wantErrContains: "invalid RollApp ID",
			wantBuyOrderIds: nil,
		},
		{
			name:    "fail - reject if RollApp does not exists",
			rollapp: []rollapp{rollApp1},
			req: &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
				RollappId: "nah_0-0",
			},
			wantErr:         true,
			wantErrContains: "RollApp not found",
			wantBuyOrderIds: nil,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			for _, rollapp := range tt.rollapp {
				s.persistRollApp(rollapp)
			}
			for _, offer := range tt.buyOffers {
				s.setBuyOrderWithFunctionsAfter(offer)
			}

			if tt.preRunFunc != nil {
				tt.preRunFunc(s)
			}

			resp, err := dymnskeeper.NewQueryServerImpl(s.dymNsKeeper).BuyOrdersOfAliasesLinkedToRollApp(sdk.WrapSDKContext(s.ctx), tt.req)

			if tt.wantErr {
				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)

			if len(tt.wantBuyOrderIds) == 0 {
				s.Empty(resp.BuyOrders)
			} else {
				var responseBuyOrderIds []string
				for _, offer := range resp.BuyOrders {
					responseBuyOrderIds = append(responseBuyOrderIds, offer.Id)
				}

				sort.Strings(tt.wantBuyOrderIds)
				sort.Strings(responseBuyOrderIds)

				s.Equal(tt.wantBuyOrderIds, responseBuyOrderIds)
			}
		})
	}
}
