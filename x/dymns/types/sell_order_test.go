package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestSellOrder_GetIdentity(t *testing.T) {
	nameSo := &SellOrder{
		GoodsId:  "my-name",
		Type:     NameOrder,
		ExpireAt: 1234,
	}
	require.Equal(t, "my-name|1|1234", nameSo.GetIdentity())
	aliasSo := &SellOrder{
		GoodsId:  "alias",
		Type:     AliasOrder,
		ExpireAt: 1234,
	}
	require.Equal(t, "alias|2|1234", aliasSo.GetIdentity())
}

func TestSellOrder_HasSetSellPrice(t *testing.T) {
	require.False(t, (&SellOrder{
		SellPrice: nil,
	}).HasSetSellPrice())
	require.False(t, (&SellOrder{
		SellPrice: &sdk.Coin{},
	}).HasSetSellPrice())
	require.False(t, (&SellOrder{
		SellPrice: dymnsutils.TestCoinP(0),
	}).HasSetSellPrice())
	require.True(t, (&SellOrder{
		SellPrice: dymnsutils.TestCoinP(1),
	}).HasSetSellPrice())
}

func TestSellOrder_HasExpiredAtCtx(t *testing.T) {
	const epoch int64 = 2
	ctx := sdk.Context{}.WithBlockTime(time.Unix(2, 0))
	require.True(t, (&SellOrder{
		ExpireAt: epoch - 1,
	}).HasExpiredAtCtx(ctx))
	require.False(t, (&SellOrder{
		ExpireAt: epoch + 1,
	}).HasExpiredAtCtx(ctx))
	require.False(t, (&SellOrder{
		ExpireAt: epoch,
	}).HasExpiredAtCtx(ctx), "SO expires after expires at")
}

func TestSellOrder_HasExpired(t *testing.T) {
	const epoch int64 = 2
	require.True(t, (&SellOrder{
		ExpireAt: epoch - 1,
	}).HasExpired(epoch))
	require.False(t, (&SellOrder{
		ExpireAt: epoch + 1,
	}).HasExpired(epoch))
	require.False(t, (&SellOrder{
		ExpireAt: epoch,
	}).HasExpired(epoch), "SO expires after expires at")
}

func TestSellOrder_HasFinished(t *testing.T) {
	oneCoin := dymnsutils.TestCoin(1)
	threeCoin := dymnsutils.TestCoin(3)
	zeroCoin := dymnsutils.TestCoin(0)

	now := time.Now().UTC()

	tests := []struct {
		name         string
		expireAt     int64
		sellPrice    *sdk.Coin
		highestBid   *SellOrderBid
		wantFinished bool
	}{
		{
			name:         "expired, without sell-price, without bid",
			expireAt:     now.Unix() - 1,
			sellPrice:    &zeroCoin,
			highestBid:   nil,
			wantFinished: true,
		},
		{
			name:         "expired, without sell-price, without bid",
			expireAt:     now.Unix() - 1,
			sellPrice:    nil,
			highestBid:   nil,
			wantFinished: true,
		},
		{
			name:         "expired, + sell-price, without bid",
			expireAt:     now.Unix() - 1,
			sellPrice:    &threeCoin,
			highestBid:   nil,
			wantFinished: true,
		},
		{
			name:      "expired, + sell-price, + bid (under sell-price)",
			expireAt:  now.Unix() - 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  oneCoin,
			},
			wantFinished: true,
		},
		{
			name:      "expired, + sell-price, + bid (= sell-price)",
			expireAt:  now.Unix() - 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  threeCoin,
			},
			wantFinished: true,
		},
		{
			name:         "not expired, without sell-price, without bid",
			expireAt:     now.Unix() + 1,
			sellPrice:    &zeroCoin,
			highestBid:   nil,
			wantFinished: false,
		},
		{
			name:         "not expired, without sell-price, without bid",
			expireAt:     now.Unix() + 1,
			sellPrice:    nil,
			highestBid:   nil,
			wantFinished: false,
		},
		{
			name:         "not expired, + sell-price, without bid",
			expireAt:     now.Unix() + 1,
			sellPrice:    &threeCoin,
			highestBid:   nil,
			wantFinished: false,
		},
		{
			name:      "not expired, + sell-price, + bid (under sell-price)",
			expireAt:  now.Unix() + 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  oneCoin,
			},
			wantFinished: false,
		},
		{
			name:      "not expired, + sell-price, + bid (= sell-price)",
			expireAt:  now.Unix() + 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  threeCoin,
			},
			wantFinished: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SellOrder{
				GoodsId:    "a",
				ExpireAt:   tt.expireAt,
				MinPrice:   oneCoin,
				SellPrice:  tt.sellPrice,
				HighestBid: tt.highestBid,
			}

			for _, orderType := range []OrderType{NameOrder, AliasOrder} {
				m.Type = orderType
				require.Equal(t, tt.wantFinished, m.HasFinishedAtCtx(
					sdk.Context{}.WithBlockTime(now),
				))
				require.Equal(t, tt.wantFinished, m.HasFinished(now.Unix()))
			}
		})
	}
}

