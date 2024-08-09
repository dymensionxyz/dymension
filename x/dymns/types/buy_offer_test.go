package types

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestBuyOrder_HasCounterpartyOfferPrice(t *testing.T) {
	require.False(t, (&BuyOrder{
		CounterpartyOfferPrice: nil,
	}).HasCounterpartyOfferPrice())
	require.False(t, (&BuyOrder{
		CounterpartyOfferPrice: &sdk.Coin{},
	}).HasCounterpartyOfferPrice())
	require.False(t, (&BuyOrder{
		CounterpartyOfferPrice: dymnsutils.TestCoinP(0),
	}).HasCounterpartyOfferPrice())
	require.True(t, (&BuyOrder{
		CounterpartyOfferPrice: dymnsutils.TestCoinP(1),
	}).HasCounterpartyOfferPrice())
}

func TestBuyOrder_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*BuyOrder)(nil)
		require.Error(t, m.Validate())
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                   string
		orderId                string
		assetId                string
		assetType              AssetType
		params                 []string
		buyer                  string
		offerPrice             sdk.Coin
		counterpartyOfferPrice *sdk.Coin
		wantErr                bool
		wantErrContains        string
	}{
		{
			name:                   "pass - (Name) valid offer",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: nil,
		},
		{
			name:                   "pass - (Alias) valid offer",
			orderId:                "201",
			assetId:                "alias",
			assetType:              TypeAlias,
			params:                 []string{"rollapp_1-1"},
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: nil,
		},
		{
			name:                   "pass - valid offer with counterparty offer price",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(2),
		},
		{
			name:                   "pass - valid offer without counterparty offer price",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: nil,
		},
		{
			name:            "fail - empty offer ID",
			orderId:         "",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "ID of offer is empty",
		},
		{
			name:            "fail - offer ID prefix not match type, case Dym-Name",
			orderId:         CreateBuyOrderId(TypeAlias, 1),
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "mismatch type of Buy-Order ID prefix and type",
		},
		{
			name:            "fail - offer ID prefix not match type, case Alias",
			orderId:         CreateBuyOrderId(TypeName, 1),
			assetId:         "my-name",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "mismatch type of Buy-Order ID prefix and type",
		},
		{
			name:            "fail - bad offer ID",
			orderId:         "@",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "ID of offer is not a valid offer id",
		},
		{
			name:            "fail - (Name) empty name",
			orderId:         "101",
			assetId:         "",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of offer is empty",
		},
		{
			name:            "fail - (Alias) empty alias",
			orderId:         "201",
			assetId:         "",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "alias of offer is empty",
		},
		{
			name:            "fail - (Name) bad name",
			orderId:         "101",
			assetId:         "@",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of offer is not a valid dym name",
		},
		{
			name:            "fail - (Alias) bad name",
			orderId:         "201",
			assetId:         "bad-alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "alias of offer is not a valid alias",
		},
		{
			name:            "fail - (Name) reject non-empty params",
			orderId:         "101",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          []string{"non-empty"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "not accept order params for asset type",
		},
		{
			name:            "fail - (Alias) reject empty params",
			orderId:         "201",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID",
		},
		{
			name:            "fail - (Alias) reject bad params",
			orderId:         "201",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"@chain"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format",
		},
		{
			name:            "fail - bad buyer",
			orderId:         "101",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "0x1",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - offer price is zero",
			orderId:         "101",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "offer price is zero",
		},
		{
			name:            "fail - offer price is empty",
			orderId:         "101",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      sdk.Coin{},
			wantErr:         true,
			wantErrContains: "offer price is zero",
		},
		{
			name:            "fail - offer price is negative",
			orderId:         "101",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(-1),
			wantErr:         true,
			wantErrContains: "offer price is negative",
		},
		{
			name:      "fail - offer price is invalid",
			orderId:   "101",
			assetId:   "my-name",
			assetType: TypeName,
			params:    nil,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "offer price is invalid",
		},
		{
			name:                   "pass - counter-party offer price is zero",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(0),
		},
		{
			name:                   "pass - counter-party offer price is empty",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: &sdk.Coin{},
		},
		{
			name:                   "fail - counter-party offer price is negative",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(-1),
			wantErr:                true,
			wantErrContains:        "counterparty offer price is negative",
		},
		{
			name:       "fail - counter-party offer price is invalid",
			orderId:    "101",
			assetId:    "my-name",
			assetType:  TypeName,
			params:     nil,
			buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice: dymnsutils.TestCoin(1),
			counterpartyOfferPrice: &sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "counterparty offer price is invalid",
		},
		{
			name:                   "pass - counterparty offer price can be less than offer price",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(1),
			wantErr:                false,
		},
		{
			name:                   "pass - counterparty offer price can be equals to offer price",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(2),
			wantErr:                false,
		},
		{
			name:                   "pass - counterparty offer price can be greater than offer price",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(3),
			wantErr:                false,
		},
		{
			name:                   "fail - counterparty offer price denom must match offer price denom",
			orderId:                "101",
			assetId:                "my-name",
			assetType:              TypeName,
			params:                 nil,
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoin2P(sdk.NewInt64Coin("u"+params.BaseDenom, 2)),
			wantErr:                true,
			wantErrContains:        "counterparty offer price denom is different from offer price denom",
		},
		{
			name:            "fail - reject unknown asset type",
			orderId:         "101",
			assetId:         "asset",
			assetType:       AssetType_AT_UNKNOWN,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "invalid asset type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BuyOrder{
				Id:                     tt.orderId,
				AssetId:                tt.assetId,
				AssetType:              tt.assetType,
				Params:                 tt.params,
				Buyer:                  tt.buyer,
				OfferPrice:             tt.offerPrice,
				CounterpartyOfferPrice: tt.counterpartyOfferPrice,
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

//goland:noinspection SpellCheckingInspection
func TestBuyOrder_GetSdkEvent(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		event := BuyOrder{
			Id:                     "1",
			AssetId:                "a",
			AssetType:              TypeName,
			Buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice:             dymnsutils.TestCoin(1),
			CounterpartyOfferPrice: dymnsutils.TestCoinP(2),
		}.GetSdkEvent("action-name")
		requireEventEquals(t, event,
			EventTypeBuyOrder,
			AttributeKeyBoId, "1",
			AttributeKeyBoAssetId, "a",
			AttributeKeyBoAssetType, TypeName.FriendlyString(),
			AttributeKeyBoBuyer, "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			AttributeKeyBoOfferPrice, "1"+params.BaseDenom,
			AttributeKeyBoCounterpartyOfferPrice, "2"+params.BaseDenom,
			AttributeKeyBoActionName, "action-name",
		)
	})

	t.Run("BO type Alias", func(t *testing.T) {
		event := BuyOrder{
			Id:                     "1",
			AssetId:                "a",
			AssetType:              TypeAlias,
			Buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice:             dymnsutils.TestCoin(1),
			CounterpartyOfferPrice: dymnsutils.TestCoinP(2),
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeBuyOrder, event.Type)
		require.Len(t, event.Attributes, 7)
		require.Equal(t, AttributeKeyBoAssetType, event.Attributes[2].Key)
		require.Equal(t, TypeAlias.FriendlyString(), event.Attributes[2].Value)
	})

	t.Run("no counterparty offer price", func(t *testing.T) {
		event := BuyOrder{
			Id:                     "1",
			AssetId:                "a",
			AssetType:              TypeName,
			Buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice:             dymnsutils.TestCoin(1),
			CounterpartyOfferPrice: nil,
		}.GetSdkEvent("action-name")
		requireEventEquals(t, event,
			EventTypeBuyOrder,
			AttributeKeyBoId, "1",
			AttributeKeyBoAssetId, "a",
			AttributeKeyBoAssetType, TypeName.FriendlyString(),
			AttributeKeyBoBuyer, "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			AttributeKeyBoOfferPrice, "1"+params.BaseDenom,
			AttributeKeyBoCounterpartyOfferPrice, "",
			AttributeKeyBoActionName, "action-name",
		)
	})
}

func TestIsValidBuyOrderId(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantValid bool
	}{
		{
			name:      "pass - positive number",
			id:        "101",
			wantValid: true,
		},
		{
			name:      "pass - positive number",
			id:        "201",
			wantValid: true,
		},
		{
			name:      "fail - reject zero",
			id:        "100",
			wantValid: false,
		},
		{
			name:      "fail - reject zero",
			id:        "200",
			wantValid: false,
		},
		{
			name:      "fail - reject empty",
			id:        "",
			wantValid: false,
		},
		{
			name:      "fail - reject 1 char",
			id:        "1",
			wantValid: false,
		},
		{
			name:      "fail - reject 2 chars",
			id:        "10",
			wantValid: false,
		},
		{
			name:      "fail - reject negative",
			id:        "10-1",
			wantValid: false,
		},
		{
			name:      "fail - reject negative",
			id:        "20-1",
			wantValid: false,
		},
		{
			name:      "fail - reject non-numeric",
			id:        "10a",
			wantValid: false,
		},
		{
			name:      "fail - reject non-numeric",
			id:        "20a",
			wantValid: false,
		},
		{
			name:      "pass - maximum is max uint64",
			id:        "10" + "18446744073709551615",
			wantValid: true,
		},
		{
			name:      "pass - maximum is max uint64",
			id:        "20" + "18446744073709551615",
			wantValid: true,
		},
		{
			name:      "fail - reject out-of-bound uint64",
			id:        "10" + "18446744073709551616", // max uint64 + 1
			wantValid: false,
		},
		{
			name:      "fail - reject out-of-bound uint64",
			id:        "20" + "18446744073709551616", // max uint64 + 1
			wantValid: false,
		},
		{
			name:      "fail - reject unrecognized prefix",
			id:        "OO1",
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantValid, IsValidBuyOrderId(tt.id))
		})
	}
}

