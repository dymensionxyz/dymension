package types

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

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
		NamePrice_1Letter:      sdk.NewInt(5000 /* DYM */).MulRaw(1e18),
		NamePrice_2Letters:     sdk.NewInt(2500 /* DYM */).MulRaw(1e18),
		NamePrice_3Letters:     sdk.NewInt(1000 /* DYM */).MulRaw(1e18),
		NamePrice_4Letters:     sdk.NewInt(100 /* DYM */).MulRaw(1e18),
		NamePrice_5PlusLetters: sdk.NewInt(5 /* DYM */).MulRaw(1e18),

		AliasPrice_1Letter:      sdk.NewInt(5000 /* DYM */).MulRaw(1e18),
		AliasPrice_2Letters:     sdk.NewInt(1000 /* DYM */).MulRaw(1e18),
		AliasPrice_3Letters:     sdk.NewInt(250 /* DYM */).MulRaw(1e18),
		AliasPrice_4Letters:     sdk.NewInt(100 /* DYM */).MulRaw(1e18),
		AliasPrice_5Letters:     sdk.NewInt(25 /* DYM */).MulRaw(1e18),
		AliasPrice_6Letters:     sdk.NewInt(10 /* DYM */).MulRaw(1e18),
		AliasPrice_7PlusLetters: sdk.NewInt(5 /* DYM */).MulRaw(1e18),

		PriceExtends:  sdk.NewInt(5 /* DYM */).MulRaw(1e18),
		PriceDenom:    params.BaseDenom,
		MinOfferPrice: sdk.NewInt(10 /* DYM */).MulRaw(1e18),
	}
}

// DefaultChainsParams returns a default set of chains configuration
func DefaultChainsParams() ChainsParams {
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
				ChainId: "dymension_100-1",
				Aliases: []string{"test"},
			},
			// reserves alias for non Cosmos-SDK chains
			// TODO DymNS: review the list
			{
				ChainId: "bitcoin",
				Aliases: []string{"btc"},
			},
			{
				ChainId: "ethereum",
				Aliases: []string{"eth", "ether"},
			},
			{
				ChainId: "solana",
				Aliases: []string{"sol"},
			},
			{
				ChainId: "avalanche",
				Aliases: []string{"avax"},
			},
			{
				ChainId: "polygon",
				Aliases: []string{"matic"},
			},
			{
				ChainId: "polkadot",
				Aliases: []string{"dot"},
			},
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
		EnableTradingName:                true,
		EnableTradingAlias:               true,
	}
}

// DefaultPreservedRegistrationParams returns a default set of preserved registration parameters
func DefaultPreservedRegistrationParams() PreservedRegistrationParams {
	// TODO DymNS: Add default preserved registration params
	return PreservedRegistrationParams{
		ExpirationEpoch: 1727740799, // 2024-09-30 23:59:59 UTC
		PreservedDymNames: []PreservedDymName{
			{
				// this is just a pseudo address, replace it with the real one
				DymName:            "big-brain-staking",
				WhitelistedAddress: "dym1nd3qxp7xec90n9exr4ua3v26r940pl9nyy8whh",
			},
		},
	}
}

// NewParams creates a new Params object from given parameters
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

// Validate checks that the parameters have valid values.
func (m *Params) Validate() error {
	if err := m.Price.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "price params: %v", err)
	}
	if err := m.Chains.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "chains params: %v", err)
	}
	if err := m.Misc.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "misc params: %v", err)
	}
	if err := m.PreservedRegistration.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "preserved registration params: %v", err)
	}
	return nil
}

// Validate checks that the PriceParams have valid values.
func (m PriceParams) Validate() error {
	return validatePriceParams(m)
}

func (m *PriceParams) GetFirstYearDymNamePrice(name string) sdkmath.Int {
	switch len(name) {
	case 1:
		return m.NamePrice_1Letter
	case 2:
		return m.NamePrice_2Letters
	case 3:
		return m.NamePrice_3Letters
	case 4:
		return m.NamePrice_4Letters
	default:
		return m.NamePrice_5PlusLetters
	}
}

func (m *PriceParams) GetAliasPrice(alias string) sdkmath.Int {
	switch len(alias) {
	case 1:
		return m.AliasPrice_1Letter
	case 2:
		return m.AliasPrice_2Letters
	case 3:
		return m.AliasPrice_3Letters
	case 4:
		return m.AliasPrice_4Letters
	case 5:
		return m.AliasPrice_5Letters
	case 6:
		return m.AliasPrice_6Letters
	default:
		return m.AliasPrice_7PlusLetters
	}
}

