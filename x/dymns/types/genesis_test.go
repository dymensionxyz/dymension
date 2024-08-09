package types

import (
	"testing"
	"time"

	"github.com/dymensionxyz/dymension/v3/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	defaultGenesis := DefaultGenesis()
	require.NotNil(t, defaultGenesis)
	require.NoError(t, defaultGenesis.Validate())
}

//goland:noinspection SpellCheckingInspection
func TestGenesisState_Validate(t *testing.T) {
	defaultGenesis := DefaultGenesis()
	require.NoError(t, defaultGenesis.Validate())

	t.Run("pass - valid genesis", func(t *testing.T) {
		require.NoError(t, (GenesisState{
			Params: DefaultParams(),
			DymNames: []DymName{
				{
					Name:       "my-name",
					Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					ExpireAt:   time.Now().Unix(),
				},
			},
			SellOrderBids: []SellOrderBid{
				{
					// this bid from a SO of type Dym-Name
					Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					Price: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
				},
				{
					// this bid from a SO of type Alias
					Bidder: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					Price: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
					Params: []string{"rollapp_1-1"},
				},
			},
			BuyOffers: []BuyOffer{
				{
					Id:      "101",
					GoodsId: "my-name",
					Type:    NameOrder,
					Buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					OfferPrice: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
				},
				{
					Id:      "202",
					GoodsId: "alias",
					Type:    AliasOrder,
					Params:  []string{"rollapp_1-1"},
					Buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					OfferPrice: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid params", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: Params{
				Price: DefaultPriceParams(),
				Misc: MiscParams{
					BeginEpochHookIdentifier: "invalid",
				},
			},
		}).Validate())

		require.Error(t, (GenesisState{
			Params: Params{
				Price: PriceParams{},
				Misc: MiscParams{
					BeginEpochHookIdentifier: "invalid",
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid dym names", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: DefaultParams(),
			DymNames: []DymName{
				{
					Name: "",
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid bid", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: DefaultParams(),
			SellOrderBids: []SellOrderBid{
				{
					Bidder: "",
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid buy offer", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: DefaultParams(),
			BuyOffers: []BuyOffer{
				{
					Buyer: "",
				},
			},
		}).Validate())
	})
}
