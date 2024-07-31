package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamKeyTable(t *testing.T) {
	m := ParamKeyTable()
	require.NotNil(t, m)
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, (&params).Validate())
}

func TestNewParams(t *testing.T) {
	params := NewParams(
		PriceParams{
			PriceDenom: "a",
		},
		ChainsParams{
			AliasesOfChainIds: []AliasesOfChainId{
				{
					ChainId: "dymension_1100-1",
					Aliases: []string{"dym", "dymension"},
				},
			},
			CoinType60ChainIds: []string{"injective-1"},
		},
		MiscParams{
			BeginEpochHookIdentifier:         "b",
			EndEpochHookIdentifier:           "c",
			GracePeriodDuration:              666 * time.Hour,
			SellOrderDuration:                333 * time.Hour,
			PreservedClosedSellOrderDuration: 222 * time.Hour,
			ProhibitSellDuration:             9999 * time.Hour,
		},
		PreservedRegistrationParams{
			ExpirationEpoch: 888,
			PreservedDymNames: []PreservedDymName{
				{
					DymName:            "an",
					WhitelistedAddress: "aa",
				},
				{
					DymName:            "bn",
					WhitelistedAddress: "ba",
				},
			},
		},
	)
	require.Equal(t, "a", params.Price.PriceDenom)
	require.Len(t, params.Chains.AliasesOfChainIds, 1)
	require.Equal(t, "dymension_1100-1", params.Chains.AliasesOfChainIds[0].ChainId)
	require.Len(t, params.Chains.AliasesOfChainIds[0].Aliases, 2)
	require.Len(t, params.Chains.CoinType60ChainIds, 1)
	require.Equal(t, params.Chains.CoinType60ChainIds[0], "injective-1")
	require.Equal(t, "b", params.Misc.BeginEpochHookIdentifier)
	require.Equal(t, "c", params.Misc.EndEpochHookIdentifier)
	require.Equal(t, 666.0, params.Misc.GracePeriodDuration.Hours())
	require.Equal(t, 333.0, params.Misc.SellOrderDuration.Hours())
	require.Equal(t, 222.0, params.Misc.PreservedClosedSellOrderDuration.Hours())
	require.Equal(t, 9999.0, params.Misc.ProhibitSellDuration.Hours())
	require.Equal(t, int64(888), params.PreservedRegistration.ExpirationEpoch)
	require.Len(t, params.PreservedRegistration.PreservedDymNames, 2)
	require.Equal(t, "an", params.PreservedRegistration.PreservedDymNames[0].DymName)
	require.Equal(t, "aa", params.PreservedRegistration.PreservedDymNames[0].WhitelistedAddress)
}

func TestDefaultPriceParams(t *testing.T) {
	priceParams := DefaultPriceParams()
	require.NoError(t, priceParams.Validate())

	t.Run("ensure setting is correct", func(t *testing.T) {
		i, ok := sdk.NewIntFromString("5" + "000000000000000000")
		require.True(t, ok)
		require.Equal(t, i, priceParams.Price_5PlusLetters)
	})

	t.Run("ensure price setting is at least 1 DYM", func(t *testing.T) {
		oneDym, ok := sdk.NewIntFromString("1" + "000000000000000000")
		require.True(t, ok)
		if oneDym.GT(priceParams.Price_5PlusLetters) {
			require.Fail(t, "price should be at least 1 DYM")
		}
		if oneDym.GT(priceParams.PriceExtends) {
			require.Fail(t, "price should be at least 1 DYM")
		}
	})
}

func TestDefaultChainsParams(t *testing.T) {
	require.NoError(t, DefaultChainsParams().Validate())
}

func TestDefaultMiscParams(t *testing.T) {
	require.NoError(t, DefaultMiscParams().Validate())
}

func TestDefaultPreservedRegistrationParams(t *testing.T) {
	require.NoError(t, DefaultPreservedRegistrationParams().Validate())
}

func TestParams_ParamSetPairs(t *testing.T) {
	params := DefaultParams()
	paramSetPairs := (&params).ParamSetPairs()
	require.Len(t, paramSetPairs, 4)
}

func TestParams_Validate(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, (&params).Validate())

	params = DefaultParams()
	params.Price.Price_1Letter = sdk.ZeroInt()
	require.Error(t, (&params).Validate())

	params = DefaultParams()
	params.Chains.CoinType60ChainIds = []string{"invalid@"}
	require.Error(t, (&params).Validate())

	params = DefaultParams()
	params.Misc.PreservedClosedSellOrderDuration = 0
	require.Error(t, (&params).Validate())

	params = DefaultParams()
	params.PreservedRegistration.ExpirationEpoch = -1
	require.Error(t, (&params).Validate())
}

