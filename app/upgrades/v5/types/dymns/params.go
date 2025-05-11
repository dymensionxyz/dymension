package dymns

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs get the params.ParamSet
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPriceParams, &m.Price, validatePriceParams),
		paramtypes.NewParamSetPair(KeyChainsParams, &m.Chains, validateChainsParams),
		paramtypes.NewParamSetPair(KeyMiscParams, &m.Misc, validateMiscParams),
	}
}

const (
	ModuleName = "dymns"
)

const (
	// MaxDymNameContactLength is the maximum length allowed for Dym-Name contact.
	MaxDymNameContactLength = 140

	// MaxConfigSize is the maximum size allowed for number Dym-Name configuration per Dym-Name.
	// This is another layer protects spamming the chain with large data.
	MaxConfigSize = 100

	// MinDymNamePriceStepsCount is the minimum number of price steps required for Dym-Name price.
	MinDymNamePriceStepsCount = 4

	// MinAliasPriceStepsCount is the minimum number of price steps required for Alias price.
	MinAliasPriceStepsCount = 4
)

// MinPriceValue is the minimum value allowed for price configuration.
var MinPriceValue = math.NewInt(1e18)

// NewParams creates a new Params object from given parameters
func NewParams(
	price PriceParams, chains ChainsParams, misc MiscParams,
) Params {
	return Params{
		Price:  price,
		Chains: chains,
		Misc:   misc,
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
	return nil
}

// Validate checks that the PriceParams have valid values.
func (m PriceParams) Validate() error {
	return validatePriceParams(m)
}

// GetFirstYearDymNamePrice returns the price for the first year of a Dym-Name registration.
func (m PriceParams) GetFirstYearDymNamePrice(name string) math.Int {
	return getElementAtIndexOrLast(m.NamePriceSteps, len(name)-1)
}

// GetAliasPrice returns the one-off-payment price for an Alias registration.
func (m PriceParams) GetAliasPrice(alias string) math.Int {
	return getElementAtIndexOrLast(m.AliasPriceSteps, len(alias)-1)
}

// getElementAtIndexOrLast returns the element at the given index or the last element if the index is out of bounds.
// TODO: negative index check https://github.com/dymensionxyz/dymension/issues/1738
func getElementAtIndexOrLast(elements []math.Int, index int) math.Int {
	if index >= len(elements) {
		return elements[len(elements)-1]
	}
	return elements[index]
}

// Validate checks that the ChainsParams have valid values.
func (m ChainsParams) Validate() error {
	return validateChainsParams(m)
}

// Validate checks that the MiscParams have valid values.
func (m MiscParams) Validate() error {
	return validateMiscParams(m)
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
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "price denom cannot be empty")
	}

	if err := (sdk.Coin{
		Denom:  m.PriceDenom,
		Amount: math.ZeroInt(),
	}).Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid price denom: %s", m.PriceDenom)
	}

	if m.MinOfferPrice.IsNil() || m.MinOfferPrice.LT(MinPriceValue) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min-offer-price is must be at least %s%s", MinPriceValue, m.PriceDenom)
	}

	const maxMinBidIncrementPercent = 10
	if m.MinBidIncrementPercent > maxMinBidIncrementPercent {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min-bid-increment-percent cannot be more than %d: %d", maxMinBidIncrementPercent, m.MinBidIncrementPercent)
	}

	return nil
}

