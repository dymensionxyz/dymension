package keeper

import (
	"strings"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// AddReverseMappingOwnerToOwnedDymName add a reverse mapping from owner to owned Dym-Name into the KVStore.
func (k Keeper) AddReverseMappingOwnerToOwnedDymName(ctx sdk.Context, owner, name string) error {
	accAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, owner)
	}

	dymNamesOwnedByAccountKey := dymnstypes.DymNamesOwnedByAccountRvlKey(accAddr)

	return k.GenericAddReverseLookupDymNamesRecord(ctx, dymNamesOwnedByAccountKey, name)
}

// GetDymNamesOwnedBy returns all Dym-Names owned by the account address.
// The action done by reverse mapping from owner to owned Dym-Name.
// The Dym-Names are filtered by the owner and excluded expired Dym-Name using the time from context.
func (k Keeper) GetDymNamesOwnedBy(
	ctx sdk.Context, owner string,
) ([]dymnstypes.DymName, error) {
	accAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, owner)
	}

	dymNamesOwnedByAccountKey := dymnstypes.DymNamesOwnedByAccountRvlKey(accAddr)

	existingOwnedDymNames := k.GenericGetReverseLookupDymNamesRecord(ctx, dymNamesOwnedByAccountKey)

	var dymNames []dymnstypes.DymName
	for _, owned := range existingOwnedDymNames.DymNames {
		dymName := k.GetDymNameWithExpirationCheck(ctx, owned)
		if dymName == nil {
			// dym-name not found or expired, skip
			continue
		}
		if dymName.Owner != owner {
			// dym-name owner mismatch, skip
			continue
		}
		dymNames = append(dymNames, *dymName)
	}

	return dymNames, nil
}

// RemoveReverseMappingOwnerToOwnedDymName removes a reverse mapping from owner to owned Dym-Name from the KVStore.
func (k Keeper) RemoveReverseMappingOwnerToOwnedDymName(ctx sdk.Context, owner, name string) error {
	accAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address: %s", owner)
	}

	dymNamesOwnedByAccountKey := dymnstypes.DymNamesOwnedByAccountRvlKey(accAddr)

	return k.GenericRemoveReverseLookupDymNamesRecord(ctx, dymNamesOwnedByAccountKey, name)
}

// AddReverseMappingConfiguredAddressToDymName add a reverse mapping from configured address to Dym-Name
// which contains the configuration, into the KVStore.
func (k Keeper) AddReverseMappingConfiguredAddressToDymName(ctx sdk.Context, configuredAddress, name string) error {
	configuredAddress = normalizeConfiguredAddressForReverseMapping(configuredAddress)
	if err := validateConfiguredAddressForReverseMapping(configuredAddress); err != nil {
		return err
	}

	return k.GenericAddReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(configuredAddress),
		name,
	)
}

// GetDymNamesContainsConfiguredAddress returns all Dym-Names that contains the configured address.
// The action done by reverse mapping from configured address to Dym-Name.
// The Dym-Names are filtered by the configured address and excluded expired Dym-Name using the time from context.
func (k Keeper) GetDymNamesContainsConfiguredAddress(
	ctx sdk.Context, configuredAddress string,
) ([]dymnstypes.DymName, error) {
	configuredAddress = normalizeConfiguredAddressForReverseMapping(configuredAddress)
	if err := validateConfiguredAddressForReverseMapping(configuredAddress); err != nil {
		return nil, err
	}

	key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(configuredAddress)

	currentDymNamesContainsConfiguredAddress := k.GenericGetReverseLookupDymNamesRecord(ctx, key)

	var dymNames []dymnstypes.DymName
	for _, name := range currentDymNamesContainsConfiguredAddress.DymNames {
		dymName := k.GetDymNameWithExpirationCheck(ctx, name)
		if dymName == nil {
			// dym-name not found, skip
			continue
		}
		dymNames = append(dymNames, *dymName)
	}

	return dymNames, nil
}

// RemoveReverseMappingConfiguredAddressToDymName removes reverse mapping from configured address
// to Dym-Names which contains it from the KVStore.
func (k Keeper) RemoveReverseMappingConfiguredAddressToDymName(ctx sdk.Context, configuredAddress, name string) error {
	configuredAddress = normalizeConfiguredAddressForReverseMapping(configuredAddress)
	if err := validateConfiguredAddressForReverseMapping(configuredAddress); err != nil {
		return err
	}

	return k.GenericRemoveReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(configuredAddress),
		name,
	)
}

