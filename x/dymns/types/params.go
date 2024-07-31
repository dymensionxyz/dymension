package types

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyPriceParams is the key for the price params
	KeyPriceParams = []byte("PriceParams")

	// KeyChainsParams is the key for the chains params
	KeyChainsParams = []byte("ChainsParams")

	// KeyMiscParams is the key for the misc params
	KeyMiscParams = []byte("MiscParams")

	// KeyPreservedRegistrationParams is the key for the preserved registration params
	KeyPreservedRegistrationParams = []byte("PreservedRegistrationParams")
)

const (
	defaultBeginEpochHookIdentifier = "day" // less-frequently for cleanup
	defaultEndEpochHookIdentifier   = "hour"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// TODO DymNS: I'm not really familiar with this kind of params update via GOV, so please test with care.

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultPriceParams(),
		DefaultChainsParams(),
		DefaultMiscParams(),
		DefaultPreservedRegistrationParams(),
	)
}

// DefaultPriceParams returns a default set of price parameters
func DefaultPriceParams() PriceParams {
	return PriceParams{
		Price_1Letter:      sdk.NewInt(5000 /* DYM */).MulRaw(1e18),
		Price_2Letters:     sdk.NewInt(2500 /* DYM */).MulRaw(1e18),
		Price_3Letters:     sdk.NewInt(1000 /* DYM */).MulRaw(1e18),
		Price_4Letters:     sdk.NewInt(100 /* DYM */).MulRaw(1e18),
		Price_5PlusLetters: sdk.NewInt(5 /* DYM */).MulRaw(1e18),
		PriceExtends:       sdk.NewInt(5 /* DYM */).MulRaw(1e18),
		PriceDenom:         params.BaseDenom,
		MinOfferPrice:      sdk.NewInt(10 /* DYM */).MulRaw(1e18),
	}
}

// DefaultChainsParams returns a default set of chains configuration
func DefaultChainsParams() ChainsParams {
	//goland:noinspection SpellCheckingInspection
	return ChainsParams{
		AliasesOfChainIds: []AliasesOfChainId{
			{
				ChainId: "dymension_1100-1",
				Aliases: []string{"dym", "dymension"},
			},
			{
				ChainId: "blumbus_111-1",
				Aliases: []string{"blumbus"},
			},
			{
				ChainId: "froopyland_100-1",
				Aliases: []string{"froopyland", "frl"},
			},
			{
				ChainId: "cosmoshub-4",
				Aliases: []string{"cosmos"},
			},
			{
				ChainId: "osmosis-1",
				Aliases: []string{"osmosis"},
			},
			{
				ChainId: "axelar-dojo-1",
				Aliases: []string{"axelar"},
			},
			{
				ChainId: "stride-1",
				Aliases: []string{"stride"},
			},
			{
				ChainId: "kava_2222-10",
				Aliases: []string{"kava"},
			},
			{
				ChainId: "evmos_9001-2",
				Aliases: []string{"evmos"},
			},
			{
				ChainId: "nim_1122-1",
				Aliases: []string{"nim"},
			},
			{
				ChainId: "mande_18071918-1",
				Aliases: []string{"mande"},
			},
			{
				ChainId: "dymension_100-1",
				Aliases: []string{"test"},
			},
		},
		CoinType60ChainIds: []string{
			"evmos_9001-2",       // Evmos Mainnet
			"evmos_9001-3",       // Evmos Mainnet Re-Genesis
			"injective-1",        // Injective Mainnet
			"cronosmainnet_25-1", // Cronos Mainnet
		},
	}
}

// DefaultMiscParams returns a default set of misc parameters
func DefaultMiscParams() MiscParams {
	return MiscParams{
		BeginEpochHookIdentifier:         defaultBeginEpochHookIdentifier,
		EndEpochHookIdentifier:           defaultEndEpochHookIdentifier,
		GracePeriodDuration:              30 * 24 * time.Hour,
		SellOrderDuration:                3 * 24 * time.Hour,
		PreservedClosedSellOrderDuration: 7 * 24 * time.Hour,
		ProhibitSellDuration:             30 * 24 * time.Hour,
	}
}

// DefaultPreservedRegistrationParams returns a default set of preserved registration parameters
func DefaultPreservedRegistrationParams() PreservedRegistrationParams {
	// TODO DymNS: Add default preserved registration params
	return PreservedRegistrationParams{
		ExpirationEpoch: 1727740799, // 2024-09-30 23:59:59 UTC
		PreservedDymNames: []PreservedDymName{
			{
				// TODO DymNS: this is just a pseudo address, replace it with the real one
				DymName:            "big-brain-staking",
				WhitelistedAddress: "dym1nd3qxp7xec90n9exr4ua3v26r940pl9nyy8whh",
			},
		},
	}
}

