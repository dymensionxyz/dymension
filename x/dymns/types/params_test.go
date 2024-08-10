package types

import (
	"fmt"
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
	moduleParams := DefaultParams()
	require.NoError(t, (&moduleParams).Validate())
}

func TestNewParams(t *testing.T) {
	moduleParams := NewParams(
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
	require.Equal(t, "a", moduleParams.Price.PriceDenom)
	require.Len(t, moduleParams.Chains.AliasesOfChainIds, 1)
	require.Equal(t, "dymension_1100-1", moduleParams.Chains.AliasesOfChainIds[0].ChainId)
	require.Len(t, moduleParams.Chains.AliasesOfChainIds[0].Aliases, 2)
	require.Equal(t, "b", moduleParams.Misc.BeginEpochHookIdentifier)
	require.Equal(t, "c", moduleParams.Misc.EndEpochHookIdentifier)
	require.Equal(t, 666.0, moduleParams.Misc.GracePeriodDuration.Hours())
	require.Equal(t, 333.0, moduleParams.Misc.SellOrderDuration.Hours())
	require.Equal(t, 222.0, moduleParams.Misc.PreservedClosedSellOrderDuration.Hours())
	require.Equal(t, 9999.0, moduleParams.Misc.ProhibitSellDuration.Hours())
	require.Equal(t, int64(888), moduleParams.PreservedRegistration.ExpirationEpoch)
	require.Len(t, moduleParams.PreservedRegistration.PreservedDymNames, 2)
	require.Equal(t, "an", moduleParams.PreservedRegistration.PreservedDymNames[0].DymName)
	require.Equal(t, "aa", moduleParams.PreservedRegistration.PreservedDymNames[0].WhitelistedAddress)
}

func TestDefaultPriceParams(t *testing.T) {
	priceParams := DefaultPriceParams()
	require.NoError(t, priceParams.Validate())

	t.Run("ensure setting is correct", func(t *testing.T) {
		i, ok := sdk.NewIntFromString("5" + "000000000000000000")
		require.True(t, ok)
		require.Equal(t, i, priceParams.NamePriceSteps[4])
	})

	t.Run("ensure price setting is at least 1 DYM", func(t *testing.T) {
		oneDym, ok := sdk.NewIntFromString("1" + "000000000000000000")
		require.True(t, ok)
		if oneDym.GT(priceParams.NamePriceSteps[4]) {
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
	moduleParams := DefaultParams()
	paramSetPairs := (&moduleParams).ParamSetPairs()
	require.Len(t, paramSetPairs, 4)
}

func TestParams_Validate(t *testing.T) {
	moduleParams := DefaultParams()
	require.NoError(t, (&moduleParams).Validate())

	moduleParams = DefaultParams()
	moduleParams.Price.MinOfferPrice = sdk.ZeroInt()
	require.Error(t, (&moduleParams).Validate())

	moduleParams = DefaultParams()
	moduleParams.Chains.AliasesOfChainIds = []AliasesOfChainId{{ChainId: "@"}}
	require.Error(t, (&moduleParams).Validate())

	moduleParams = DefaultParams()
	moduleParams.Misc.PreservedClosedSellOrderDuration = 0
	require.Error(t, (&moduleParams).Validate())

	moduleParams = DefaultParams()
	moduleParams.PreservedRegistration.ExpirationEpoch = -1
	require.Error(t, (&moduleParams).Validate())
}

func TestPriceParams_Validate(t *testing.T) {
	t.Run("pass - default must be valid", func(t *testing.T) {
		defaultPriceParams := DefaultPriceParams()

		// copy to ensure no new fields are added
		validPriceParams := PriceParams{
			NamePriceSteps:  defaultPriceParams.NamePriceSteps,
			AliasPriceSteps: defaultPriceParams.AliasPriceSteps,
			PriceExtends:    defaultPriceParams.PriceExtends,
			PriceDenom:      defaultPriceParams.PriceDenom,
			MinOfferPrice:   defaultPriceParams.MinOfferPrice,
		}

		require.NoError(t, validPriceParams.Validate())
	})

	t.Run("fail - price steps must be ordered descending", func(t *testing.T) {
		for i := 0; i < len(DefaultPriceParams().NamePriceSteps)-1; i++ {
			priceParams := DefaultPriceParams()
			priceParams.NamePriceSteps[i], priceParams.NamePriceSteps[i+1] = priceParams.NamePriceSteps[i+1], priceParams.NamePriceSteps[i]
			require.ErrorContains(t,
				priceParams.Validate(),
				fmt.Sprintf("previous Dym-Name price step must be greater than the next step at: %d", i),
			)
		}

		for i := 0; i < len(DefaultPriceParams().AliasPriceSteps)-1; i++ {
			priceParams := DefaultPriceParams()
			priceParams.AliasPriceSteps[i], priceParams.AliasPriceSteps[i+1] = priceParams.AliasPriceSteps[i+1], priceParams.AliasPriceSteps[i]
			require.ErrorContains(t,
				priceParams.Validate(),
				fmt.Sprintf("previous alias price step must be greater than the next step at: %d", i),
			)
		}
	})

	t.Run("mix - minimum price step count", func(t *testing.T) {
		defaultPriceParams := DefaultPriceParams()

		for size := 0; size <= (MinDymNamePriceStepsCount+MinAliasPriceStepsCount)*2; size++ {
			priceSteps := make([]sdkmath.Int, size)
			for i := 0; i < size; i++ {
				priceSteps[i] = sdk.NewInt(int64(1000 - i)).MulRaw(1e18)
			}

			m1 := defaultPriceParams
			m1.NamePriceSteps = priceSteps
			if size >= MinDymNamePriceStepsCount {
				require.NoError(t, m1.Validate())
			} else {
				require.ErrorContains(t, m1.Validate(), "price steps must have at least")
			}

			m2 := defaultPriceParams
			m2.AliasPriceSteps = priceSteps
			if size >= MinAliasPriceStepsCount {
				require.NoError(t, m2.Validate())
			} else {
				require.ErrorContains(t, m2.Validate(), "price steps must have at least")
			}
		}

		require.GreaterOrEqual(t, MinDymNamePriceStepsCount, 3, "why it so low?")
		require.GreaterOrEqual(t, MinAliasPriceStepsCount, 3, "why it so low?")
	})

	t.Run("fail - price denom", func(t *testing.T) {
		m := DefaultPriceParams()

		m.PriceDenom = ""
		require.ErrorContains(t, m.Validate(), "price denom cannot be empty")

		for _, denom := range []string{"-", "--", "0"} {
			m.PriceDenom = denom
			require.ErrorContains(t, m.Validate(), "invalid price denom")
		}
	})

	t.Run("fail - price is too low", func(t *testing.T) {
		defaultPriceParams := DefaultPriceParams()

		type tc struct {
			name     string
			modifier func(PriceParams, sdkmath.Int) PriceParams
		}

		tests := []tc{
			{
				name: "price extends",
				modifier: func(p PriceParams, v sdkmath.Int) PriceParams {
					p.PriceExtends = v
					return p
				},
			},
			{
				name: "min offer price",
				modifier: func(p PriceParams, v sdkmath.Int) PriceParams {
					p.MinOfferPrice = v
					return p
				},
			},
		}

		for i := 0; i < len(defaultPriceParams.NamePriceSteps); i++ {
			tests = append(tests, tc{
				name: fmt.Sprintf("name price steps [%d]", i),
				modifier: func(p PriceParams, v sdkmath.Int) PriceParams {
					p.NamePriceSteps[i] = v
					return p
				},
			})
		}

		for i := 0; i < len(defaultPriceParams.AliasPriceSteps); i++ {
			tests = append(tests, tc{
				name: fmt.Sprintf("alias price steps [%d]", i),
				modifier: func(p PriceParams, v sdkmath.Int) PriceParams {
					p.AliasPriceSteps[i] = v
					return p
				},
			})
		}

		for _, test := range tests {
			for _, badPrice := range []sdkmath.Int{{}, sdkmath.NewInt(-1), sdkmath.ZeroInt(), MinPriceValue.Sub(sdkmath.NewInt(1))} {
				t.Run(fmt.Sprintf("%s with v = %v", test.name, badPrice), func(t *testing.T) {
					p := test.modifier(DefaultPriceParams(), badPrice)
					err := (&p).Validate()
					require.Error(t, err)
					require.Contains(t, err.Error(), "must be at least")
				})
			}
		}
	})

	t.Run("fail - yearly extends price can not be higher than last step price", func(t *testing.T) {
		defaultPriceParams := DefaultPriceParams()
		defaultPriceParams.PriceExtends = defaultPriceParams.NamePriceSteps[len(defaultPriceParams.NamePriceSteps)-1].AddRaw(1)

		require.ErrorContains(
			t, defaultPriceParams.Validate(),
			"Dym-Name price step for the first year must be greater or equals to the yearly extends price",
		)
	})

	t.Run("fail - invalid type", func(t *testing.T) {
		require.Error(t, validatePriceParams("hello world"))
		require.Error(t, validatePriceParams(&PriceParams{}), "not accept pointer")
	})
}

func TestPriceParams_GetPrice(t *testing.T) {
	priceParams := DefaultPriceParams()

	t.Run("for name length <= number of price steps, use the corresponding price step", func(t *testing.T) {
		runes := make([]rune, 0)
		for length := 1; length <= len(priceParams.NamePriceSteps); length++ {
			runes = append(runes, 'a')
			require.Equal(t, priceParams.NamePriceSteps[length-1], priceParams.GetFirstYearDymNamePrice(string(runes)))
		}
	})

	t.Run("for name length >= number of price steps, use the last price step", func(t *testing.T) {
		wantPrice := priceParams.NamePriceSteps[len(priceParams.NamePriceSteps)-1]

		runes := make([]rune, len(priceParams.NamePriceSteps))
		for i := range runes {
			runes[i] = 'a'
		}
		require.Equal(t, wantPrice, priceParams.GetFirstYearDymNamePrice(string(runes)))

		for extraLettersCount := 1; extraLettersCount < 1000; extraLettersCount++ {
			runes = append(runes, 'a')
			require.Equal(t, wantPrice, priceParams.GetFirstYearDymNamePrice(string(runes)))
		}
	})

	t.Run("for alias length <= number of price steps, use the corresponding price step", func(t *testing.T) {
		runes := make([]rune, 0)
		for length := 1; length <= len(priceParams.AliasPriceSteps); length++ {
			runes = append(runes, 'a')
			require.Equal(t, priceParams.AliasPriceSteps[length-1], priceParams.GetAliasPrice(string(runes)))
		}
	})

	t.Run("for alias length >= number of price steps, use the last price step", func(t *testing.T) {
		wantPrice := priceParams.AliasPriceSteps[len(priceParams.AliasPriceSteps)-1]

		runes := make([]rune, len(priceParams.AliasPriceSteps))
		for i := range runes {
			runes[i] = 'a'
		}
		require.Equal(t, wantPrice, priceParams.GetAliasPrice(string(runes)))

		for extraLettersCount := 1; extraLettersCount < 1000; extraLettersCount++ {
			runes = append(runes, 'a')
			require.Equal(t, wantPrice, priceParams.GetAliasPrice(string(runes)))
		}
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
			name:     "pass - default is valid",
			modifier: func(p ChainsParams) ChainsParams { return p },
		},
		{
			name: "pass - alias: empty is valid",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = nil
				return p
			},
		},
		{
			name: "pass - alias: empty alias of chain is valid",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "dymension_1100-1", Aliases: nil},
				}
				return p
			},
		},
		{
			name: "pass - alias: valid and correct alias",
			modifier: func(p ChainsParams) ChainsParams {
				p.AliasesOfChainIds = []AliasesOfChainId{
					{ChainId: "blumbus_111-1", Aliases: []string{"bb", "blumbus"}},
					{ChainId: "dymension_1100-1", Aliases: []string{"dym"}},
				}
				return p
			},
		},
		{
			name: "fail - alias: chain_id and alias must be unique among all, case alias & alias",
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
			name: "fail - alias: chain_id and alias must be unique among all, case chain-id & alias",
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
			name: "fail - alias: reject if chain-id format is bad",
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
			name: "fail - alias: reject if chain-id format is bad",
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
			name: "fail - alias: reject if alias format is bad",
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

	t.Run("fail - invalid type", func(t *testing.T) {
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
			name:     "pass - default is valid",
			modifier: func(p MiscParams) MiscParams { return p },
		},
		{
			name: "pass - all = 1 is valid",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 1 * time.Nanosecond
				p.SellOrderDuration = 1 * time.Nanosecond
				p.PreservedClosedSellOrderDuration = 1 * time.Nanosecond
				p.ProhibitSellDuration = 1 * time.Nanosecond
				return p
			},
		},
		{
			name: "pass - minimum allowed",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 0
				p.SellOrderDuration = 1 * time.Nanosecond
				p.PreservedClosedSellOrderDuration = 1 * time.Nanosecond
				p.ProhibitSellDuration = 1 * time.Nanosecond
				return p
			},
		},
		{
			name: "pass - begin epoch hour is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "hour"
				return p
			},
		},
		{
			name: "pass - end epoch hour is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "hour"
				return p
			},
		},
		{
			name: "pass - begin epoch day is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "day"
				return p
			},
		},
		{
			name: "pass - end epoch day is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "day"
				return p
			},
		},
		{
			name: "pass - begin epoch week is valid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "week"
				return p
			},
		},
		{
			name: "pass - end epoch week is valid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "week"
				return p
			},
		},
		{
			name: "fail - begin other epoch is invalid",
			modifier: func(p MiscParams) MiscParams {
				p.BeginEpochHookIdentifier = "invalid"
				return p
			},
			wantErr:         true,
			wantErrContains: "invalid epoch identifier: invalid",
		},
		{
			name: "fail - end other epoch is invalid",
			modifier: func(p MiscParams) MiscParams {
				p.EndEpochHookIdentifier = "invalid"
				return p
			},
			wantErr:         true,
			wantErrContains: "invalid epoch identifier: invalid",
		},
		{
			name: "pass - grace period = 0 is valid",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = 0
				return p
			},
		},
		{
			name: "fail - grace period can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.GracePeriodDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "grace period duration cannot be negative",
		},
		{
			name: "fail - days SO duration can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.SellOrderDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "Sell Orders duration can not be zero",
		},
		{
			name: "fail - days SO duration can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.SellOrderDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "Sell Orders duration can not be zero",
		},
		{
			name: "fail - days preserved closed SO duration can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.PreservedClosedSellOrderDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "preserved closed Sell Orders duration can not be zero",
		},
		{
			name: "fail - days preserved closed SO duration can not be negative",
			modifier: func(p MiscParams) MiscParams {
				p.PreservedClosedSellOrderDuration = -1 * time.Nanosecond
				return p
			},
			wantErr:         true,
			wantErrContains: "preserved closed Sell Orders duration can not be zero",
		},
		{
			name: "fail - days prohibit sell can not be zero",
			modifier: func(p MiscParams) MiscParams {
				p.ProhibitSellDuration = 0
				return p
			},
			wantErr:         true,
			wantErrContains: "prohibit sell duration cannot be zero",
		},
		{
			name: "fail - days prohibit sell can not be negative",
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

	t.Run("fail - invalid type", func(t *testing.T) {
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
			name:     "pass - default is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams { return p },
		},
		{
			name: "pass - valid",
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
			name: "pass - expiration epoch = 0 is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = 0
				return p
			},
		},
		{
			name: "fail - negative expiration epoch is invalid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = -1
				return p
			},
			wantErr:         true,
			wantErrContains: "expiration epoch cannot be negative",
		},
		{
			name: "pass - expiration epoch in the past is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.ExpirationEpoch = 1 // epoch 1 is in the past
				return p
			},
		},
		{
			name: "pass - empty preserved Dym-Name list is valid",
			modifier: func(p PreservedRegistrationParams) PreservedRegistrationParams {
				p.PreservedDymNames = nil
				return p
			},
		},
		{
			name: "fail - Dym-Name must be valid",
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
			name: "fail - Dym-Name must be valid, not allow @ part",
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
			name: "fail - address must be valid bech32",
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
			name: "fail - address hex is not allowed",
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
			name: "fail - duplicated pairs is now allowed",
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

	t.Run("fail - invalid type", func(t *testing.T) {
		require.Error(t, validatePreservedRegistrationParams("hello world"))
		require.Error(t, validatePreservedRegistrationParams(&PreservedRegistrationParams{}), "not accept pointer")
	})
}