// validateConfiguredAddressForReverseMapping validates the configured address for reverse mapping.
func validateConfiguredAddressForReverseMapping(configuredAddress string) error {
	if configuredAddress == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "configured address cannot be blank")
	}
	return nil
}

// normalizeConfiguredAddressForReverseMapping normalizes the configured address for reverse mapping
// before putting it into the KVStore.
func normalizeConfiguredAddressForReverseMapping(configuredAddress string) string {
	configuredAddress = strings.TrimSpace(configuredAddress)

	if dymnsutils.IsValidHexAddress(configuredAddress) {
		// if the address is hex format, then treat the chain is case-insensitive address,
		// like Ethereum, where the address is case-insensitive and checksum address contains mixed case
		configuredAddress = strings.ToLower(configuredAddress)
	}

	return configuredAddress
}

// AddReverseMappingFallbackAddressToDymName add a reverse mapping
// from fallback address to Dym-Name which contains the fallback address, into the KVStore.
func (k Keeper) AddReverseMappingFallbackAddressToDymName(ctx sdk.Context, fallbackAddr dymnstypes.FallbackAddress, name string) error {
	if err := fallbackAddr.ValidateBasic(); err != nil {
		return err
	}

	return k.GenericAddReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(fallbackAddr),
		name,
	)
}

// GetDymNamesContainsFallbackAddress returns all Dym-Names
// that contains the fallback address.
// The action done by reverse mapping from fallback address to Dym-Name.
// The Dym-Names are filtered by the fallback address and excluded expired Dym-Name using the time from context.
func (k Keeper) GetDymNamesContainsFallbackAddress(
	ctx sdk.Context, fallbackAddr dymnstypes.FallbackAddress,
) ([]dymnstypes.DymName, error) {
	if err := fallbackAddr.ValidateBasic(); err != nil {
		return nil, err
	}

	key := dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(fallbackAddr)

	currentDymNamesContainsFallbackAddress := k.GenericGetReverseLookupDymNamesRecord(ctx, key)

	var dymNames []dymnstypes.DymName
	for _, name := range currentDymNamesContainsFallbackAddress.DymNames {
		dymName := k.GetDymNameWithExpirationCheck(ctx, name)
		if dymName == nil {
			// dym-name not found, skip
			continue
		}
		dymNames = append(dymNames, *dymName)
	}

	return dymNames, nil
}

// RemoveReverseMappingFallbackAddressToDymName removes reverse mapping
// from fallback address to Dym-Names which contains it from the KVStore.
func (k Keeper) RemoveReverseMappingFallbackAddressToDymName(ctx sdk.Context, fallbackAddr dymnstypes.FallbackAddress, name string) error {
	if err := fallbackAddr.ValidateBasic(); err != nil {
		return err
	}

	return k.GenericRemoveReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(fallbackAddr),
		name,
	)
}

// GenericAddReverseLookupDymNamesRecord is a utility method that help to add a reverse lookup record for Dym-Names.
func (k Keeper) GenericAddReverseLookupDymNamesRecord(ctx sdk.Context, key []byte, name string) error {
	return k.GenericAddReverseLookupRecord(
		ctx,
		key, name,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupDymNames{
				DymNames: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupDymNames
			k.cdc.MustUnmarshal(bz, &record)
			return record.DymNames
		},
	)
}

// GenericGetReverseLookupDymNamesRecord is a utility method that help to get a reverse lookup record for Dym-Names.
func (k Keeper) GenericGetReverseLookupDymNamesRecord(
	ctx sdk.Context, key []byte,
) dymnstypes.ReverseLookupDymNames {
	dymNames := k.GenericGetReverseLookupRecord(
		ctx,
		key,
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupDymNames
			k.cdc.MustUnmarshal(bz, &record)
			return record.DymNames
		},
	)

	return dymnstypes.ReverseLookupDymNames{
		DymNames: dymNames,
	}
}

// GenericRemoveReverseLookupDymNamesRecord is a utility method that help to remove a reverse lookup record for Dym-Names.
func (k Keeper) GenericRemoveReverseLookupDymNamesRecord(ctx sdk.Context, key []byte, name string) error {
	return k.GenericRemoveReverseLookupRecord(
		ctx,
		key, name,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupDymNames{
				DymNames: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupDymNames
			k.cdc.MustUnmarshal(bz, &record)
			return record.DymNames
		},
	)
}