func TestSellOrder_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*SellOrder)(nil)
		require.Error(t, m.Validate())
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		_type           OrderType
		expireAt        int64
		minPrice        sdk.Coin
		sellPrice       *sdk.Coin
		highestBid      *SellOrderBid
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid sell order",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(1),
			},
		},
		{
			name:      "pass - (Alias) valid sell order",
			dymName:   "alias",
			_type:     AliasOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(1),
				Params: []string{"rollapp_1-1"},
			},
		},
		{
			name:      "fail - (Alias) reject invalid bid",
			dymName:   "alias",
			_type:     AliasOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(1),
				Params: nil, // empty
			},
			wantErr:         true,
			wantErrContains: "SO highest bid is invalid",
		},
		{
			name:      "pass - (Name) valid sell order without bid",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
		},
		{
			name:      "pass - (Alias) valid sell order without bid",
			dymName:   "alias",
			_type:     AliasOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
		},
		{
			name:      "pass - (Name) valid sell order without setting sell price",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: nil,
		},
		{
			name:      "pass - (Alias) valid sell order without setting sell price",
			dymName:   "alias",
			_type:     AliasOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: nil,
		},
		{
			name:            "fail - (Name) reject empty name",
			dymName:         "",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of SO is empty",
		},
		{
			name:            "fail - (Alias) reject empty alias",
			dymName:         "",
			_type:           AliasOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "alias of SO is empty",
		},
		{
			name:            "fail - reject unknown type",
			dymName:         "goods",
			_type:           OrderType_OT_UNKNOWN,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "invalid SO type",
		},
		{
			name:            "fail - (Name) reject bad name",
			dymName:         "-my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of SO is not a valid dym name",
		},
		{
			name:            "fail - (Alias) reject bad alias",
			dymName:         "bad-alias",
			_type:           AliasOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "alias of SO is not a valid alias",
		},
		{
			name:            "fail - empty time",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        0,
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "SO expiry is empty",
		},
		{
			name:            "fail - min price is zero",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "fail - min price is empty",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        sdk.Coin{},
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "fail - min price is negative",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(-1),
			wantErr:         true,
			wantErrContains: "SO min price is negative",
		},
		{
			name:     "fail - min price is invalid",
			dymName:  "my-name",
			_type:    NameOrder,
			expireAt: time.Now().Unix(),
			minPrice: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "SO min price is invalid",
		},
		{
			name:            "fail - sell price is negative",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			sellPrice:       dymnsutils.TestCoinP(-1),
			wantErr:         true,
			wantErrContains: "SO sell price is negative",
		},
		{
			name:     "fail - sell price is invalid",
			dymName:  "my-name",
			_type:    NameOrder,
			expireAt: time.Now().Unix(),
			minPrice: dymnsutils.TestCoin(1),
			sellPrice: &sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "SO sell price is invalid",
		},
		{
			name:            "fail - sell price is less than min price",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(2),
			sellPrice:       dymnsutils.TestCoinP(1),
			wantErr:         true,
			wantErrContains: "SO sell price is less than min price",
		},
		{
			name:            "fail - sell price denom must match min price denom",
			dymName:         "my-name",
			_type:           NameOrder,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			sellPrice:       dymnsutils.TestCoin2P(sdk.NewInt64Coin("u"+params.BaseDenom, 2)),
			wantErr:         true,
			wantErrContains: "SO sell price denom is different from min price denom",
		},
		{
			name:      "fail - invalid highest bid",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			highestBid: &SellOrderBid{
				Bidder: "0x1",
				Price:  dymnsutils.TestCoin(1),
			},
			wantErr:         true,
			wantErrContains: "SO bidder is not a valid bech32 account address",
		},
		{
			name:      "fail - highest bid < min price",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(2),
			sellPrice: dymnsutils.TestCoinP(3),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(1),
			},
			wantErr:         true,
			wantErrContains: "SO highest bid price is less than min price",
		},
		{
			name:      "fail - highest bid > sell price",
			dymName:   "my-name",
			_type:     NameOrder,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(2),
			sellPrice: dymnsutils.TestCoinP(3),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(4),
			},
			wantErr:         true,
			wantErrContains: "SO sell price is less than highest bid price",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SellOrder{
				GoodsId:    tt.dymName,
				Type:       tt._type,
				ExpireAt:   tt.expireAt,
				MinPrice:   tt.minPrice,
				SellPrice:  tt.sellPrice,
				HighestBid: tt.highestBid,
			}

			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSellOrderBid_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*SellOrderBid)(nil)
		require.Error(t, m.Validate(NameOrder))
		require.Error(t, m.Validate(AliasOrder))
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		bidder          string
		price           sdk.Coin
		params          []string
		orderType       OrderType
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid sell order bid",
			bidder:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			params:    nil,
			orderType: NameOrder,
			price:     dymnsutils.TestCoin(1),
		},
		{
			name:      "pass - (Alias) valid sell order bid",
			bidder:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			params:    []string{"rollapp_1-1"},
			orderType: AliasOrder,
			price:     dymnsutils.TestCoin(1),
		},
		{
			name:            "fail - empty bidder",
			bidder:          "",
			price:           dymnsutils.TestCoin(1),
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bidder is empty",
		},
		{
			name:            "fail - bad bidder",
			bidder:          "0x1",
			price:           dymnsutils.TestCoin(1),
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bidder is not a valid bech32 account address",
		},
		{
			name:            "fail - zero price",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           dymnsutils.TestCoin(0),
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bid price is zero",
		},
		{
			name:            "fail - zero price",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           sdk.Coin{},
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bid price is zero",
		},
		{
			name:   "fail - negative price",
			bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bid price is negative",
		},
		{
			name:   "fail - invalid price",
			bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			params:          nil,
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "SO bid price is invalid",
		},
		{
			name:            "fail - (Name) bad params",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           dymnsutils.TestCoin(1),
			params:          []string{"non-empty"},
			orderType:       NameOrder,
			wantErr:         true,
			wantErrContains: "not accept order params for order type",
		},
		{
			name:            "fail - (Alias) bad params",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           dymnsutils.TestCoin(1),
			params:          nil,
			orderType:       AliasOrder,
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SellOrderBid{
				Bidder: tt.bidder,
				Price:  tt.price,
				Params: tt.params,
			}
			err := m.Validate(tt.orderType)
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestHistoricalSellOrders_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*HistoricalSellOrders)(nil)
		require.Error(t, m.Validate())
	})

	tests := []struct {
		name            string
		sellOrders      []SellOrder
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "pass - valid",
			sellOrders: []SellOrder{
				{
					GoodsId:   "my-name",
					Type:      NameOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "my-name",
					Type:     NameOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
		},
		{
			name: "pass - valid",
			sellOrders: []SellOrder{
				{
					GoodsId:   "alias",
					Type:      AliasOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "alias",
					Type:     AliasOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
		},
		{
			name:       "pass - allow empty",
			sellOrders: []SellOrder{},
		},
		{
			name: "fail - reject if SO element is invalid",
			sellOrders: []SellOrder{
				{
					GoodsId:  "a",
					Type:     NameOrder,
					ExpireAt: 1,
					MinPrice: dymnsutils.TestCoin(0), // invalid
				},
				{
					GoodsId:  "a",
					Type:     NameOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name: "fail - reject if duplicated SO",
			sellOrders: []SellOrder{
				{
					GoodsId:   "my-name",
					Type:      NameOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "my-name",
					Type:     NameOrder,
					ExpireAt: 1,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SO is not unique",
		},
		{
			name: "fail - reject if duplicated SO",
			sellOrders: []SellOrder{
				{
					GoodsId:   "alias",
					Type:      AliasOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "alias",
					Type:     AliasOrder,
					ExpireAt: 1,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SO is not unique",
		},
		{
			name: "fail - reject if SO element has different goods ID",
			sellOrders: []SellOrder{
				{
					GoodsId:   "aaa",
					Type:      NameOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "bbb",
					Type:     NameOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SOs have different goods ID",
		},
		{
			name: "fail - reject if SO element has different goods ID",
			sellOrders: []SellOrder{
				{
					GoodsId:   "aaa",
					Type:      AliasOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "bbb",
					Type:     AliasOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SOs have different goods ID",
		},
		{
			name: "fail - reject if SO element has mixed order types",
			sellOrders: []SellOrder{
				{
					GoodsId:   "aaa",
					Type:      NameOrder,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					GoodsId:  "aaa",
					Type:     AliasOrder,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SOs have different order type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HistoricalSellOrders{
				SellOrders: tt.sellOrders,
			}

			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSellOrder_GetSdkEvent(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		event := SellOrder{
			GoodsId:   "a",
			Type:      NameOrder,
			ExpireAt:  123456,
			MinPrice:  dymnsutils.TestCoin(1),
			SellPrice: dymnsutils.TestCoinP(3),
			HighestBid: &SellOrderBid{
				Bidder: "d",
				Price:  dymnsutils.TestCoin(2),
			},
		}.GetSdkEvent("action-name")
		requireEventEquals(t, event,
			EventTypeSellOrder,
			AttributeKeySoGoodsId, "a",
			AttributeKeySoType, NameOrder.FriendlyString(),
			AttributeKeySoExpiryEpoch, "123456",
			AttributeKeySoMinPrice, "1"+params.BaseDenom,
			AttributeKeySoSellPrice, "3"+params.BaseDenom,
			AttributeKeySoHighestBidder, "d",
			AttributeKeySoHighestBidPrice, "2"+params.BaseDenom,
			AttributeKeySoActionName, "action-name",
		)
	})

	t.Run("SO type alias", func(t *testing.T) {
		event := SellOrder{
			GoodsId:   "a",
			Type:      AliasOrder,
			ExpireAt:  123456,
			MinPrice:  dymnsutils.TestCoin(1),
			SellPrice: dymnsutils.TestCoinP(3),
			HighestBid: &SellOrderBid{
				Bidder: "d",
				Price:  dymnsutils.TestCoin(2),
			},
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeSellOrder, event.Type)
		require.Len(t, event.Attributes, 8)
		require.Equal(t, AttributeKeySoType, event.Attributes[1].Key)
		require.Equal(t, AliasOrder.FriendlyString(), event.Attributes[1].Value)
	})

	t.Run("no sell-price", func(t *testing.T) {
		event := SellOrder{
			GoodsId:  "a",
			Type:     NameOrder,
			ExpireAt: 123456,
			MinPrice: dymnsutils.TestCoin(1),
			HighestBid: &SellOrderBid{
				Bidder: "d",
				Price:  dymnsutils.TestCoin(2),
			},
		}.GetSdkEvent("action-name")
		requireEventEquals(t, event,
			EventTypeSellOrder,
			AttributeKeySoGoodsId, "a",
			AttributeKeySoType, NameOrder.FriendlyString(),
			AttributeKeySoExpiryEpoch, "123456",
			AttributeKeySoMinPrice, "1"+params.BaseDenom,
			AttributeKeySoSellPrice, "0"+params.BaseDenom,
			AttributeKeySoHighestBidder, "d",
			AttributeKeySoHighestBidPrice, "2"+params.BaseDenom,
			AttributeKeySoActionName, "action-name",
		)
	})

	t.Run("no highest bid", func(t *testing.T) {
		event := SellOrder{
			GoodsId:   "a",
			Type:      NameOrder,
			ExpireAt:  123456,
			MinPrice:  dymnsutils.TestCoin(1),
			SellPrice: dymnsutils.TestCoinP(3),
		}.GetSdkEvent("action-name")
		requireEventEquals(t, event,
			EventTypeSellOrder,
			AttributeKeySoGoodsId, "a",
			AttributeKeySoType, NameOrder.FriendlyString(),
			AttributeKeySoExpiryEpoch, "123456",
			AttributeKeySoMinPrice, "1"+params.BaseDenom,
			AttributeKeySoSellPrice, "3"+params.BaseDenom,
			AttributeKeySoHighestBidder, "",
			AttributeKeySoHighestBidPrice, "0"+params.BaseDenom,
			AttributeKeySoActionName, "action-name",
		)
	})
}

func TestActiveSellOrdersExpiration_Validate(t *testing.T) {
	tests := []struct {
		name            string
		records         []ActiveSellOrdersExpirationRecord
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "pass",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "b", ExpireAt: 1},
			},
			wantErr: false,
		},
		{
			name: "fail - name must be unique",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "a", ExpireAt: 1},
			},
			wantErr:         true,
			wantErrContains: "active SO is not unique",
		},
		{
			name: "pass - expire at can be duplicated",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "b", ExpireAt: 2},
			},
			wantErr: false,
		},
		{
			name: "fail - expire at must be > 0",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 0}, {GoodsId: "b", ExpireAt: -1},
			},
			wantErr:         true,
			wantErrContains: "active SO expiry is empty",
		},
		{
			name: "fail - must be sorted",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1}, {GoodsId: "a", ExpireAt: 1},
			},
			wantErr:         true,
			wantErrContains: "active SO names are not sorted",
		},
		{
			name:    "pass - empty list",
			records: []ActiveSellOrdersExpirationRecord{},
			wantErr: false,
		},
		{
			name:    "pass - nil list",
			records: nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ActiveSellOrdersExpiration{
				Records: tt.records,
			}

			err := m.Validate()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestActiveSellOrdersExpiration_Sort(t *testing.T) {
	tests := []struct {
		name    string
		records []ActiveSellOrdersExpirationRecord
		want    []ActiveSellOrdersExpirationRecord
	}{
		{
			name: "can sort",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 2}, {GoodsId: "a", ExpireAt: 2},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "b", ExpireAt: 2},
			},
		},
		{
			name: "sort by name asc",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1}, {GoodsId: "a", ExpireAt: 2}, {GoodsId: "c", ExpireAt: 3},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 3},
			},
		},
		{
			name: "can sort one",
			records: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2},
			},
		},
		{
			name:    "empty list",
			records: []ActiveSellOrdersExpirationRecord{},
			want:    []ActiveSellOrdersExpirationRecord{},
		},
		{
			name:    "nil list",
			records: nil,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ActiveSellOrdersExpiration{
				Records: tt.records,
			}

			m.Sort()

			require.Equal(t, tt.want, m.Records)
		})
	}
}