func TestPreservedRegistrationParams_IsDuringWhitelistRegistrationPeriod(t *testing.T) {
	moduleParams := PreservedRegistrationParams{
		ExpirationEpoch: 100,
	}

	ctxAtEpoch := func(epoch int64) sdk.Context {
		ctx := sdk.Context{}
		ctx = ctx.WithBlockTime(time.Unix(epoch, 0))
		return ctx
	}

	require.True(t, moduleParams.IsDuringWhitelistRegistrationPeriod(ctxAtEpoch(99)))
	require.True(t, moduleParams.IsDuringWhitelistRegistrationPeriod(ctxAtEpoch(100)))
	require.False(t, moduleParams.IsDuringWhitelistRegistrationPeriod(ctxAtEpoch(101)))

	moduleParams = PreservedRegistrationParams{
		ExpirationEpoch: 0,
	}
	require.False(t, moduleParams.IsDuringWhitelistRegistrationPeriod(ctxAtEpoch(1)))
}

func Test_validateEpochIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		i       interface{}
		wantErr bool
	}{
		{
			name: "pass - 'hour' is valid",
			i:    "hour",
		},
		{
			name: "pass - 'day' is valid",
			i:    "day",
		},
		{
			name: "pass - 'week' is valid",
			i:    "week",
		},
		{
			name:    "fail - empty",
			i:       "",
			wantErr: true,
		},
		{
			name:    "fail - not string",
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
