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
			BuyOrders: []BuyOrder{
				{
					Id:        "101",
					AssetId:   "my-name",
					AssetType: TypeName,
					Buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					OfferPrice: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
				},
				{
					Id:        "202",
					AssetId:   "alias",
					AssetType: TypeAlias,
					Params:    []string{"rollapp_1-1"},
					Buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					OfferPrice: sdk.Coin{
						Denom:  params.BaseDenom,
						Amount: sdk.OneInt(),
					},
				},
			},
			AliasesOfRollapps: []AliasesOfChainId{
				{
					ChainId: "rollapp_1-1",
					Aliases: []string{"alias"},
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid params", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: Params{
				Price: DefaultPriceParams(),
				Misc: MiscParams{
					EndEpochHookIdentifier: "invalid",
				},
			},
		}).Validate())

		require.Error(t, (GenesisState{
			Params: Params{
				Price: PriceParams{},
				Misc: MiscParams{
					EndEpochHookIdentifier: "invalid",
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

	t.Run("fail - duplicated dym names", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: DefaultParams(),
			DymNames: []DymName{
				{
					Name:       "my-name",
					Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					ExpireAt:   time.Now().Unix(),
				},
				{
					Name:       "my-name",
					Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					ExpireAt:   time.Now().Unix(),
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
			BuyOrders: []BuyOrder{
				{
					Buyer: "",
				},
			},
		}).Validate())
	})

	t.Run("fail - invalid aliases of RollApps", func(t *testing.T) {
		require.Error(t, (GenesisState{
			Params: DefaultParams(),
			AliasesOfRollapps: []AliasesOfChainId{
				{},
			},
		}).Validate())
	})
}