func NewParams(
	price PriceParams, chains ChainsParams,
	misc MiscParams, preservedRegistration PreservedRegistrationParams,
) Params {
	return Params{
		Price:                 price,
		Chains:                chains,
		Misc:                  misc,
		PreservedRegistration: preservedRegistration,
	}
}

// ParamSetPairs get the params.ParamSet
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPriceParams, &m.Price, validatePriceParams),
		paramtypes.NewParamSetPair(KeyChainsParams, &m.Chains, validateChainsParams),
		paramtypes.NewParamSetPair(KeyMiscParams, &m.Misc, validateMiscParams),
		paramtypes.NewParamSetPair(KeyPreservedRegistrationParams, &m.PreservedRegistration, validatePreservedRegistrationParams),
	}
}

func (m *Params) Validate() error {
	if err := m.Price.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("price params: %v", err)
	}
	if err := m.Chains.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("chains params: %v", err)
	}
	if err := m.Misc.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("misc params: %v", err)
	}
	if err := m.PreservedRegistration.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("preserved registration params: %v", err)
	}
	return nil
}

func (m PriceParams) Validate() error {
	return validatePriceParams(m)
}

func (m ChainsParams) Validate() error {
	return validateChainsParams(m)
}

func (m MiscParams) Validate() error {
	return validateMiscParams(m)
}

func (m PreservedRegistrationParams) Validate() error {
	return validatePreservedRegistrationParams(m)
}

func (m PreservedRegistrationParams) IsDuringWhitelistRegistrationPeriod(ctx sdk.Context) bool {
	return m.ExpirationEpoch >= ctx.BlockTime().Unix()
}

func validateEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	switch v {
	case "hour", "day", "week":
	default:
		return fmt.Errorf("invalid epoch identifier: %s", v)
	}
	return nil
}

