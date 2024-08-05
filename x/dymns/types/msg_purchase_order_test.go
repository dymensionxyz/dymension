package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestMsgPurchaseOrder_ValidateBasic(t *testing.T) {
	validOffer := dymnsutils.TestCoin(100)

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		goodsId         string
		orderType       MarketOrderType
		offer           sdk.Coin
		buyer           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid",
			goodsId:   "my-name",
			orderType: MarketOrderType_MOT_DYM_NAME,
			offer:     validOffer,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid",
			goodsId:   "alias",
			orderType: MarketOrderType_MOT_ALIAS,
			offer:     validOffer,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - (Name) reject empty name",
			goodsId:         "",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) reject empty alias",
			goodsId:         "",
			orderType:       MarketOrderType_MOT_ALIAS,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) bad name",
			goodsId:         "-my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) bad alias",
			goodsId:         "bad-alias",
			orderType:       MarketOrderType_MOT_ALIAS,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) missing offer",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - (Alias) missing offer",
			goodsId:         "alias",
			orderType:       MarketOrderType_MOT_ALIAS,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - (Name) offer can not be zero",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           dymnsutils.TestCoin(0),
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:            "fail - (Alias) offer can not be zero",
			goodsId:         "alias",
			orderType:       MarketOrderType_MOT_ALIAS,
			offer:           dymnsutils.TestCoin(0),
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:      "fail - offer can not be negative",
			goodsId:   "my-name",
			orderType: MarketOrderType_MOT_DYM_NAME,
			offer: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - missing buyer",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           validOffer,
			buyer:           "",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid buyer",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - buyer must be dym1",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			offer:           validOffer,
			buyer:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - reject unknown order type",
			goodsId:         "goods",
			orderType:       MarketOrderType_MOT_UNKNOWN,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid order type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPurchaseOrder{
				GoodsId:   tt.goodsId,
				OrderType: tt.orderType,
				Offer:     tt.offer,
				Buyer:     tt.buyer,
			}

			err := m.ValidateBasic()
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