// Validate checks that the ChainsParams have valid values.
func (m ChainsParams) Validate() error {
	return validateChainsParams(m)
}

// Validate checks that the MiscParams have valid values.
func (m MiscParams) Validate() error {
	return validateMiscParams(m)
}

// Validate checks that the PreservedRegistrationParams have valid values.
func (m PreservedRegistrationParams) Validate() error {
	return validatePreservedRegistrationParams(m)
}

// IsDuringWhitelistRegistrationPeriod returns true if still in the preserved registration period.
// It checks if the current block time is less than the expiration epoch.
func (m PreservedRegistrationParams) IsDuringWhitelistRegistrationPeriod(ctx sdk.Context) bool {
	return m.ExpirationEpoch >= ctx.BlockTime().Unix()
}

// validateEpochIdentifier checks if the given epoch identifier is valid.
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

// validatePriceParams checks if the given PriceParams are valid.
func validatePriceParams(i interface{}) error {
	m, ok := i.(PriceParams)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid parameter type: %T", i)
	}

	if err := validateNamePriceParams(m); err != nil {
		return err
	}

	if err := validateAliasPriceParams(m); err != nil {
		return err
	}

	if m.PriceDenom == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name price denom cannot be empty")
	}

	if err := (sdk.Coin{
		Denom:  m.PriceDenom,
		Amount: sdk.ZeroInt(),
	}).Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name price denom: %s", err)
	}

	if m.MinOfferPrice.IsNil() || m.MinOfferPrice.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min-offer-price is zero")
	} else if m.MinOfferPrice.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min-offer-price is negative")
	}

	return nil
}

// validateNamePriceParams checks if Dym-Name price in the given PriceParams are valid.
func validateNamePriceParams(m PriceParams) error {
	validateNamePrice := func(price sdkmath.Int, letterDesc string) error {
		if price.IsNil() || price.IsZero() {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Dym-Name price is zero for: %s", letterDesc)
		} else if price.IsNegative() {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Dym-Name price is negative for: %s", letterDesc)
		}
		return nil
	}

	if err := validateNamePrice(m.NamePrice_1Letter, "1 letter"); err != nil {
		return err
	}

	if err := validateNamePrice(m.NamePrice_2Letters, "2 letters"); err != nil {
		return err
	}

	if err := validateNamePrice(m.NamePrice_3Letters, "3 letters"); err != nil {
		return err
	}

	if err := validateNamePrice(m.NamePrice_4Letters, "4 letters"); err != nil {
		return err
	}

	if err := validateNamePrice(m.NamePrice_5PlusLetters, "5+ letters"); err != nil {
		return err
	}

	if err := validateNamePrice(m.PriceExtends, "yearly extends"); err != nil {
		return err
	}

	if m.NamePrice_1Letter.LTE(m.NamePrice_2Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "1 letter price must be greater than 2 letters price")
	}

	if m.NamePrice_2Letters.LTE(m.NamePrice_3Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "2 letters price must be greater than 3 letters price")
	}

	if m.NamePrice_3Letters.LTE(m.NamePrice_4Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "3 letters price must be greater than 4 letters price")
	}

	if m.NamePrice_4Letters.LTE(m.NamePrice_5PlusLetters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "4 letters price must be greater than 5+ letters price")
	}

	if m.NamePrice_5PlusLetters.LT(m.PriceExtends) {
		return errorsmod.Wrap(
			gerrc.ErrInvalidArgument,
			"5 letters price must be greater or equals to yearly extend price",
		)
	}

	return nil
}