func validatePriceParams(i interface{}) error {
	m, ok := i.(PriceParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	validatePrice := func(price sdkmath.Int, letterDesc string) error {
		if price.IsNil() || price.IsZero() {
			return ErrValidationFailed.Wrapf("%s Dym-Name price is zero", letterDesc)
		} else if price.IsNegative() {
			return ErrValidationFailed.Wrapf("%s Dym-Name price is negative", letterDesc)
		}
		return nil
	}

	if err := validatePrice(m.Price_1Letter, "1 letter"); err != nil {
		return err
	}

	if err := validatePrice(m.Price_2Letters, "2 letters"); err != nil {
		return err
	}

	if err := validatePrice(m.Price_3Letters, "3 letters"); err != nil {
		return err
	}

	if err := validatePrice(m.Price_4Letters, "4 letters"); err != nil {
		return err
	}

	if err := validatePrice(m.Price_5PlusLetters, "5+ letters"); err != nil {
		return err
	}

	if err := validatePrice(m.PriceExtends, "yearly extends"); err != nil {
		return err
	}

	if m.Price_1Letter.LTE(m.Price_2Letters) {
		return ErrValidationFailed.Wrap("1 letter price must be greater than 2 letters price")
	}

	if m.Price_2Letters.LTE(m.Price_3Letters) {
		return ErrValidationFailed.Wrap("2 letters price must be greater than 3 letters price")
	}

	if m.Price_3Letters.LTE(m.Price_4Letters) {
		return ErrValidationFailed.Wrap("3 letters price must be greater than 4 letters price")
	}

	if m.Price_4Letters.LTE(m.Price_5PlusLetters) {
		return ErrValidationFailed.Wrap("4 letters price must be greater than 5+ letters price")
	}

	if m.Price_5PlusLetters.LT(m.PriceExtends) {
		return ErrValidationFailed.Wrap("5 letters price must be greater or equals to yearly extend price")
	}

	if m.PriceDenom == "" {
		return ErrValidationFailed.Wrap("Dym-Name price denom cannot be empty")
	}

	if err := (sdk.Coin{
		Denom:  m.PriceDenom,
		Amount: sdk.ZeroInt(),
	}).Validate(); err != nil {
		return ErrValidationFailed.Wrapf("invalid Dym-Name price denom: %s", err)
	}

	if m.MinOfferPrice.IsNil() || m.MinOfferPrice.IsZero() {
		return ErrValidationFailed.Wrap("min-offer-price is zero")
	} else if m.MinOfferPrice.IsNegative() {
		return ErrValidationFailed.Wrap("min-offer-price is negative")
	}

	return nil
}

func validateChainsParams(i interface{}) error {
	m, ok := i.(ChainsParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	uniqueChainIdAliasAmongAliasConfig := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, record := range m.AliasesOfChainIds {
		chainID := record.ChainId
		aliases := record.Aliases
		if len(chainID) < 3 {
			return ErrValidationFailed.Wrapf("alias: chain ID `%s` must be at least 3 characters", chainID)
		}

		if !dymnsutils.IsValidChainIdFormat(chainID) {
			return ErrValidationFailed.Wrapf("alias: chain ID `%s` is not well-formed", chainID)
		}

		if _, ok := uniqueChainIdAliasAmongAliasConfig[chainID]; ok {
			return ErrValidationFailed.Wrapf("alias: chain ID and alias must unique among all, found duplicated '%s'", chainID)
		}
		uniqueChainIdAliasAmongAliasConfig[chainID] = true

		for _, alias := range aliases {
			if !dymnsutils.IsValidAlias(alias) {
				return ErrValidationFailed.Wrapf("alias `%s` is not well-formed", alias)
			}

			if _, ok := uniqueChainIdAliasAmongAliasConfig[alias]; ok {
				return ErrValidationFailed.Wrapf("alias: chain ID and alias must unique among all, found duplicated '%s'", alias)
			}
			uniqueChainIdAliasAmongAliasConfig[alias] = true
		}
	}

	uniqueChainIdAmongCoinType60ChainsConfig := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, chainID := range m.CoinType60ChainIds {
		if len(chainID) < 3 {
			return ErrValidationFailed.Wrapf("coin-type-60 chains: chain ID `%s` must be at least 3 characters", chainID)
		}

		if !dymnsutils.IsValidChainIdFormat(chainID) {
			return ErrValidationFailed.Wrapf("coin-type-60 chains: chain ID `%s` is not well-formed", chainID)
		}

		if _, ok := uniqueChainIdAmongCoinType60ChainsConfig[chainID]; ok {
			return ErrValidationFailed.Wrapf("coin-type-60 chains: chain ID is not unique '%s'", chainID)
		}
		uniqueChainIdAmongCoinType60ChainsConfig[chainID] = true
	}

	return nil
}

func validateMiscParams(i interface{}) error {
	m, ok := i.(MiscParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := validateEpochIdentifier(m.BeginEpochHookIdentifier); err != nil {
		return ErrValidationFailed.Wrapf("begin epoch hook identifier: %v", err)
	}

	if err := validateEpochIdentifier(m.EndEpochHookIdentifier); err != nil {
		return ErrValidationFailed.Wrapf("end epoch hook identifier: %v", err)
	}

	if m.GracePeriodDuration < 0 {
		return ErrValidationFailed.Wrap("grace period duration cannot be negative")
	}

	if m.SellOrderDuration <= 0 {
		return ErrValidationFailed.Wrap("Sell Orders duration can not be zero")
	}

	if m.PreservedClosedSellOrderDuration <= 0 {
		return ErrValidationFailed.Wrap("preserved closed Sell Orders duration can not be zero")
	}

	if m.ProhibitSellDuration <= 0 {
		return ErrValidationFailed.Wrapf("prohibit sell duration cannot be zero")
	}

	return nil
}

func validatePreservedRegistrationParams(i interface{}) error {
	m, ok := i.(PreservedRegistrationParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if m.ExpirationEpoch < 0 {
		return ErrValidationFailed.Wrap("expiration epoch cannot be negative")
	}

	uniquePairs := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, preservedDymName := range m.PreservedDymNames {
		if !dymnsutils.IsValidDymName(preservedDymName.DymName) {
			return ErrValidationFailed.Wrapf("preserved dym name `%s` is not well-formed", preservedDymName.DymName)
		}

		if !dymnsutils.IsValidBech32AccountAddress(preservedDymName.WhitelistedAddress, true) {
			return ErrValidationFailed.Wrapf("preserved dym name `%s` has invalid whitelisted address `%s`", preservedDymName.DymName, preservedDymName.WhitelistedAddress)
		}

		pairKey := fmt.Sprintf("%s|%s", preservedDymName.DymName, preservedDymName.WhitelistedAddress)
		if _, ok := uniquePairs[pairKey]; ok {
			return ErrValidationFailed.Wrapf("preserved dym name `%s` with whitelisted address `%s` is duplicated", preservedDymName.DymName, preservedDymName.WhitelistedAddress)
		}
		uniquePairs[pairKey] = true
	}

	return nil
}
