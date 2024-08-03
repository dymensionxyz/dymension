package types

import (
	"fmt"
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// Validate checks if the DymName record is valid.
func (m *DymName) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "dym name is nil")
	}
	if m.Name == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is empty")
	}
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}
	if m.Owner == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is empty")
	}
	if !dymnsutils.IsValidBech32AccountAddress(m.Owner, true) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address: %s", m.Owner)
	}
	if m.Controller == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is empty")
	}
	if !dymnsutils.IsValidBech32AccountAddress(m.Controller, true) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}
	if m.ExpireAt == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "expiry is empty")
	}

	uniqueConfig := make(map[string]bool)
	// Describe usage of Go Map: only used for validation
	for _, config := range m.Configs {
		if err := config.Validate(); err != nil {
			return err
		}

		configIdentity := config.GetIdentity()
		if _, duplicated := uniqueConfig[configIdentity]; duplicated {
			return errorsmod.Wrapf(
				gerrc.ErrInvalidArgument, "dym name config is not unique: %s", configIdentity,
			)
		}
		uniqueConfig[configIdentity] = true
	}

	if len(m.Contact) > MaxDymNameContactLength {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"invalid contact length; got: %d, max: %d", len(m.Contact), MaxDymNameContactLength,
		)
	}

	return nil
}

// Validate checks if the DymNameConfig record is valid.
func (m *DymNameConfig) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "dym name config is nil")
	}

	if m.ChainId == "" {
		// ok to be empty
	} else if !dymnsutils.IsValidChainIdFormat(m.ChainId) {
		return errorsmod.Wrap(
			gerrc.ErrInvalidArgument,
			"dym name config chain id must be a valid chain id format",
		)
	}

	if m.Path == "" {
		// ok to be empty
	} else if !dymnsutils.IsValidSubDymName(m.Path) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "dym name config path must be a valid dym name")
	}

	if m.Value != strings.ToLower(m.Value) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "dym name config value must be lowercase")
	}

	if m.Type == DymNameConfigType_NAME {
		if !m.IsDelete() {
			if m.ChainId == "" {
				if !dymnsutils.IsValidBech32AccountAddress(m.Value, false) {
					return errorsmod.Wrap(
						gerrc.ErrInvalidArgument,
						"dym name config value must be a valid bech32 account address",
					)
				}
			} else {
				if !dymnsutils.PossibleAccountRegardlessChain(m.Value) {
					return errorsmod.Wrapf(
						gerrc.ErrInvalidArgument,
						"dym name config value: %s", m.Value,
					)
				}
			}
		}
	} else {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"dym name config type must be: %s", DymNameConfigType_NAME.String(),
		)
	}

	return nil
}

// Validate checks if the ReverseLookupDymNames record is valid.
func (m *ReverseLookupDymNames) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "reverse lookup record is nil")
	}

	for _, name := range m.DymNames {
		if !dymnsutils.IsValidDymName(name) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid dym name: %s", name)
		}
	}

	return nil
}

// IsExpiredAtCtx returns true if the Dym-Name is expired at the given context.
// It compares the expiry with the block time in context.
func (m DymName) IsExpiredAtCtx(ctx sdk.Context) bool {
	return m.ExpireAt < ctx.BlockTime().Unix()
}

// IsProhibitedTradingAt checks if the Dym-Name is prohibited from trading at the given anchor time.
// The prohibition is based on the expiry time
// and Dym-Name is prohibited from trading after the given duration before expiry.
func (m DymName) IsProhibitedTradingAt(anchor time.Time, prohibitSellDuration time.Duration) bool {
	prohibitSellingAfterEpoch := time.Unix(m.ExpireAt, 0).UTC().Add(
		-1 * prohibitSellDuration,
	).Unix()

	return prohibitSellingAfterEpoch < anchor.Unix()
}

// GetSdkEvent returns the sdk event contains information of Dym-Name.
// Fired when Dym-Name record is set into store.
func (m DymName) GetSdkEvent() sdk.Event {
	return sdk.NewEvent(
		EventTypeSetDymName,
		sdk.NewAttribute(AttributeKeyDymName, m.Name),
		sdk.NewAttribute(AttributeKeyDymNameOwner, m.Owner),
		sdk.NewAttribute(AttributeKeyDymNameController, m.Controller),
		sdk.NewAttribute(AttributeKeyDymNameExpiryEpoch, fmt.Sprintf("%d", m.ExpireAt)),
		sdk.NewAttribute(AttributeKeyDymNameConfigCount, fmt.Sprintf("%d", len(m.Configs))),
		sdk.NewAttribute(AttributeKeyDymNameHasContactDetails, fmt.Sprintf("%t", m.Contact != "")),
	)
}

// GetIdentity returns the unique identity of the DymNameConfig record.
// Used for uniqueness check.
func (m DymNameConfig) GetIdentity() string {
	return strings.ToLower(fmt.Sprintf("%s|%s|%s", m.Type, m.ChainId, m.Path))
}