func TestActiveSellOrdersExpiration_Add(t *testing.T) {
	tests := []struct {
		name      string
		existing  []ActiveSellOrdersExpirationRecord
		addName   string
		addExpiry int64
		want      []ActiveSellOrdersExpirationRecord
	}{
		{
			name:      "can add",
			existing:  []ActiveSellOrdersExpirationRecord{{GoodsId: "a", ExpireAt: 1}},
			addName:   "b",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 2},
			},
		},
		{
			name:      "add will perform sort",
			existing:  []ActiveSellOrdersExpirationRecord{{GoodsId: "b", ExpireAt: 1}},
			addName:   "a",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 2}, {GoodsId: "b", ExpireAt: 1},
			},
		},
		{
			name:      "add can override existing",
			existing:  []ActiveSellOrdersExpirationRecord{{GoodsId: "b", ExpireAt: 1}},
			addName:   "b",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 2},
			},
		},
		{
			name: "add can override existing",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1}, {GoodsId: "d", ExpireAt: 1},
			},
			addName:   "c",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 2}, {GoodsId: "d", ExpireAt: 1},
			},
		},
		{
			name:      "can add to nil",
			existing:  nil,
			addName:   "a",
			addExpiry: 1,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1},
			},
		},
		{
			name:      "can add to empty",
			existing:  []ActiveSellOrdersExpirationRecord{},
			addName:   "a",
			addExpiry: 1,
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ActiveSellOrdersExpiration{
				Records: tt.existing,
			}
			m.Add(tt.addName, tt.addExpiry)

			require.Equal(t, tt.want, m.Records)
		})
	}
}