// validateAliasPriceParams checks if Alias price in the given PriceParams are valid.
func validateAliasPriceParams(m PriceParams) error {
	validateAliasPrice := func(price sdkmath.Int, letterDesc string) error {
		if price.IsNil() || price.IsZero() {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Alias price is zero for: %s", letterDesc)
		} else if price.IsNegative() {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Alias price is negative for: %s", letterDesc)
		}
		return nil
	}

	if err := validateAliasPrice(m.AliasPrice_1Letter, "1 letter"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_2Letters, "2 letters"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_3Letters, "3 letters"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_4Letters, "4 letters"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_5Letters, "5 letters"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_6Letters, "6 letters"); err != nil {
		return err
	}

	if err := validateAliasPrice(m.AliasPrice_7PlusLetters, "7+ letters"); err != nil {
		return err
	}

	if m.AliasPrice_1Letter.LTE(m.AliasPrice_2Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "1 letter price must be greater than 2 letters price")
	}

	if m.AliasPrice_2Letters.LTE(m.AliasPrice_3Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "2 letters price must be greater than 3 letters price")
	}

	if m.AliasPrice_3Letters.LTE(m.AliasPrice_4Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "3 letters price must be greater than 4 letters price")
	}

	if m.AliasPrice_4Letters.LTE(m.AliasPrice_5Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "4 letters price must be greater than 5 letters price")
	}

	if m.AliasPrice_5Letters.LTE(m.AliasPrice_6Letters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "5 letters price must be greater than 6 letters price")
	}

	if m.AliasPrice_6Letters.LTE(m.AliasPrice_7PlusLetters) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "6 letters price must be greater than 7+ letters price")
	}

	return nil
}

// validateChainsParams checks if the given ChainsParams are valid.
func validateChainsParams(i interface{}) error {
	m, ok := i.(ChainsParams)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid parameter type: %T", i)
	}

	uniqueChainIdAliasAmongAliasConfig := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, record := range m.AliasesOfChainIds {
		chainID := record.ChainId
		aliases := record.Aliases
		if len(chainID) < 3 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias: chain ID must be at least 3 characters: %s", chainID)
		}

		if !dymnsutils.IsValidChainIdFormat(chainID) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias: chain ID is not well-formed: %s", chainID)
		}

		if _, ok := uniqueChainIdAliasAmongAliasConfig[chainID]; ok {
			return errorsmod.Wrapf(
				gerrc.ErrInvalidArgument,
				"alias: chain ID and alias must unique among all, found duplicated: %s", chainID,
			)
		}
		uniqueChainIdAliasAmongAliasConfig[chainID] = true

		for _, alias := range aliases {
			if !dymnsutils.IsValidAlias(alias) {
				return errorsmod.Wrapf(
					gerrc.ErrInvalidArgument,
					"alias is not well-formed: %s", alias,
				)
			}

			if _, ok := uniqueChainIdAliasAmongAliasConfig[alias]; ok {
				return errorsmod.Wrapf(
					gerrc.ErrInvalidArgument,
					"alias: chain ID and alias must unique among all, found duplicated: %s", alias,
				)
			}
			uniqueChainIdAliasAmongAliasConfig[alias] = true
		}
	}

	return nil
}

// validateMiscParams checks if the given MiscParams are valid.
func validateMiscParams(i interface{}) error {
	m, ok := i.(MiscParams)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid parameter type: %T", i)
	}

	if err := validateEpochIdentifier(m.BeginEpochHookIdentifier); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "begin epoch hook identifier: %v", err)
	}

	if err := validateEpochIdentifier(m.EndEpochHookIdentifier); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "end epoch hook identifier: %v", err)
	}

	if m.GracePeriodDuration < 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "grace period duration cannot be negative")
	}

	if m.SellOrderDuration <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Sell Orders duration can not be zero")
	}

	if m.PreservedClosedSellOrderDuration <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "preserved closed Sell Orders duration can not be zero")
	}

	if m.ProhibitSellDuration <= 0 {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "prohibit sell duration cannot be zero")
	}

	return nil
}

// validatePreservedRegistrationParams checks if the given PreservedRegistrationParams are valid.
func validatePreservedRegistrationParams(i interface{}) error {
	m, ok := i.(PreservedRegistrationParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if m.ExpirationEpoch < 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "expiration epoch cannot be negative")
	}

	uniquePairs := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, preservedDymName := range m.PreservedDymNames {
		if !dymnsutils.IsValidDymName(preservedDymName.DymName) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "preserved Dym-Name is not well-formed: %s", preservedDymName.DymName)
		}

		if !dymnsutils.IsValidBech32AccountAddress(preservedDymName.WhitelistedAddress, true) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "preserved Dym-Name has invalid whitelisted address: %s", preservedDymName.WhitelistedAddress)
		}

		pairKey := fmt.Sprintf("%s|%s", preservedDymName.DymName, preservedDymName.WhitelistedAddress)
		if _, ok := uniquePairs[pairKey]; ok {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "preserved dym name and whitelisted address pair is duplicated: %s & %s", preservedDymName.DymName, preservedDymName.WhitelistedAddress)
		}
		uniquePairs[pairKey] = true
	}

	return nil
}
