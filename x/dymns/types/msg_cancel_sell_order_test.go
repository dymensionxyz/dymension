package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgCancelSellOrder_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		goodsId         string
		orderType       MarketOrderType
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid",
			goodsId:   "my-name",
			orderType: MarketOrderType_MOT_DYM_NAME,
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid",
			goodsId:   "alias",
			orderType: MarketOrderType_MOT_ALIAS,
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - (Name) not allow empty name",
			goodsId:         "",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) not allow empty alias",
			goodsId:         "",
			orderType:       MarketOrderType_MOT_ALIAS,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) not allow invalid name",
			goodsId:         "-my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) not allow invalid alias",
			goodsId:         "bad-alias",
			orderType:       MarketOrderType_MOT_ALIAS,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - invalid owner",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			owner:           "dym1fl48vsnmsdzcv85q5",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - missing owner",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - owner must be dym1",
			goodsId:         "my-name",
			orderType:       MarketOrderType_MOT_DYM_NAME,
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - not supported order type",
			goodsId:         "goods",
			orderType:       MarketOrderType_MOT_UNKNOWN,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid order type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgCancelSellOrder{
				GoodsId:   tt.goodsId,
				OrderType: tt.orderType,
				Owner:     tt.owner,
			}

			err := m.ValidateBasic()
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