// IsDefaultNameConfig checks if the config is a default name config, satisfy the following conditions:
//   - Type is NAME
//   - ChainId is empty (means host chain)
//   - Path is empty (means root Dym-Name)
func (m DymNameConfig) IsDefaultNameConfig() bool {
	return m.Type == DymNameConfigType_NAME &&
		m.ChainId == "" &&
		m.Path == ""
}

// IsDelete checks if the config is a delete operation.
// A delete operation is when the value is empty.
func (m DymNameConfig) IsDelete() bool {
	return m.Value == ""
}

// DymNameConfigs is a list of DymNameConfig records.
// Used to add some operations on the list.
type DymNameConfigs []DymNameConfig

// DefaultNameConfigs returns a list of default name configs.
// It returns a list instead of a single record with purpose is to negotiate case
// where both add and delete operations are present.
func (m DymNameConfigs) DefaultNameConfigs() DymNameConfigs {
	var defaultConfigs DymNameConfigs
	for _, config := range m {
		if config.IsDefaultNameConfig() {
			defaultConfigs = append(defaultConfigs, config)
		}
	}
	return defaultConfigs
}

// GetAddressesForReverseMapping parses the Dym-Name configuration and returns a map of addresses to their configurations.
func (m *DymName) GetAddressesForReverseMapping() (
	configuredAddressesToConfigs map[string][]DymNameConfig,
	fallbackAddressesToConfigs map[string][]DymNameConfig, // TODO DymNS: rename related
	// Describe usage of Go Map: used to mapping each address to its configuration,
	// caller should have responsibility to handle the result and aware of iterating over map can cause non-determinism
) {
	if err := m.Validate(); err != nil {
		// should validate before calling this method
		panic(err)
	}

	configuredAddressesToConfigs = make(map[string][]DymNameConfig)
	fallbackAddressesToConfigs = make(map[string][]DymNameConfig)

	addConfiguredAddress := func(address string, config DymNameConfig) {
		configuredAddressesToConfigs[address] = append(configuredAddressesToConfigs[address], config)
	}

	addFallbackAddress := func(fallbackAddr FallbackAddress, config DymNameConfig) {
		strAddr := fallbackAddr.String()
		fallbackAddressesToConfigs[strAddr] = append(fallbackAddressesToConfigs[strAddr], config)
	}

	var nameConfigs []DymNameConfig
	for _, config := range m.Configs {
		if config.Type == DymNameConfigType_NAME {
			nameConfigs = append(nameConfigs, config)
		}
	}

	var defaultConfig *DymNameConfig
	for i, config := range nameConfigs {
		if config.IsDefaultNameConfig() {
			if config.Value == "" {
				config.Value = m.Owner
				nameConfigs[i] = config
			}

			defaultConfig = &config
			break
		}
	}

	if defaultConfig == nil {
		// add a fake record to be used to generate default address
		nameConfigs = append(nameConfigs, DymNameConfig{
			Type:    DymNameConfigType_NAME,
			ChainId: "",
			Path:    "",
			Value:   m.Owner,
		})
	}

	for _, config := range nameConfigs {
		if config.Value == "" {
			continue
		}

		// TODO DymNS: remove this comment because we are going to widely support other formats
		// just a friendly reminder, in the current implementation,
		// config value is always a bech32 account address

		if config.IsDefaultNameConfig() {
			// default config is for host chain only so value must be valid bech32
			accAddr, err := sdk.AccAddressFromBech32(config.Value)
			if err != nil {
				// should not happen as configuration should be validated before calling this method
				panic(err)
			}

			addConfiguredAddress(config.Value, config)
			addFallbackAddress(FallbackAddress(accAddr), config)

			continue
		}

		addConfiguredAddress(config.Value, config)

		// note: this config is not a default config, is not a part of fallback mechanism,
		// so we don't add fallback address for this config
	}

	return
}

// Distinct returns a new list of dym names with duplicates removed.
// Result will be sorted.
func (m ReverseLookupDymNames) Distinct() (distinct ReverseLookupDymNames) {
	return ReverseLookupDymNames{
		DymNames: StringList(m.DymNames).Distinct(),
	}
}

// Combine merges the dym names from the current list and the other list.
// Result will be sorted distinct.
func (m ReverseLookupDymNames) Combine(other ReverseLookupDymNames) ReverseLookupDymNames {
	return ReverseLookupDymNames{
		DymNames: StringList(m.DymNames).Combine(other.DymNames),
	}.Distinct()
}

// Exclude removes the dym names from the current list that are in the toBeExcluded list.
// Result will be sorted distinct.
func (m ReverseLookupDymNames) Exclude(toBeExcluded ReverseLookupDymNames) (afterExcluded ReverseLookupDymNames) {
	return ReverseLookupDymNames{
		DymNames: StringList(m.DymNames).Exclude(toBeExcluded.DymNames),
	}.Distinct()
}

// Sort sorts the dym names in the list.
func (m ReverseLookupDymNames) Sort() ReverseLookupDymNames {
	sort.Strings(m.DymNames)
	return m
}