func TestActiveSellOrdersExpiration_Remove(t *testing.T) {
	tests := []struct {
		name       string
		existing   []ActiveSellOrdersExpirationRecord
		removeName string
		want       []ActiveSellOrdersExpirationRecord
	}{
		{
			name: "can remove",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1},
			},
			removeName: "a",
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1},
			},
		},
		{
			name: "remove the last one",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1},
			},
			removeName: "a",
			want:       []ActiveSellOrdersExpirationRecord{},
		},
		{
			name: "remove in head",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1},
			},
			removeName: "a",
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1},
			},
		},
		{
			name: "remove in middle",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1},
			},
			removeName: "b",
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1},
			},
		},
		{
			name: "remove in tails",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1}, {GoodsId: "c", ExpireAt: 1},
			},
			removeName: "c",
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "a", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1},
			},
		},
		{
			name: "remove keep order",
			existing: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "c", ExpireAt: 1}, {GoodsId: "b", ExpireAt: 1}, {GoodsId: "a", ExpireAt: 1},
			},
			removeName: "b",
			want: []ActiveSellOrdersExpirationRecord{
				{GoodsId: "c", ExpireAt: 1}, {GoodsId: "a", ExpireAt: 1},
			},
		},
		{
			name:       "can remove from nil",
			existing:   nil,
			removeName: "a",
			want:       nil,
		},
		{
			name:       "can remove from empty",
			existing:   []ActiveSellOrdersExpirationRecord{},
			removeName: "a",
			want:       []ActiveSellOrdersExpirationRecord{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ActiveSellOrdersExpiration{
				Records: tt.existing,
			}
			m.Remove(tt.removeName)

			require.Equal(t, tt.want, m.Records)
		})
	}
}

func requireEventEquals(t *testing.T, event sdk.Event, wantType string, wantAttributePairs ...string) {
	require.NotNil(t, event)
	require.True(t, len(wantAttributePairs)%2 == 0, "size of expected attr pairs must be even")
	require.Equal(t, wantType, event.Type)
	require.Len(t, event.Attributes, len(wantAttributePairs)/2)
	for i := 0; i < len(wantAttributePairs); i += 2 {
		require.Equal(t, wantAttributePairs[i], event.Attributes[i/2].Key)
		require.Equal(t, wantAttributePairs[i+1], event.Attributes[i/2].Value)
	}
}