func TestCreateBuyOrderId(t *testing.T) {
	tests := []struct {
		name      string
		_type     AssetType
		i         uint64
		want      string
		wantPanic bool
	}{
		{
			name:  "pass - type Dym-Name",
			_type: TypeName,
			i:     1,
			want:  "101",
		},
		{
			name:  "pass - type Alias",
			_type: TypeAlias,
			i:     1,
			want:  "201",
		},
		{
			name:  "pass - type Dym-Name, max uint64",
			_type: TypeName,
			i:     math.MaxUint64,
			want:  "10" + "18446744073709551615",
		},
		{
			name:  "pass - type Alias, max uint64",
			_type: TypeAlias,
			i:     math.MaxUint64,
			want:  "20" + "18446744073709551615",
		},
		{
			name:      "fail - reject unknown type",
			_type:     AssetType_AT_UNKNOWN,
			i:         1,
			wantPanic: true,
		},
		{
			name:      "fail - reject bad input number",
			_type:     TypeName,
			i:         0,
			wantPanic: true,
		},
		{
			name:      "fail - reject bad input number",
			_type:     TypeAlias,
			i:         0,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() {
					_ = CreateBuyOrderId(tt._type, tt.i)
				})
				return
			}
			got := CreateBuyOrderId(tt._type, tt.i)
			require.Equal(t, tt.want, got)
			require.True(t, IsValidBuyOrderId(got))
		})
	}
}