func TestPriceParams_Validate(t *testing.T) {
	validPriceParams := PriceParams{
		Price_1Letter:      sdk.NewInt(6),
		Price_2Letters:     sdk.NewInt(5),
		Price_3Letters:     sdk.NewInt(4),
		Price_4Letters:     sdk.NewInt(3),
		Price_5PlusLetters: sdk.NewInt(2),
		PriceExtends:       sdk.NewInt(2),
		PriceDenom:         "adym",
		MinOfferPrice:      sdk.NewInt(7),
	}

	require.NoError(t, validPriceParams.Validate())

	t.Run("price denom", func(t *testing.T) {
		m := validPriceParams
		m.PriceDenom = ""
		require.Error(t, m.Validate())

		m.PriceDenom = "--"
		require.Error(t, m.Validate())
	})

	t.Run("min offer price", func(t *testing.T) {
		m := validPriceParams

		m.MinOfferPrice = sdkmath.Int{}
		require.Error(t, m.Validate())

		m.MinOfferPrice = sdk.ZeroInt()
		require.Error(t, m.Validate())

		m.MinOfferPrice = sdk.NewInt(-1)
		require.Error(t, m.Validate())
	})

	type modifierPrice func(PriceParams, sdkmath.Int) PriceParams
	type swapPrice func(PriceParams) PriceParams

	testsInvalidPrice := []struct {
		name          string
		modifierPrice modifierPrice
		swapPrice     swapPrice
	}{
		{
			name:          "invalid 1 letter price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.Price_1Letter = v; return p },
			swapPrice: func(params PriceParams) PriceParams {
				backup := params.Price_1Letter
				params.Price_1Letter = params.Price_2Letters
				params.Price_2Letters = backup
				return params
			},
		},
		{
			name:          "invalid 2 letters price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.Price_2Letters = v; return p },
			swapPrice: func(params PriceParams) PriceParams {
				backup := params.Price_2Letters
				params.Price_2Letters = params.Price_3Letters
				params.Price_3Letters = backup
				return params
			},
		},
		{
			name:          "invalid 3 letters price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.Price_3Letters = v; return p },
			swapPrice: func(params PriceParams) PriceParams {
				backup := params.Price_3Letters
				params.Price_3Letters = params.Price_4Letters
				params.Price_4Letters = backup
				return params
			},
		},
		{
			name:          "invalid 4 letters price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.Price_4Letters = v; return p },
			swapPrice: func(params PriceParams) PriceParams {
				backup := params.Price_4Letters
				params.Price_4Letters = params.Price_5PlusLetters
				params.Price_5PlusLetters = backup
				return params
			},
		},
		{
			name:          "invalid 5+ letters price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.Price_5PlusLetters = v; return p },
		},
		{
			name:          "invalid yearly extends price",
			modifierPrice: func(p PriceParams, v sdkmath.Int) PriceParams { p.PriceExtends = v; return p },
			swapPrice: func(params PriceParams) PriceParams {
				params.PriceExtends = params.Price_5PlusLetters.Add(params.Price_5PlusLetters)
				return params
			},
		},
	}
	for _, tt := range testsInvalidPrice {
		t.Run(tt.name, func(t *testing.T) {
			err1 := tt.modifierPrice(validPriceParams, sdk.ZeroInt()).Validate()
			require.Error(t, err1)
			require.Contains(t, err1.Error(), "is zero")
			err2 := tt.modifierPrice(validPriceParams, sdk.NewInt(-1)).Validate()
			require.Error(t, err2)
			require.Contains(t, err2.Error(), "is negative")

			if tt.swapPrice != nil {
				err3 := tt.swapPrice(validPriceParams).Validate()
				require.Error(t, err3)
				require.Contains(t, err3.Error(), "must be greater")
			}
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		require.Error(t, validatePriceParams("hello world"))
		require.Error(t, validatePriceParams(&PriceParams{}), "not accept pointer")
	})
}

//goland:noinspection SpellCheckingInspection
func TestChainsParams_Validate(t *testing.T) {
	tests := []struct {
		name            string
		modifier        func(params ChainsParams) ChainsParams
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "default is valid",
			modifier: func(p ChainsParams) ChainsParams { return p },
		},
		{
			name: "alias: empty is valid",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = nil
				return p
			},
		},
		{
			name: "coin-type-60-chains: empty is valid",
			modifier: func(p ChainsParams) ChainsParams {
				p.CoinType60ChainIds = nil
				return p
			},
		},
		{
			name: "alias: empty alias of chain is valid",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension_1100-1", Aliases: nil},
				}
				return p
			},
		},
		{
			name: "alias: valid and correct alias",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "blumbus_111-1", Aliases: []string{"bb", "blumbus"}},
					{ChainId: "dymension_1100-1", Aliases: []string{"dym"}},
				}
				return p
			},
		},
		{
			name: "coin-type-60-chains: valid and correct alias",
			modifier: func(p ChainsParams) ChainsParams {
				p.CoinType60ChainIds = []string{"injective-1", "cronosmainnet_25-1"}
				return p
			},
		},
		{
			name: "alias: chain_id and alias must be unique among all, case alias & alias",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension_1100-1", Aliases: []string{"dym"}},
					{ChainId: "blumbus_111-1", Aliases: []string{"dym", "blumbus"}},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "chain ID and alias must unique among all",
		},
		{
			name: "coin-type-60-chains: chain_id must be unique among all",
			modifier: func(p ChainsParams) ChainsParams {
				p.CoinType60ChainIds = []string{"injective-1", "cronosmainnet_25-1", "injective-1"}
				return p
			},
			wantErr:         true,
			wantErrContains: "chain ID is not unique",
		},
		{
			name: "alias: chain_id and alias must be unique among all, case chain-id & alias",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension_1100-1", Aliases: []string{"dym", "dymension"}},
					{ChainId: "blumbus_111-1", Aliases: []string{"blumbus", "cosmoshub"}},
					{ChainId: "cosmoshub", Aliases: []string{"cosmos"}},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "chain ID and alias must unique among all",
		},
		{
			name: "alias: reject if chain-id format is bad",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension@", Aliases: []string{"dym"}},
					{ChainId: "blumbus_111-1", Aliases: []string{"blumbus"}},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "is not well-formed",
		},
		{
			name: "coin-type-60-chains: reject if chain-id format is bad",
			modifier: func(p ChainsParams) ChainsParams {
				p.CoinType60ChainIds = []string{"injective@"}
				return p
			},
			wantErr:         true,
			wantErrContains: "is not well-formed",
		},
		{
			name: "coin-type-60-chains: reject if chain-id format is bad",
			modifier: func(p ChainsParams) ChainsParams {
				p.CoinType60ChainIds = []string{"in"}
				return p
			},
			wantErr:         true,
			wantErrContains: "must be at least 3 characters",
		},
		{
			name: "alias: reject if chain-id format is bad",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "d", Aliases: []string{"dym"}},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "must be at least 3 characters",
		},
		{
			name: "alias: reject if alias format is bad",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension_1100-1", Aliases: []string{"dym-dym"}},
					{ChainId: "blumbus_111-1", Aliases: []string{"blumbus"}},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "is not well-formed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.modifier(DefaultChainsParams()).Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		require.Error(t, validateChainsParams("hello world"))
		require.Error(t, validateChainsParams(&ChainsParams{}), "not accept pointer")
	})
}

