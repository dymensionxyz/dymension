package keeper

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// AddReverseMappingOwnerToOwnedDymName add a reverse mapping from owner to owned Dym-Name into the KVStore.
func (k Keeper) AddReverseMappingOwnerToOwnedDymName(ctx sdk.Context, owner, name string) error {
	_, bzAccAddr, err := bech32.DecodeAndConvert(owner)
	if err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, owner)
	}

	dymNamesOwnedByAccountKey := dymnstypes.DymNamesOwnedByAccountRvlKey(bzAccAddr)

	return k.GenericAddReverseLookupDymNamesRecord(ctx, dymNamesOwnedByAccountKey, name)
}

// GetDymNamesOwnedBy returns all Dym-Names owned by the account address.
// The action done by reverse mapping from owner to owned Dym-Name.
// The Dym-Names are filtered by the owner and excluded expired Dym-Name using the time from context.
func (k Keeper) GetDymNamesOwnedBy(
	ctx sdk.Context, owner string,
) ([]dymnstypes.DymName, error) {
	_, bzAccAddr, err := bech32.DecodeAndConvert(owner)
	if err != nil {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, owner)
	}

	dymNamesOwnedByAccountKey := dymnstypes.DymNamesOwnedByAccountRvlKey(bzAccAddr)

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
	return strings.ToLower(strings.TrimSpace(configuredAddress))
}

// AddReverseMappingHexAddressToDymName add a reverse mapping
// from hex address (coin-type 60, secp256k1, ethereum address)
// to Dym-Name which contains the hex address, into the KVStore.
func (k Keeper) AddReverseMappingHexAddressToDymName(ctx sdk.Context, bzHexAddr []byte, name string) error {
	if err := validateHexAddressForReverseMapping(bzHexAddr); err != nil {
		return err
	}

	return k.GenericAddReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.HexAddressToDymNamesIncludeRvlKey(bzHexAddr),
		name,
	)
}

// GetDymNamesContainsHexAddress returns all Dym-Names
// that contains the hex address (coin-type 60, secp256k1, ethereum address).
// The action done by reverse mapping from hex address to Dym-Name.
// The Dym-Names are filtered by the hex address and excluded expired Dym-Name using the time from context.
func (k Keeper) GetDymNamesContainsHexAddress(
	ctx sdk.Context, bzHexAddr []byte,
) ([]dymnstypes.DymName, error) {
	if err := validateHexAddressForReverseMapping(bzHexAddr); err != nil {
		return nil, err
	}

	key := dymnstypes.HexAddressToDymNamesIncludeRvlKey(bzHexAddr)

	currentDymNamesContainsHexAddress := k.GenericGetReverseLookupDymNamesRecord(ctx, key)

	var dymNames []dymnstypes.DymName
	for _, name := range currentDymNamesContainsHexAddress.DymNames {
		dymName := k.GetDymNameWithExpirationCheck(ctx, name)
		if dymName == nil {
			// dym-name not found, skip
			continue
		}
		dymNames = append(dymNames, *dymName)
	}

	return dymNames, nil
}

// RemoveReverseMappingHexAddressToDymName removes reverse mapping
// from hex address (coin-type 60, secp256k1, ethereum address)
// to Dym-Names which contains it from the KVStore.
func (k Keeper) RemoveReverseMappingHexAddressToDymName(ctx sdk.Context, bzHexAddr []byte, name string) error {
	if err := validateHexAddressForReverseMapping(bzHexAddr); err != nil {
		return err
	}

	return k.GenericRemoveReverseLookupDymNamesRecord(
		ctx,
		dymnstypes.HexAddressToDymNamesIncludeRvlKey(bzHexAddr),
		name,
	)
}

// validateHexAddressForReverseMapping validates the hex address for reverse mapping.
func validateHexAddressForReverseMapping(bzHexAddr []byte) error {
	if length := len(bzHexAddr); length != 20 && length != 32 {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "hex address must be 20 or 32 bytes, got: %d", length)
	}
	return nil
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
