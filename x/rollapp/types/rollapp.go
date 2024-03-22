package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Creator)
	if err != nil {
		return errorsmod.Wrap(err, ErrInvalidCreatorAddress.Error())
	}
	if r.GetMaxSequencers() == 0 {
		return ErrInvalidMaxSequencers
	}

	// validate rollappId
	_, _, err = GetValidEIP155ChainId(r.RollappId)
	if err != nil {
		return err
	}

	// verifies that there's no duplicate address in PermissionedAddresses
	// and addresses are in Bech32 format
	permissionedAddresses := r.GetPermissionedAddresses()
	if len(permissionedAddresses) > 0 {
		duplicateAddresses := make(map[string]bool)
		for _, item := range permissionedAddresses {
			// check if the item/element exist in the duplicateAddresses map
			_, exist := duplicateAddresses[item]
			if exist {
				return errorsmod.Wrapf(ErrPermissionedAddressesDuplicate, "address: %s", item)
			}
			// check Bech32 format
			if _, err := sdk.AccAddressFromBech32(item); err != nil {
				return errorsmod.Wrapf(ErrInvalidPermissionedAddress, "invalid permissioned address: %s", err)
			}
			// mark as exist
			duplicateAddresses[item] = true
		}
	}

	// verifies that token metadata, if any, must be valid
	if len(r.TokenMetadata) > 0 {
		for _, metadata := range r.TokenMetadata {
			if err := metadata.Validate(); err != nil {
				return errorsmod.Wrapf(ErrInvalidTokenMetadata, "%s: %v", metadata.Base, err)
			}
		}
	}

	// genesisAccounts address validation
	if r.GenesisState != nil {
		for _, acc := range r.GenesisState.GenesisAccounts {
			_, err := sdk.AccAddressFromBech32(acc.Address)
			if err != nil {
				return errorsmod.Wrapf(err, "invalid genesis account address (%s)", acc.Address)
			}
		}
	}

	return nil
}