// validateNamePriceParams checks if Dym-Name price in the given PriceParams are valid.
func validateNamePriceParams(m PriceParams) error {
	if len(m.NamePriceSteps) < MinDymNamePriceStepsCount {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"Dym-Name price steps must have at least %d steps", MinDymNamePriceStepsCount,
		)
	}

	for i, namePriceStep := range m.NamePriceSteps {
		if namePriceStep.IsNil() || namePriceStep.LT(MinPriceValue) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"Dym-Name price step at index %d must be at least %s%s", i, MinPriceValue, m.PriceDenom,
			)
		}
	}

	if m.PriceExtends.IsNil() || m.PriceExtends.LT(MinPriceValue) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"Dym-Name yearly extends price must be at least %s%s", MinPriceValue, m.PriceDenom,
		)
	}

	for i := 0; i < len(m.NamePriceSteps)-1; i++ {
		left := m.NamePriceSteps[i]
		right := m.NamePriceSteps[i+1]
		if left.LTE(right) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"previous Dym-Name price step must be greater than the next step at: %d", i,
			)
		}
	}

	lastPriceStep := m.NamePriceSteps[len(m.NamePriceSteps)-1]
	if lastPriceStep.LT(m.PriceExtends) {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"Dym-Name price step for the first year must be greater or equals to the yearly extends price: %s < %s",
			lastPriceStep, m.PriceExtends,
		)
	}

	return nil
}

// validateAliasPriceParams checks if Alias price in the given PriceParams are valid.
func validateAliasPriceParams(m PriceParams) error {
	if len(m.AliasPriceSteps) < MinAliasPriceStepsCount {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"alias price steps must have at least %d steps", MinAliasPriceStepsCount,
		)
	}

	for i, aliasPriceStep := range m.AliasPriceSteps {
		if aliasPriceStep.IsNil() || aliasPriceStep.LT(MinPriceValue) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"alias price step at index %d must be at least %s%s", i, MinPriceValue, m.PriceDenom,
			)
		}
	}

	for i := 0; i < len(m.AliasPriceSteps)-1; i++ {
		left := m.AliasPriceSteps[i]
		right := m.AliasPriceSteps[i+1]
		if left.LTE(right) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"previous alias price step must be greater than the next step at: %d", i,
			)
		}
	}

	return nil
}

// validateChainsParams checks if the given ChainsParams are valid.
func validateChainsParams(i interface{}) error {
	m, ok := i.(ChainsParams)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid parameter type: %T", i)
	}

	if err := validateAliasesOfChainIds(m.AliasesOfChainIds); err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInvalidArgument, err), "alias of chain-id")
	}

	return nil
}

func validateAliasesOfChainIds(aliasesOfChainIds []AliasesOfChainId) error {
	uniqueChainIdAliasAmongAliasConfig := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, record := range aliasesOfChainIds {
		chainID := record.ChainId
		aliases := record.Aliases
		if len(chainID) < 3 {
			return fmt.Errorf("chain ID must be at least 3 characters: %s", chainID)
		}

		if !dymnsutils.IsValidChainIdFormat(chainID) {
			return fmt.Errorf("chain ID is not well-formed: %s", chainID)
		}

		if _, ok := uniqueChainIdAliasAmongAliasConfig[chainID]; ok {
			return fmt.Errorf(
				"chain ID and alias must unique among all, found duplicated: %s", chainID,
			)
		}
		uniqueChainIdAliasAmongAliasConfig[chainID] = true

		for _, alias := range aliases {
			if !dymnsutils.IsValidAlias(alias) {
				return fmt.Errorf(
					"alias is not well-formed: %s", alias,
				)
			}

			if _, ok := uniqueChainIdAliasAmongAliasConfig[alias]; ok {
				return fmt.Errorf(
					"chain ID and alias must unique among all, found duplicated: %s", alias,
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

	if err := validateEpochIdentifier(m.EndEpochHookIdentifier); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "end epoch hook identifier: %v", err)
	}

	const minGracePeriodDuration = 30 * // number of days
		24 * time.Hour // hours per day
	if m.GracePeriodDuration < minGracePeriodDuration {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "grace period duration cannot be less than: %s", minGracePeriodDuration)
	}

	const maxSellOrderDuration = 7 * // number of days
		24 * time.Hour // hours per day
	if m.SellOrderDuration <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Sell Orders duration can not be zero")
	} else if m.SellOrderDuration > maxSellOrderDuration {
		// Sell Order duration cannot be too high because in increase the store size, potential causing DDoS
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "Sell Orders duration cannot be more than: %s", maxSellOrderDuration)
	}

	return nil
}
