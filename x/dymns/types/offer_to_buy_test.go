package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestOfferToBuy_HasCounterpartyOfferPrice(t *testing.T) {
	require.False(t, (&OfferToBuy{
		CounterpartyOfferPrice: nil,
	}).HasCounterpartyOfferPrice())
	require.False(t, (&OfferToBuy{
		CounterpartyOfferPrice: &sdk.Coin{},
	}).HasCounterpartyOfferPrice())
	require.False(t, (&OfferToBuy{
		CounterpartyOfferPrice: dymnsutils.TestCoinP(0),
	}).HasCounterpartyOfferPrice())
	require.True(t, (&OfferToBuy{
		CounterpartyOfferPrice: dymnsutils.TestCoinP(1),
	}).HasCounterpartyOfferPrice())
}

func TestOfferToBuy_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*OfferToBuy)(nil)
		require.Error(t, m.Validate())
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name                   string
		offerId                string
		dymName                string
		buyer                  string
		offerPrice             sdk.Coin
		counterpartyOfferPrice *sdk.Coin
		wantErr                bool
		wantErrContains        string
	}{
		{
			name:                   "pass - valid offer",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: nil,
		},
		{
			name:                   "pass - valid offer with counterparty offer price",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(2),
		},
		{
			name:                   "pass - valid offer without counterparty offer price",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: nil,
		},
		{
			name:            "reject - empty offer ID",
			offerId:         "",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "ID of offer is empty",
		},
		{
			name:            "reject - bad offer ID",
			offerId:         "@",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "ID of offer is not a valid offer id",
		},
		{
			name:            "reject - empty name",
			offerId:         "1",
			dymName:         "",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of offer is empty",
		},
		{
			name:            "reject - bad name",
			offerId:         "1",
			dymName:         "@",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "Dym-Name of offer is not a valid dym name",
		},
		{
			name:            "reject - bad buyer",
			offerId:         "1",
			dymName:         "a",
			buyer:           "0x1",
			offerPrice:      dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "reject - offer price is zero",
			offerId:         "1",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "offer price is zero",
		},
		{
			name:            "reject - offer price is empty",
			offerId:         "1",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      sdk.Coin{},
			wantErr:         true,
			wantErrContains: "offer price is zero",
		},
		{
			name:            "reject - offer price is negative",
			offerId:         "1",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:      dymnsutils.TestCoin(-1),
			wantErr:         true,
			wantErrContains: "offer price is negative",
		},
		{
			name:    "reject - offer price is invalid",
			offerId: "1",
			dymName: "a",
			buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "offer price is invalid",
		},
		{
			name:                   "pass - counter-party offer price is zero",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(0),
		},
		{
			name:                   "pass - counter-party offer price is empty",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: &sdk.Coin{},
		},
		{
			name:                   "reject - counter-party offer price is negative",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoinP(-1),
			wantErr:                true,
			wantErrContains:        "counterparty offer price is negative",
		},
		{
			name:       "reject - counter-party offer price is invalid",
			offerId:    "1",
			dymName:    "a",
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
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(1),
			wantErr:                false,
		},
		{
			name:                   "pass - counterparty offer price can be equals to offer price",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(2),
			wantErr:                false,
		},
		{
			name:                   "pass - counterparty offer price can be greater than offer price",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(2),
			counterpartyOfferPrice: dymnsutils.TestCoinP(3),
			wantErr:                false,
		},
		{
			name:                   "reject - counterparty offer price denom must match offer price denom",
			offerId:                "1",
			dymName:                "a",
			buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offerPrice:             dymnsutils.TestCoin(1),
			counterpartyOfferPrice: dymnsutils.TestCoin2P(sdk.NewInt64Coin("u"+params.BaseDenom, 2)),
			wantErr:                true,
			wantErrContains:        "counterparty offer price denom is different from offer price denom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OfferToBuy{
				Id:                     tt.offerId,
				Name:                   tt.dymName,
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
func TestOfferToBuy_GetSdkEvent(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		event := OfferToBuy{
			Id:                     "1",
			Name:                   "a",
			Buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice:             dymnsutils.TestCoin(1),
			CounterpartyOfferPrice: dymnsutils.TestCoinP(2),
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeOfferToBuy, event.Type)
		require.Len(t, event.Attributes, 5)
		require.Equal(t, AttributeKeyOtbId, event.Attributes[0].Key)
		require.Equal(t, "1", event.Attributes[0].Value)
		require.Equal(t, AttributeKeyOtbName, event.Attributes[1].Key)
		require.Equal(t, "a", event.Attributes[1].Value)
		require.Equal(t, AttributeKeyOtbOfferPrice, event.Attributes[2].Key)
		require.Equal(t, "1"+params.BaseDenom, event.Attributes[2].Value)
		require.Equal(t, AttributeKeyOtbCounterpartyOfferPrice, event.Attributes[3].Key)
		require.Equal(t, "2"+params.BaseDenom, event.Attributes[3].Value)
		require.Equal(t, AttributeKeySoActionName, event.Attributes[4].Key)
		require.Equal(t, "action-name", event.Attributes[4].Value)
	})

	t.Run("no counterparty offer price", func(t *testing.T) {
		event := OfferToBuy{
			Id:                     "1",
			Name:                   "a",
			Buyer:                  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice:             dymnsutils.TestCoin(1),
			CounterpartyOfferPrice: nil,
		}.GetSdkEvent("action-name")
		require.NotNil(t, event)
		require.Equal(t, EventTypeOfferToBuy, event.Type)
		require.Len(t, event.Attributes, 5)
		require.Equal(t, AttributeKeyOtbId, event.Attributes[0].Key)
		require.Equal(t, "1", event.Attributes[0].Value)
		require.Equal(t, AttributeKeyOtbName, event.Attributes[1].Key)
		require.Equal(t, "a", event.Attributes[1].Value)
		require.Equal(t, AttributeKeyOtbOfferPrice, event.Attributes[2].Key)
		require.Equal(t, "1"+params.BaseDenom, event.Attributes[2].Value)
		require.Equal(t, AttributeKeyOtbCounterpartyOfferPrice, event.Attributes[3].Key)
		require.Empty(t, event.Attributes[3].Value)
		require.Equal(t, AttributeKeySoActionName, event.Attributes[4].Key)
		require.Equal(t, "action-name", event.Attributes[4].Value)
	})
}