func TestMiscParams_Validate(t *testing.T) {
	tests := []struct {
		name            string
		modifier        func(MiscParams) MiscParams
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "default is valid",
			modifier: func(p MiscParams) MiscParams { return p },
		},
		{
			name: "all = 1 is valid",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 1 * time.Nanosecond
				p.SellOrderDuration = 1 * time.Nanosecond
				p.PreservedClosedSellOrderDuration = 1 * time.Nanosecond
				p.ProhibitSellDuration = 1 * time.Nanosecond
				return p
			},
		},
		{
			name: "minimum allowed",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 0
				p.SellOrderDuration = 1 * time.Nanosecond
				p.PreservedClosedSellOrderDuration = 1 * time.Nanosecond
				p.ProhibitSellDuration = 1 * time.Nanosecond
				return p
			},
		},
		{
			name: "epoch hour is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "hour"
				return p
			},
		},
		{
			name: "epoch hour is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "hour"
				return p
			},
		},
		{
			name: "epoch day is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "day"
				return p
			},
		},
		{
			name: "epoch day is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "day"
				return p
			},
		},
		{
			name: "epoch week is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "week"
				return p
			},
		},
		{
			name: "epoch week is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "week"
				return p
			},
		},
		{
			name: "other epoch is invalid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "invalid"
				return p
			},
			wantErr:         true,
			wantErrContains: "invalid epoch identifier: invalid",
		},
		{
			name: "other epoch is invalid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "invalid"
				return p
			},
			wantErr:         true,
			wantErrContains: "invalid epoch identifier: invalid",
		},
		{
			name: "grace period = 0 is valid",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 0
				return p
			},
		},
		{
			name: "grace period can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "grace period duration cannot be negative",
		},
		{
			name: "days SO duration can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.SellOrderDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "Sell Orders duration can not be zero",
		},
		{
			name: "days SO duration can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.SellOrderDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "Sell Orders duration can not be zero",
		},
		{
			name: "days preserved closed SO duration can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.PreservedClosedSellOrderDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "preserved closed Sell Orders duration can not be zero",
		},
		{
			name: "days preserved closed SO duration can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.PreservedClosedSellOrderDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "preserved closed Sell Orders duration can not be zero",
		},
		{
			name: "days prohibit sell can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.ProhibitSellDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "prohibit sell duration cannot be zero",
		},
		{
			name: "days prohibit sell can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.ProhibitSellDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "prohibit sell duration cannot be zero",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.modifier(DefaultMiscParams()).Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		require.Error(t, validateMiscParams("hello world"))
		require.Error(t, validateMiscParams(&MiscParams{}), "not accept pointer")
	})
}

