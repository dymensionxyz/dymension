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
	m := &SellOrder{
		Name:     "aabb",
		ExpireAt: 1234,
	}
	require.Equal(t, "aabb|1234", m.GetIdentity())
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
		name       string
		expireAt   int64
		sellPrice  *sdk.Coin
		highestBid *SellOrderBid
		want       bool
	}{
		{
			name:       "expired, without sell-price, without bid",
			expireAt:   now.Unix() - 1,
			sellPrice:  &zeroCoin,
			highestBid: nil,
			want:       true,
		},
		{
			name:       "expired, without sell-price, without bid",
			expireAt:   now.Unix() - 1,
			sellPrice:  nil,
			highestBid: nil,
			want:       true,
		},
		{
			name:       "expired, + sell-price, without bid",
			expireAt:   now.Unix() - 1,
			sellPrice:  &threeCoin,
			highestBid: nil,
			want:       true,
		},
		{
			name:      "expired, + sell-price, + bid (under sell-price)",
			expireAt:  now.Unix() - 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  oneCoin,
			},
			want: true,
		},
		{
			name:      "expired, + sell-price, + bid (= sell-price)",
			expireAt:  now.Unix() - 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  threeCoin,
			},
			want: true,
		},
		{
			name:       "not expired, without sell-price, without bid",
			expireAt:   now.Unix() + 1,
			sellPrice:  &zeroCoin,
			highestBid: nil,
			want:       false,
		},
		{
			name:       "not expired, without sell-price, without bid",
			expireAt:   now.Unix() + 1,
			sellPrice:  nil,
			highestBid: nil,
			want:       false,
		},
		{
			name:       "not expired, + sell-price, without bid",
			expireAt:   now.Unix() + 1,
			sellPrice:  &threeCoin,
			highestBid: nil,
			want:       false,
		},
		{
			name:      "not expired, + sell-price, + bid (under sell-price)",
			expireAt:  now.Unix() + 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  oneCoin,
			},
			want: false,
		},
		{
			name:      "not expired, + sell-price, + bid (= sell-price)",
			expireAt:  now.Unix() + 1,
			sellPrice: &threeCoin,
			highestBid: &SellOrderBid{
				Bidder: "x",
				Price:  threeCoin,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SellOrder{
				Name:       "a",
				ExpireAt:   tt.expireAt,
				MinPrice:   oneCoin,
				SellPrice:  tt.sellPrice,
				HighestBid: tt.highestBid,
			}

			require.Equal(t, tt.want, m.HasFinishedAtCtx(
				sdk.Context{}.WithBlockTime(now),
			))
			require.Equal(t, tt.want, m.HasFinished(now.Unix()))
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
		_type           MarketOrderType
		expireAt        int64
		minPrice        sdk.Coin
		sellPrice       *sdk.Coin
		highestBid      *SellOrderBid
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - valid sell order",
			dymName:   "my-name",
			_type:     MarketOrderType_MOT_DYM_NAME,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			highestBid: &SellOrderBid{
				Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				Price:  dymnsutils.TestCoin(1),
			},
		},
		{
			name:      "pass - valid sell order without bid",
			dymName:   "my-name",
			_type:     MarketOrderType_MOT_DYM_NAME,
			expireAt:  time.Now().Unix(),
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
		},
		{
			name:     "pass - valid sell order without setting sell price",
			dymName:  "my-name",
			_type:    MarketOrderType_MOT_DYM_NAME,
			expireAt: time.Now().Unix(),
			minPrice: dymnsutils.TestCoin(1),
		},
		{
			name:            "fail - empty name",
			dymName:         "",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of SO is empty",
		},
		{
			name:            "fail - type is unknown",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_UNKNOWN,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Sell-Order type must be",
		},
		{
			name:            "fail - type is alias (not yet supported)",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_ALIAS,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Sell-Order type must be",
		},
		{
			name:            "fail - bad name",
			dymName:         "-my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of SO is not a valid dym name",
		},
		{
			name:            "fail - empty time",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        0,
			minPrice:        dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "SO expiry is empty",
		},
		{
			name:            "fail - min price is zero",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "fail - min price is empty",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        sdk.Coin{},
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "fail - min price is negative",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(-1),
			wantErr:         true,
			wantErrContains: "SO min price is negative",
		},
		{
			name:     "fail - min price is invalid",
			dymName:  "my-name",
			_type:    MarketOrderType_MOT_DYM_NAME,
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
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			sellPrice:       dymnsutils.TestCoinP(-1),
			wantErr:         true,
			wantErrContains: "SO sell price is negative",
		},
		{
			name:     "fail - sell price is invalid",
			dymName:  "my-name",
			_type:    MarketOrderType_MOT_DYM_NAME,
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
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(2),
			sellPrice:       dymnsutils.TestCoinP(1),
			wantErr:         true,
			wantErrContains: "SO sell price is less than min price",
		},
		{
			name:            "fail - sell price denom must match min price denom",
			dymName:         "my-name",
			_type:           MarketOrderType_MOT_DYM_NAME,
			expireAt:        time.Now().Unix(),
			minPrice:        dymnsutils.TestCoin(1),
			sellPrice:       dymnsutils.TestCoin2P(sdk.NewInt64Coin("u"+params.BaseDenom, 2)),
			wantErr:         true,
			wantErrContains: "SO sell price denom is different from min price denom",
		},
		{
			name:      "fail - invalid highest bid",
			dymName:   "my-name",
			_type:     MarketOrderType_MOT_DYM_NAME,
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
			_type:     MarketOrderType_MOT_DYM_NAME,
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
			_type:     MarketOrderType_MOT_DYM_NAME,
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
				Name:       tt.dymName,
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
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSellOrderBid_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*SellOrderBid)(nil)
		require.Error(t, m.Validate())
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		bidder          string
		price           sdk.Coin
		wantErr         bool
		wantErrContains string
	}{
		{
			name:   "valid sell order bid",
			bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:  dymnsutils.TestCoin(1),
		},
		{
			name:            "empty bidder",
			bidder:          "",
			price:           dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "SO bidder is empty",
		},
		{
			name:            "bad bidder",
			bidder:          "0x1",
			price:           dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "SO bidder is not a valid bech32 account address",
		},
		{
			name:            "zero price",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "SO bid price is zero",
		},
		{
			name:            "zero price",
			bidder:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price:           sdk.Coin{},
			wantErr:         true,
			wantErrContains: "SO bid price is zero",
		},
		{
			name:   "negative price",
			bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "SO bid price is negative",
		},
		{
			name:   "invalid price",
			bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			price: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "SO bid price is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SellOrderBid{
				Bidder: tt.bidder,
				Price:  tt.price,
			}
			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
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
					Name:      "a",
					Type:      MarketOrderType_MOT_DYM_NAME,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					Name:     "a",
					Type:     MarketOrderType_MOT_DYM_NAME,
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
					Name:     "a",
					Type:     MarketOrderType_MOT_DYM_NAME,
					ExpireAt: 1,
					MinPrice: dymnsutils.TestCoin(0), // invalid
				},
				{
					Name:     "a",
					Type:     MarketOrderType_MOT_DYM_NAME,
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
					Name:      "a",
					Type:      MarketOrderType_MOT_DYM_NAME,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					Name:      "a",
					Type:      MarketOrderType_MOT_DYM_NAME,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SO is not unique",
		},
		{
			name: "fail - reject if SO element has different Dym-Name",
			sellOrders: []SellOrder{
				{
					Name:      "aaa",
					Type:      MarketOrderType_MOT_DYM_NAME,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
					SellPrice: dymnsutils.TestCoinP(1),
				},
				{
					Name:     "bbb",
					Type:     MarketOrderType_MOT_DYM_NAME,
					ExpireAt: 2,
					MinPrice: dymnsutils.TestCoin(1),
				},
			},
			wantErr:         true,
			wantErrContains: "historical SOs have different Dym-Name",
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
			Name:      "a",
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
		require.Len(t, event.Attributes, 7)
		require.Equal(t, AttributeKeySoName, event.Attributes[0].Key)
		require.Equal(t, "a", event.Attributes[0].Value)
		require.Equal(t, AttributeKeySoExpiryEpoch, event.Attributes[1].Key)
		require.Equal(t, "123456", event.Attributes[1].Value)
		require.Equal(t, AttributeKeySoMinPrice, event.Attributes[2].Key)
		require.Equal(t, "1"+params.BaseDenom, event.Attributes[2].Value)
		require.Equal(t, AttributeKeySoSellPrice, event.Attributes[3].Key)
		require.Equal(t, "3"+params.BaseDenom, event.Attributes[3].Value)
		require.Equal(t, AttributeKeySoHighestBidder, event.Attributes[4].Key)
		require.Equal(t, "d", event.Attributes[4].Value)
		require.Equal(t, AttributeKeySoHighestBidPrice, event.Attributes[5].Key)
		require.Equal(t, "2"+params.BaseDenom, event.Attributes[5].Value)
		require.Equal(t, AttributeKeySoActionName, event.Attributes[6].Key)
		require.Equal(t, "action-name", event.Attributes[6].Value)
	})

	t.Run("no sell-price", func(t *testing.T) {
		event := SellOrder{
			Name:     "a",
			ExpireAt: 123456,
			MinPrice: dymnsutils.TestCoin(1),
			HighestBid: &SellOrderBid{
				Bidder: "d",
				Price:  dymnsutils.TestCoin(2),
			},
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeSellOrder, event.Type)
		require.Len(t, event.Attributes, 7)
		require.Equal(t, AttributeKeySoName, event.Attributes[0].Key)
		require.Equal(t, "a", event.Attributes[0].Value)
		require.Equal(t, AttributeKeySoExpiryEpoch, event.Attributes[1].Key)
		require.Equal(t, "123456", event.Attributes[1].Value)
		require.Equal(t, AttributeKeySoMinPrice, event.Attributes[2].Key)
		require.Equal(t, "1"+params.BaseDenom, event.Attributes[2].Value)
		require.Equal(t, AttributeKeySoSellPrice, event.Attributes[3].Key)
		require.Equal(t, "0"+params.BaseDenom, event.Attributes[3].Value)
		require.Equal(t, AttributeKeySoHighestBidder, event.Attributes[4].Key)
		require.Equal(t, "d", event.Attributes[4].Value)
		require.Equal(t, AttributeKeySoHighestBidPrice, event.Attributes[5].Key)
		require.Equal(t, "2"+params.BaseDenom, event.Attributes[5].Value)
		require.Equal(t, AttributeKeySoActionName, event.Attributes[6].Key)
		require.Equal(t, "action-name", event.Attributes[6].Value)
	})
	t.Run("no highest bid", func(t *testing.T) {
		event := SellOrder{
			Name:      "a",
			ExpireAt:  123456,
			MinPrice:  dymnsutils.TestCoin(1),
			SellPrice: dymnsutils.TestCoinP(3),
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeSellOrder, event.Type)
		require.Len(t, event.Attributes, 7)
		require.Equal(t, AttributeKeySoName, event.Attributes[0].Key)
		require.Equal(t, "a", event.Attributes[0].Value)
		require.Equal(t, AttributeKeySoExpiryEpoch, event.Attributes[1].Key)
		require.Equal(t, "123456", event.Attributes[1].Value)
		require.Equal(t, AttributeKeySoMinPrice, event.Attributes[2].Key)
		require.Equal(t, "1"+params.BaseDenom, event.Attributes[2].Value)
		require.Equal(t, AttributeKeySoSellPrice, event.Attributes[3].Key)
		require.Equal(t, "3"+params.BaseDenom, event.Attributes[3].Value)
		require.Equal(t, AttributeKeySoHighestBidder, event.Attributes[4].Key)
		require.Empty(t, event.Attributes[4].Value)
		require.Equal(t, AttributeKeySoHighestBidPrice, event.Attributes[5].Key)
		require.Equal(t, "0"+params.BaseDenom, event.Attributes[5].Value)
		require.Equal(t, AttributeKeySoActionName, event.Attributes[6].Key)
		require.Equal(t, "action-name", event.Attributes[6].Value)
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
				{Name: "a", ExpireAt: 2}, {Name: "b", ExpireAt: 1},
			},
			wantErr: false,
		},
		{
			name: "fail - name must be unique",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2}, {Name: "a", ExpireAt: 1},
			},
			wantErr:         true,
			wantErrContains: "active SO is not unique",
		},
		{
			name: "pass - expire at can be duplicated",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2}, {Name: "b", ExpireAt: 2},
			},
			wantErr: false,
		},
		{
			name: "fail - expire at must be > 0",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 0}, {Name: "b", ExpireAt: -1},
			},
			wantErr:         true,
			wantErrContains: "active SO expiry is empty",
		},
		{
			name: "fail - must be sorted",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1}, {Name: "a", ExpireAt: 1},
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
				{Name: "b", ExpireAt: 2}, {Name: "a", ExpireAt: 2},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2}, {Name: "b", ExpireAt: 2},
			},
		},
		{
			name: "sort by name asc",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1}, {Name: "a", ExpireAt: 2}, {Name: "c", ExpireAt: 3},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2}, {Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 3},
			},
		},
		{
			name: "can sort one",
			records: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2},
			},
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2},
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
			existing:  []ActiveSellOrdersExpirationRecord{{Name: "a", ExpireAt: 1}},
			addName:   "b",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 2},
			},
		},
		{
			name:      "add will perform sort",
			existing:  []ActiveSellOrdersExpirationRecord{{Name: "b", ExpireAt: 1}},
			addName:   "a",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 2}, {Name: "b", ExpireAt: 1},
			},
		},
		{
			name:      "add can override existing",
			existing:  []ActiveSellOrdersExpirationRecord{{Name: "b", ExpireAt: 1}},
			addName:   "b",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 2},
			},
		},
		{
			name: "add can override existing",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 1}, {Name: "d", ExpireAt: 1},
			},
			addName:   "c",
			addExpiry: 2,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 2}, {Name: "d", ExpireAt: 1},
			},
		},
		{
			name:      "can add to nil",
			existing:  nil,
			addName:   "a",
			addExpiry: 1,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1},
			},
		},
		{
			name:      "can add to empty",
			existing:  []ActiveSellOrdersExpirationRecord{},
			addName:   "a",
			addExpiry: 1,
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1},
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
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 1},
			},
			removeName: "a",
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1},
			},
		},
		{
			name: "remove the last one",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1},
			},
			removeName: "a",
			want:       []ActiveSellOrdersExpirationRecord{},
		},
		{
			name: "remove in head",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 1},
			},
			removeName: "a",
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 1},
			},
		},
		{
			name: "remove in middle",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 1},
			},
			removeName: "b",
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "c", ExpireAt: 1},
			},
		},
		{
			name: "remove in tails",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 1}, {Name: "c", ExpireAt: 1},
			},
			removeName: "c",
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "a", ExpireAt: 1}, {Name: "b", ExpireAt: 1},
			},
		},
		{
			name: "remove keep order",
			existing: []ActiveSellOrdersExpirationRecord{
				{Name: "c", ExpireAt: 1}, {Name: "b", ExpireAt: 1}, {Name: "a", ExpireAt: 1},
			},
			removeName: "b",
			want: []ActiveSellOrdersExpirationRecord{
				{Name: "c", ExpireAt: 1}, {Name: "a", ExpireAt: 1},
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