//goland:noinspection SpellCheckingInspection
func TestPreservedRegistrationParams_Validate(t *testing.T) {
	tests := []struct {
		name            string
		modifier        func(PreservedRegistrationParams) PreservedRegistrationParams
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "default is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams { return p },
		},
		{
			name: "valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = 1
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "a",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
					{
						DymName:            "b",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
				}
				return p
			},
		},
		{
			name: "expiration epoch = 0 is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = 0
				return p
			},
		},
		{
			name: "negative expiration epoch is invalid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = -1
				return p
			},
			wantErr:         true,
			wantErrContains: "expiration epoch cannot be negative",
		},
		{
			name: "expiration epoch in the pass is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = 1 // epoch 1 is in the past
				return p
			},
		},
		{
			name: "empty preserved Dym-Name list is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = nil
				return p
			},
		},
		{
			name: "Dym-Name must be valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "!a!",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "is not well-formed",
		},
		{
			name: "Dym-Name must be valid, not allow @ part",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "invalid@dym",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "is not well-formed",
		},
		{
			name: "address must be valid bech32",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "a",
						WhitelistedAddress: "dym1fl48vsnms",
					},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "has invalid whitelisted address",
		},
		{
			name: "address hex is not allowed",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "a",
						WhitelistedAddress: "0x1234567890123456789012345678901234567890",
					},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "has invalid whitelisted address",
		},
		{
			name: "duplicated pairs is now allowed",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = []PreservedDymName{
					{
						DymName:            "a",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
					{
						DymName:            "bbbb",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
					{
						// duplicated
						DymName:            "a",
						WhitelistedAddress: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					},
				}
				return p
			},
			wantErr:         true,
			wantErrContains: "is duplicated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.modifier(DefaultPreservedRegistrationParams()).Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		require.Error(t, validatePreservedRegistrationParams("hello world"))
		require.Error(t, validatePreservedRegistrationParams(&PreservedRegistrationParams{}), "not accept pointer")
	})
}

func TestPreservedRegistrationParams_IsDuringWhitelistRegistrationPeriod(t *testing.T) {
	params := PreservedRegistrationParams{
		ExpirationEpoch: 100,
	}
	require.True(t, params.IsDuringWhitelistRegistrationPeriod(99))
	require.True(t, params.IsDuringWhitelistRegistrationPeriod(100))
	require.False(t, params.IsDuringWhitelistRegistrationPeriod(101))

	params = PreservedRegistrationParams{
		ExpirationEpoch: 0,
	}
	require.False(t, params.IsDuringWhitelistRegistrationPeriod(1))
}

func Test_validateEpochIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		i       interface{}
		wantErr bool
	}{
		{
			name: "'hour' is valid",
			i:    "hour",
		},
		{
			name: "'day' is valid",
			i:    "day",
		},
		{
			name: "'week' is valid",
			i:    "week",
		},
		{
			name:    "empty",
			i:       "",
			wantErr: true,
		},
		{
			name:    "not string",
			i:       1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				require.Error(t, validateEpochIdentifier(tt.i))
			} else {
				require.NoError(t, validateEpochIdentifier(tt.i))
			}
		})
	}
}
